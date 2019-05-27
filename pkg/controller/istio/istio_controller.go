/*
Copyright 2019 Banzai Cloud.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package istio

import (
	"context"
	"flag"
	"net"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/gofrs/uuid"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	remoteistioCtrl "github.com/banzaicloud/istio-operator/pkg/controller/remoteistio"
	"github.com/banzaicloud/istio-operator/pkg/crds"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources"
	"github.com/banzaicloud/istio-operator/pkg/resources/citadel"
	"github.com/banzaicloud/istio-operator/pkg/resources/cni"
	"github.com/banzaicloud/istio-operator/pkg/resources/common"
	"github.com/banzaicloud/istio-operator/pkg/resources/galley"
	"github.com/banzaicloud/istio-operator/pkg/resources/gateways"
	"github.com/banzaicloud/istio-operator/pkg/resources/istiocoredns"
	"github.com/banzaicloud/istio-operator/pkg/resources/mixer"
	"github.com/banzaicloud/istio-operator/pkg/resources/nodeagent"
	"github.com/banzaicloud/istio-operator/pkg/resources/pilot"
	"github.com/banzaicloud/istio-operator/pkg/resources/sidecarinjector"
	"github.com/banzaicloud/istio-operator/pkg/util"
	objectmatch "github.com/banzaicloud/k8s-objectmatcher"
)

const finalizerID = "istio-operator.finializer.banzaicloud.io"
const istioSecretTypePrefix = "istio.io"
const localNetworkName = "local-network"

var log = logf.Log.WithName("controller")
var watchCreatedResourcesEvents bool

func init() {
	flag.BoolVar(&watchCreatedResourcesEvents, "watch-created-resources-events", true, "Whether to watch created resources events")
}

// Add creates a new Config Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	dynamic, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		return emperror.Wrap(err, "failed to create dynamic client")
	}
	crd, err := crds.New(mgr.GetConfig(), crds.InitCrds())
	if err != nil {
		return emperror.Wrap(err, "unable to set up crd operator")
	}
	return add(mgr, newReconciler(mgr, dynamic, crd))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, d dynamic.Interface, crd *crds.CrdOperator) reconcile.Reconciler {
	return &ReconcileConfig{
		Client:      mgr.GetClient(),
		dynamic:     d,
		crdOperator: crd,
		mgr:         mgr,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("config-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	err = initWatches(c, mgr.GetScheme(), watchCreatedResourcesEvents, log)
	if err != nil {
		return err
	}
	return nil
}

var _ reconcile.Reconciler = &ReconcileConfig{}

// ReconcileConfig reconciles a Config object
type ReconcileConfig struct {
	client.Client
	dynamic     dynamic.Interface
	crdOperator *crds.CrdOperator
	mgr         manager.Manager
}

type ReconcileComponent func(log logr.Logger, istio *istiov1beta1.Istio) error

// +kubebuilder:rbac:groups="",resources=nodes;services;endpoints;pods;replicationcontrollers;services;endpoints;pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=serviceaccounts;configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups="apps",resources=replicasets,verbs=get;list;watch
// +kubebuilder:rbac:groups="apps",resources=deployments;daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apps",resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="extensions",resources=ingresses;ingresses/status,verbs=*
// +kubebuilder:rbac:groups="extensions",resources=deployments,verbs=get
// +kubebuilder:rbac:groups="extensions",resources=deployments/finalizers,verbs=update
// +kubebuilder:rbac:groups="extensions",resources=replicasets,verbs=get;list;watch
// +kubebuilder:rbac:groups="policy",resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="autoscaling",resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=*
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterroles;clusterrolebindings;roles;rolebindings;,verbs=get;list;watch;create;update;patch;delete

// +kubebuilder:rbac:groups=istio.banzaicloud.io,resources=istios,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=istio.banzaicloud.io,resources=istios/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=authentication.istio.io;cloud.istio.io;config.istio.io;istio.istio.io;networking.istio.io;rbac.istio.io;scalingpolicy.istio.io,resources=*,verbs=*

// Reconcile reads that state of the cluster for a Config object and makes changes based on the state read
// and what is in the Config.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
func (r *ReconcileConfig) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger := log.WithValues("trigger", request.Namespace+"/"+request.Name, "correlationID", uuid.Must(uuid.NewV4()).String())
	// Fetch the Config instance
	config := &istiov1beta1.Istio{}
	err := r.Get(context.TODO(), request.NamespacedName, config)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	logger.Info("Reconciling Istio")

	if !config.Spec.Version.IsSupported() {
		err = errors.New("intended Istio version is unsupported by this version of the operator")
		logger.Error(err, "", "version", config.Spec.Version)
		return reconcile.Result{
			Requeue: false,
		}, nil
	}

	// Set default values where not set
	istiov1beta1.SetDefaults(config)
	result, err := r.reconcile(logger, config)
	if err != nil {
		updateErr := updateStatus(r.Client, config, istiov1beta1.ReconcileFailed, err.Error(), logger)
		if updateErr != nil {
			logger.Error(updateErr, "failed to update state")
			return result, errors.WithStack(err)
		}
		return result, emperror.Wrap(err, "could not reconcile istio")
	}
	return result, nil
}

func (r *ReconcileConfig) reconcile(logger logr.Logger, config *istiov1beta1.Istio) (reconcile.Result, error) {

	if config.Status.Status == "" {
		err := updateStatus(r.Client, config, istiov1beta1.Created, "", logger)
		if err != nil {
			return reconcile.Result{}, errors.WithStack(err)
		}
	}

	// add finalizer strings and update
	if config.ObjectMeta.DeletionTimestamp.IsZero() {
		if !util.ContainsString(config.ObjectMeta.Finalizers, finalizerID) {
			config.ObjectMeta.Finalizers = append(config.ObjectMeta.Finalizers, finalizerID)
			if err := r.Update(context.Background(), config); err != nil {
				return reconcile.Result{}, emperror.Wrap(err, "could not add finalizer to config")
			}
			return reconcile.Result{
				Requeue: true,
			}, nil
		}
	} else {
		// Deletion timestamp set, config is marked for deletion
		if util.ContainsString(config.ObjectMeta.Finalizers, finalizerID) {
			if config.Status.Status == istiov1beta1.Reconciling && config.Status.ErrorMessage == "" {
				logger.Info("cannot remove Istio while reconciling")
				return reconcile.Result{}, nil
			}
			// Remove remote istio resources
			r.deleteRemoteIstios(config, logger)
			// Set citadel deployment as owner reference to istio secrets for garbage cleanup
			r.setCitadelAsOwnerReferenceToIstioSecrets(config, appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      citadel.GetDeploymentName(),
					Namespace: config.Namespace,
				},
			}, logger)
			config.ObjectMeta.Finalizers = util.RemoveString(config.ObjectMeta.Finalizers, finalizerID)
			if err := r.Update(context.Background(), config); err != nil {
				return reconcile.Result{}, emperror.Wrap(err, "could not remove finalizer from config")
			}
		}

		logger.Info("Istio removed")

		return reconcile.Result{}, nil
	}

	if config.Status.Status == istiov1beta1.Reconciling {
		logger.Info("cannot trigger reconcile while already reconciling")
		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: time.Duration(30) * time.Second,
		}, nil
	}

	err := updateStatus(r.Client, config, istiov1beta1.Reconciling, "", logger)
	if err != nil {
		return reconcile.Result{}, errors.WithStack(err)
	}

	logger.Info("reconciling CRDs")
	err = r.crdOperator.Reconcile(config, logger)
	if err != nil {
		logger.Error(err, "unable to reconcile CRDs")
		return reconcile.Result{}, err
	}

	if util.PointerToBool(config.Spec.MeshExpansion) {
		meshNetworks, err := r.getMeshNetworks(config, logger)
		if err != nil {
			return reconcile.Result{}, err
		}
		config.Spec.SetMeshNetworks(meshNetworks)
	}

	reconcilers := []resources.ComponentReconciler{
		common.New(r.Client, config, false),
		citadel.New(citadel.Configuration{
			DeployMeshPolicy: true,
		}, r.Client, r.dynamic, config),
		galley.New(r.Client, config),
		pilot.New(r.Client, r.dynamic, config),
		gateways.New(r.Client, r.dynamic, config),
		mixer.New(r.Client, r.dynamic, config),
		cni.New(r.Client, config),
		sidecarinjector.New(r.Client, config),
		nodeagent.New(r.Client, config),
		istiocoredns.New(r.Client, config),
	}

	for _, rec := range reconcilers {
		err = rec.Reconcile(logger)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	ingressGatewayAddress, err := r.getIngressGatewayAddress(config, logger)
	if err != nil {
		log.Info(err.Error())
		updateStatus(r.Client, config, istiov1beta1.ReconcileFailed, err.Error(), logger)
		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: time.Duration(30) * time.Second,
		}, nil
	}

	config.Status.GatewayAddress = ingressGatewayAddress

	err = updateStatus(r.Client, config, istiov1beta1.Available, "", logger)
	if err != nil {
		return reconcile.Result{}, errors.WithStack(err)
	}
	logger.Info("reconcile finished")
	return reconcile.Result{}, nil
}

func (r *ReconcileConfig) getIngressGatewayAddress(istio *istiov1beta1.Istio, logger logr.Logger) ([]string, error) {
	var service corev1.Service

	ips := make([]string, 0)

	err := r.Get(context.TODO(), client.ObjectKey{
		Name:      "istio-ingressgateway",
		Namespace: istio.Namespace,
	}, &service)
	if err != nil && !k8serrors.IsNotFound(err) {
		return ips, err
	}

	if len(service.Status.LoadBalancer.Ingress) < 1 {
		return ips, errors.New("invalid ingress status")
	}

	if service.Status.LoadBalancer.Ingress[0].IP != "" {
		ips = []string{
			service.Status.LoadBalancer.Ingress[0].IP,
		}
	} else if service.Status.LoadBalancer.Ingress[0].Hostname != "" {
		hostIPs, err := net.LookupIP(service.Status.LoadBalancer.Ingress[0].Hostname)
		if err != nil {
			return ips, err
		}
		for _, ip := range hostIPs {
			if ip.To4() != nil {
				ips = append(ips, ip.String())
			}
		}
	}

	return ips, nil
}

func (r *ReconcileConfig) getMeshNetworks(config *istiov1beta1.Istio, logger logr.Logger) (*istiov1beta1.MeshNetworks, error) {
	meshNetworks := make(map[string]istiov1beta1.MeshNetwork)

	if len(config.Status.GatewayAddress) > 0 {
		gateways := make([]istiov1beta1.MeshNetworkGateway, 0)
		for _, address := range config.Status.GatewayAddress {
			gateways = append(gateways, istiov1beta1.MeshNetworkGateway{
				Address: address, Port: 443,
			})
		}
		meshNetworks[localNetworkName] = istiov1beta1.MeshNetwork{
			Endpoints: []istiov1beta1.MeshNetworkEndpoint{
				{
					FromCIDR: "127.0.0.1/8",
				},
			},
			Gateways: gateways,
		}
	}

	remoteIstios := remoteistioCtrl.GetRemoteIstiosByOwnerReference(r.mgr, config, logger)
	for _, remoteIstio := range remoteIstios {
		gateways := make([]istiov1beta1.MeshNetworkGateway, 0)
		if len(remoteIstio.Status.GatewayAddress) > 0 {
			for _, address := range remoteIstio.Status.GatewayAddress {
				gateways = append(gateways, istiov1beta1.MeshNetworkGateway{
					Address: address, Port: 443,
				})
			}
		} else {
			continue
		}

		meshNetworks[remoteIstio.Name] = istiov1beta1.MeshNetwork{
			Endpoints: []istiov1beta1.MeshNetworkEndpoint{
				{
					FromRegistry: remoteIstio.Name,
				},
			},
			Gateways: gateways,
		}
	}

	return &istiov1beta1.MeshNetworks{
		Networks: meshNetworks,
	}, nil
}

func (r *ReconcileConfig) setCitadelAsOwnerReferenceToIstioSecrets(config *istiov1beta1.Istio, deployment appsv1.Deployment, logger logr.Logger) error {
	var secrets corev1.SecretList

	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      deployment.Name,
		Namespace: deployment.Namespace,
	}, &deployment)
	if k8serrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	err = r.Client.List(context.TODO(), &client.ListOptions{
		Namespace: config.Namespace,
	}, &secrets)
	if err != nil {
		return err
	}

	for _, secret := range secrets.Items {
		if !strings.HasPrefix(string(secret.Type), istioSecretTypePrefix) {
			continue
		}

		logger.V(0).Info("setting owner reference to secret", "secret", secret.Name, "refs", secret.GetOwnerReferences())

		refs, err := k8sutil.SetOwnerReferenceToObject(&secret, &deployment)
		if err != nil {
			return err
		}

		secret.SetOwnerReferences(refs)
		err = r.Update(context.TODO(), &secret)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ReconcileConfig) deleteRemoteIstios(config *istiov1beta1.Istio, logger logr.Logger) {
	remoteIstios := remoteistioCtrl.GetRemoteIstiosByOwnerReference(r.mgr, config, logger)
	for _, remoteIstio := range remoteIstios {
		err := r.Client.Delete(context.Background(), &remoteIstio)
		if err != nil {
			logger.Error(err, "could not delete remote istio resource", "name", remoteIstio.Name)
		}
	}
}

func updateStatus(c client.Client, config *istiov1beta1.Istio, status istiov1beta1.ConfigState, errorMessage string, logger logr.Logger) error {
	typeMeta := config.TypeMeta
	config.Status.Status = status
	config.Status.ErrorMessage = errorMessage
	err := c.Status().Update(context.Background(), config)
	if k8serrors.IsNotFound(err) {
		err = c.Update(context.Background(), config)
	}
	if err != nil {
		if !k8serrors.IsConflict(err) {
			return emperror.Wrapf(err, "could not update Istio state to '%s'", status)
		}
		err := c.Get(context.TODO(), types.NamespacedName{
			Namespace: config.Namespace,
			Name:      config.Name,
		}, config)
		if err != nil {
			return emperror.Wrap(err, "could not get config for updating status")
		}
		config.Status.Status = status
		config.Status.ErrorMessage = errorMessage
		err = c.Status().Update(context.Background(), config)
		if k8serrors.IsNotFound(err) {
			err = c.Update(context.Background(), config)
		}
		if err != nil {
			return emperror.Wrapf(err, "could not update Istio state to '%s'", status)
		}
	}
	// update loses the typeMeta of the config that's used later when setting ownerrefs
	config.TypeMeta = typeMeta
	logger.Info("Istio state updated", "status", status)
	return nil
}

func initWatches(c controller.Controller, scheme *runtime.Scheme, watchCreatedResourcesEvents bool, logger logr.Logger) error {
	// Watch for changes to Config
	err := c.Watch(&source.Kind{Type: &istiov1beta1.Istio{TypeMeta: metav1.TypeMeta{Kind: "Istio", APIVersion: "istio.banzaicloud.io/v1beta1"}}}, &handler.EnqueueRequestForObject{}, k8sutil.GetWatchPredicateForIstio())
	if err != nil {
		return err
	}

	// Watch for RemoteIstio changes to trigger reconciliation
	err = c.Watch(&source.Kind{Type: &istiov1beta1.RemoteIstio{TypeMeta: metav1.TypeMeta{Kind: "RemoteIstio", APIVersion: "istio.banzaicloud.io/v1beta1"}}}, &handler.EnqueueRequestsFromMapFunc{
		ToRequests: handler.ToRequestsFunc(func(object handler.MapObject) []reconcile.Request {
			own := object.Meta.GetOwnerReferences()
			if len(own) < 1 {
				return nil
			}
			return []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Name:      own[0].Name,
						Namespace: object.Meta.GetNamespace(),
					},
				},
			}
		}),
	})
	if err != nil {
		return err
	}

	// Watch for changes to Istio coreDNS service
	err = c.Watch(&source.Kind{Type: &corev1.Service{TypeMeta: metav1.TypeMeta{Kind: "Service", APIVersion: "v1"}}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &istiov1beta1.Istio{},
	}, k8sutil.GetWatchPredicateForIstioService("istiocoredns"))
	if err != nil {
		return err
	}

	if !watchCreatedResourcesEvents {
		return nil
	}

	// Initialize object matcher
	objectMatcher := objectmatch.New(logf.NewDelegatingLogger(logf.NullLogger{}))

	// Initialize owner matcher
	ownerMatcher := k8sutil.NewOwnerReferenceMatcher(&istiov1beta1.Istio{TypeMeta: metav1.TypeMeta{Kind: "Istio", APIVersion: "istio.banzaicloud.io/v1beta1"}}, true, scheme)

	// Watch for changes to resources managed by the operator
	for _, t := range []runtime.Object{
		&corev1.ServiceAccount{TypeMeta: metav1.TypeMeta{Kind: "ServiceAccount", APIVersion: "v1"}},
		&rbacv1.Role{TypeMeta: metav1.TypeMeta{Kind: "ClusterRole", APIVersion: "v1"}},
		&rbacv1.RoleBinding{TypeMeta: metav1.TypeMeta{Kind: "ClusterRoleBinding", APIVersion: "v1"}},
		&rbacv1.ClusterRole{TypeMeta: metav1.TypeMeta{Kind: "ClusterRole", APIVersion: "v1"}},
		&rbacv1.ClusterRoleBinding{TypeMeta: metav1.TypeMeta{Kind: "ClusterRoleBinding", APIVersion: "v1"}},
		&corev1.ConfigMap{TypeMeta: metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"}},
		&corev1.Service{TypeMeta: metav1.TypeMeta{Kind: "Service", APIVersion: "v1"}},
		&appsv1.Deployment{TypeMeta: metav1.TypeMeta{Kind: "Deployment", APIVersion: "v1"}},
		&appsv1.DaemonSet{TypeMeta: metav1.TypeMeta{Kind: "DaemonSet", APIVersion: "v1"}},
		&autoscalingv2beta1.HorizontalPodAutoscaler{TypeMeta: metav1.TypeMeta{Kind: "HorizontalPodAutoscaler", APIVersion: "v2beta1"}},
		&admissionregistrationv1beta1.MutatingWebhookConfiguration{TypeMeta: metav1.TypeMeta{Kind: "MutatingWebhookConfiguration", APIVersion: "v1beta1"}},
	} {
		err = c.Watch(&source.Kind{Type: t}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &istiov1beta1.Istio{},
		}, predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				return false
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				related, object, err := ownerMatcher.Match(e.Object)
				if err != nil {
					logger.Error(err, "could not determine relation", "kind", e.Object.GetObjectKind())
				}
				if related {
					logger.Info("related object deleted", "trigger", object.GetName())
				}
				return true
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				objectsEquals, err := objectMatcher.Match(e.ObjectOld, e.ObjectNew)
				if err != nil {
					logger.Error(err, "could not match objects", "kind", e.ObjectOld.GetObjectKind())
				} else if objectsEquals {
					return false
				}
				related, object, err := ownerMatcher.Match(e.ObjectNew)
				if err != nil {
					logger.Error(err, "could not determine relation", "kind", e.ObjectNew.GetObjectKind())
				}
				if related {
					logger.Info("related object changed", "trigger", object.GetName())
				}
				return true
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func RemoveFinalizers(c client.Client) error {
	var istios istiov1beta1.IstioList

	err := c.List(context.TODO(), &client.ListOptions{}, &istios)
	if err != nil {
		return emperror.Wrap(err, "could not list Istio resources")
	}
	for _, istio := range istios.Items {
		istio.ObjectMeta.Finalizers = util.RemoveString(istio.ObjectMeta.Finalizers, finalizerID)
		if err := c.Update(context.Background(), &istio); err != nil {
			return emperror.WrapWith(err, "could not remove finalizer from Istio resource", "name", istio.GetName())
		}
		if err := updateStatus(c, &istio, istiov1beta1.Unmanaged, "", log); err != nil {
			return emperror.Wrap(err, "could not update status of Istio resource")
		}
	}

	return nil
}

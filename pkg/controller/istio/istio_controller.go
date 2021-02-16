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
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/gofrs/uuid"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/config"
	remoteistioCtrl "github.com/banzaicloud/istio-operator/pkg/controller/remoteistio"
	"github.com/banzaicloud/istio-operator/pkg/crds"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	k8sutil_mgw "github.com/banzaicloud/istio-operator/pkg/k8sutil/mgw"
	"github.com/banzaicloud/istio-operator/pkg/resources"
	"github.com/banzaicloud/istio-operator/pkg/resources/base"
	"github.com/banzaicloud/istio-operator/pkg/resources/citadel"
	"github.com/banzaicloud/istio-operator/pkg/resources/cni"
	"github.com/banzaicloud/istio-operator/pkg/resources/egressgateway"
	"github.com/banzaicloud/istio-operator/pkg/resources/galley"
	"github.com/banzaicloud/istio-operator/pkg/resources/ingressgateway"
	"github.com/banzaicloud/istio-operator/pkg/resources/istiocoredns"
	"github.com/banzaicloud/istio-operator/pkg/resources/istiod"
	"github.com/banzaicloud/istio-operator/pkg/resources/meshexpansion"
	"github.com/banzaicloud/istio-operator/pkg/resources/mixer"
	"github.com/banzaicloud/istio-operator/pkg/resources/mixerlesstelemetry"
	"github.com/banzaicloud/istio-operator/pkg/resources/nodeagent"
	"github.com/banzaicloud/istio-operator/pkg/resources/pilot"
	"github.com/banzaicloud/istio-operator/pkg/resources/sidecarinjector"
	"github.com/banzaicloud/istio-operator/pkg/resources/webhookcert"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

const finalizerID = "istio-operator.finializer.banzaicloud.io"
const istioSecretTypePrefix = "istio.io"

var log = logf.Log.WithName("controller")
var watchCreatedResourcesEvents bool

type IstioReconciler interface {
	reconcile.Reconciler
	initWatches(watchCreatedResourcesEvents bool) error
	setController(ctrl controller.Controller)
}

func init() {
	flag.BoolVar(&watchCreatedResourcesEvents, "watch-created-resources-events", true, "Whether to watch created resources events")
}

// Add creates a new Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, operatorConfig config.Configuration) error {
	dynamic, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		return emperror.Wrap(err, "failed to create dynamic client")
	}
	crd, err := crds.New(mgr, istiov1beta1.SupportedIstioVersion)
	if err != nil {
		return emperror.Wrap(err, "unable to set up CRD reconciler")
	}
	err = crd.LoadCRDs()
	if err != nil {
		return emperror.Wrap(err, "unable to load CRDs from manifests")
	}
	r := newReconciler(mgr, operatorConfig, dynamic, crd)
	err = newController(mgr, r)
	if err != nil {
		return emperror.Wrap(err, "failed to create controller")
	}
	return nil
}

// newReconciler returns a new IstioReconciler
func newReconciler(mgr manager.Manager, operatorConfig config.Configuration, d dynamic.Interface, crd *crds.CRDReconciler) reconcile.Reconciler {
	return &ReconcileIstio{
		Client:        mgr.GetClient(),
		dynamic:       d,
		crdReconciler: crd,
		mgr:           mgr,
		recorder:      mgr.GetEventRecorderFor("istio-controller"),

		operatorConfig: operatorConfig,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func newController(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	ctrl, err := controller.New("istio-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	if r, ok := r.(IstioReconciler); ok {
		r.setController(ctrl)
		err = r.initWatches(watchCreatedResourcesEvents)
		if err != nil {
			return emperror.Wrapf(err, "could not init watches")
		}
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileIstio{}

// ReconcileConfig reconciles a Config object
type ReconcileIstio struct {
	client.Client
	dynamic          dynamic.Interface
	crdReconciler    *crds.CRDReconciler
	mgr              manager.Manager
	recorder         record.EventRecorder
	ctrl             controller.Controller
	operatorConfig   config.Configuration
	watchersInitOnce sync.Once
}

type ReconcileComponent func(log logr.Logger, istio *istiov1beta1.Istio) error

// +kubebuilder:rbac:groups="",resources=nodes;services;endpoints;pods;replicationcontrollers;services;endpoints;pods,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups="",resources=serviceaccounts;configmaps;pods;events,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups="",resources=namespaces/finalizers,verbs=update
// +kubebuilder:rbac:groups="apps",resources=replicasets,verbs=get;list;watch
// +kubebuilder:rbac:groups="apps",resources=deployments;daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apps",resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="extensions",resources=ingresses;ingresses/status,verbs=*
// +kubebuilder:rbac:groups="extensions",resources=deployments,verbs=get;list;watch
// +kubebuilder:rbac:groups="extensions",resources=deployments/finalizers,verbs=update
// +kubebuilder:rbac:groups="extensions",resources=replicasets,verbs=get;list;watch
// +kubebuilder:rbac:groups="policy",resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="autoscaling",resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=*
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterroles;clusterrolebindings;roles;rolebindings;,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="authentication.k8s.io",resources=tokenreviews,verbs=create
// +kubebuilder:rbac:groups="certificates.k8s.io",resources=certificatesigningrequests;certificatesigningrequests/approval;certificatesigningrequests/status,verbs=update;create;get;delete;watch
// +kubebuilder:rbac:groups="certificates.k8s.io",resources=signers,resourceNames=kubernetes.io/legacy-unknown,verbs=approve
// +kubebuilder:rbac:groups="coordination.k8s.io",resources=leases,verbs=get;list;create;update
// +kubebuilder:rbac:groups="extensions;networking.k8s.io",resources=ingresses,verbs=get;list;watch
// +kubebuilder:rbac:groups="extensions;networking.k8s.io",resources=ingresses/status,verbs=*
// +kubebuilder:rbac:groups="networking.k8s.io",resources=ingressclasses;ingresses,verbs=get;list;watch
// +kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses/status,verbs=*
// +kubebuilder:rbac:groups="networking.x-k8s.io",resources=*,verbs=get;list;watch
// +kubebuilder:rbac:groups="discovery.k8s.io",resources=endpointslices,verbs=get;list;watch

// +kubebuilder:rbac:groups=istio.banzaicloud.io,resources=istios;istios/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=istio.banzaicloud.io,resources=istios/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=authentication.istio.io;cloud.istio.io;config.istio.io;istio.istio.io;networking.istio.io;security.istio.io;scalingpolicy.istio.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations;mutatingwebhookconfigurations,verbs=*
// +kubebuilder:rbac:groups="",resources=secrets;services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="authorization.k8s.io",resources=subjectaccessreviews,verbs=create

// Reconcile reads that state of the cluster for a Config object and makes changes based on the state read
// and what is in the Config.Spec
func (r *ReconcileIstio) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger := log.WithValues("trigger", request.Namespace+"/"+request.Name, "correlationID", uuid.Must(uuid.NewV4()).String())
	// Fetch the Config instance
	config := &istiov1beta1.Istio{}
	err := r.Get(context.TODO(), request.NamespacedName, config)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if !config.Spec.Version.IsSupported() {
		if config.Status.Status == istiov1beta1.Created || config.Status.Status == istiov1beta1.Unmanaged {
			err = errors.New("intended Istio version is unsupported by this version of the operator")
			logger.Error(err, "", "version", config.Spec.Version)
		}
		return reconcile.Result{
			Requeue: false,
		}, nil
	}

	logger.Info("Reconciling Istio")

	// Temporary solution to make sure legacy components are not enabled
	// TODO: delete this when legacy components are removed
	err = r.validateLegacyIstioComponentsAreDisabled(config)
	if err != nil {
		logger.Error(err, "legacy Istio control plane components cannot be enabled starting from Istio 1.6, disable them first")
		return reconcile.Result{
			Requeue: false,
		}, nil
	}

	err = r.autoSetIstioRevisions(config)
	if err != nil {
		updateErr := updateStatus(r.Client, config, istiov1beta1.ReconcileFailed, err.Error(), logger)
		if updateErr != nil {
			logger.Error(updateErr, "failed to update state")
			return reconcile.Result{}, errors.WithStack(err)
		}
		logger.Error(err, "")
		return reconcile.Result{
			RequeueAfter: time.Minute * 5,
		}, nil
	}

	// Set default values where not set
	istiov1beta1.SetDefaults(config)

	r.checkMeshWidePolicyConflict(config, logger)

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

func (r *ReconcileIstio) setController(ctrl controller.Controller) {
	r.ctrl = ctrl
}

func (r *ReconcileIstio) reconcile(logger logr.Logger, config *istiov1beta1.Istio) (reconcile.Result, error) {

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
				RequeueAfter: time.Second * 1,
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

	err := updateStatus(r.Client, config, istiov1beta1.Reconciling, "", logger)
	if err != nil {
		return reconcile.Result{}, errors.WithStack(err)
	}

	logger.Info("reconciling CRDs")
	err = r.crdReconciler.Reconcile(config, logger)
	if err != nil {
		logger.Error(err, "unable to reconcile CRDs")
		return reconcile.Result{}, err
	}

	r.watchersInitOnce.Do(func() {
		nn := types.NamespacedName{
			Namespace: config.Namespace,
			Name:      config.Name,
		}
		err = r.watchMeshWidePolicy(nn)
		if err != nil {
			logger.Error(err, "unable to watch mesh wide policy")
		}
		err = r.watchCRDs(nn)
		if err != nil {
			logger.Error(err, "unable to watch CRDs")
		}
	})

	meshNetworks, err := r.getMeshNetworks(config, logger)
	if err != nil {
		return reconcile.Result{}, err
	}
	config.Spec.SetMeshNetworks(meshNetworks)

	reconcilers := []resources.ComponentReconciler{
		base.New(r.Client, config, false),
		webhookcert.New(r.Client, config, r.operatorConfig),
		citadel.New(citadel.Configuration{
			DeployMeshWidePolicy: true,
		}, r.Client, r.dynamic, config),
		galley.New(r.Client, config),
		sidecarinjector.New(r.Client, config),
		mixer.NewPolicyReconciler(r.Client, r.dynamic, config),
		mixer.NewTelemetryReconciler(r.Client, r.dynamic, config),
		pilot.New(r.Client, r.dynamic, config),
		istiod.New(r.Client, r.dynamic, config, r.mgr.GetScheme(), r.operatorConfig),
		cni.New(r.Client, config),
		nodeagent.New(r.Client, config),
		istiocoredns.New(r.Client, config),
		mixerlesstelemetry.New(r.Client, r.dynamic, config),
		ingressgateway.New(r.Client, r.dynamic, config, false),
		egressgateway.New(r.Client, r.dynamic, config, false),
		meshexpansion.New(r.Client, r.dynamic, config, false),
	}

	for _, rec := range reconcilers {
		err = rec.Reconcile(logger)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	err = k8sutil_mgw.SetGatewayAddress(r.Client, config, config)
	if err != nil {
		log.Info(fmt.Sprintf("ingress gateway address pending: %s", err.Error()))
		updateStatus(r.Client, config, istiov1beta1.ReconcileFailed, err.Error(), logger)
		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: time.Duration(30) * time.Second,
		}, nil
	}

	err = updateStatus(r.Client, config, istiov1beta1.Available, "", logger)
	if err != nil {
		return reconcile.Result{}, errors.WithStack(err)
	}

	logger.Info("reconcile finished")

	// requeue if mesh networks changed
	meshNetworks, err = r.getMeshNetworks(config, logger)
	if err != nil {
		return reconcile.Result{}, err
	}
	if !reflect.DeepEqual(meshNetworks.Networks, config.Spec.GetMeshNetworks().Networks) {
		logger.Info("meshnetwork settings changed, trigger reconciliation")
		return reconcile.Result{
			Requeue: true,
		}, nil
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileIstio) validateLegacyIstioComponentsAreDisabled(config *istiov1beta1.Istio) error {
	if util.PointerToBool(config.Spec.Citadel.Enabled) {
		return errors.New("Citadel cannot be enabled")
	}

	if util.PointerToBool(config.Spec.Galley.Enabled) {
		return errors.New("Galley cannot be enabled")
	}

	if util.PointerToBool(config.Spec.SidecarInjector.Enabled) {
		return errors.New("Sidecar injector cannot be enabled")
	}

	if util.PointerToBool(config.Spec.NodeAgent.Enabled) {
		return errors.New("Node agent cannot be enabled")
	}

	return nil
}

func (r *ReconcileIstio) checkMeshWidePolicyConflict(config *istiov1beta1.Istio, logger logr.Logger) {
	if config.Spec.MTLS != nil && config.Spec.MeshPolicy.MTLSMode != "" {
		mTLS := util.PointerToBool(config.Spec.MTLS)
		if (mTLS && config.Spec.MeshPolicy.MTLSMode != istiov1beta1.STRICT) ||
			(!mTLS && config.Spec.MeshPolicy.MTLSMode != istiov1beta1.PERMISSIVE) {
			warningMessage := fmt.Sprintf(
				"Value '%t' set in spec.mtls is overridden by value '%s' set in spec.meshPolicy.mtlsMode",
				mTLS,
				config.Spec.MeshPolicy.MTLSMode,
			)
			logger.Info(warningMessage)
			r.recorder.Event(
				config,
				"Warning",
				"MeshWidePolicyConflict",
				warningMessage,
			)
		}
	}
}

func (r *ReconcileIstio) getMeshNetworks(config *istiov1beta1.Istio, logger logr.Logger) (*istiov1beta1.MeshNetworks, error) {
	meshNetworks := make(map[string]istiov1beta1.MeshNetwork)

	localNetwork := istiov1beta1.MeshNetwork{
		Endpoints: []istiov1beta1.MeshNetworkEndpoint{
			{
				FromRegistry: config.Spec.ClusterName,
			},
		},
	}

	if len(config.Status.GatewayAddress) > 0 {
		gateways := make([]istiov1beta1.MeshNetworkGateway, 0)
		for _, address := range config.Status.GatewayAddress {
			gateways = append(gateways, istiov1beta1.MeshNetworkGateway{
				Address: address, Port: 15443,
			})
		}
		localNetwork.Gateways = gateways
	}

	meshNetworks[config.Spec.NetworkName] = localNetwork

	remoteIstios := remoteistioCtrl.GetRemoteIstiosByOwnerReference(r.mgr, config, logger)
	for _, remoteIstio := range remoteIstios {
		gateways := make([]istiov1beta1.MeshNetworkGateway, 0)
		if len(remoteIstio.Status.GatewayAddress) > 0 {
			for _, address := range remoteIstio.Status.GatewayAddress {
				gateways = append(gateways, istiov1beta1.MeshNetworkGateway{
					Address: address, Port: 15443,
				})
			}
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

func (r *ReconcileIstio) setCitadelAsOwnerReferenceToIstioSecrets(config *istiov1beta1.Istio, deployment appsv1.Deployment, logger logr.Logger) error {
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

	err = r.Client.List(context.Background(), &secrets, client.InNamespace(config.Namespace))
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

func (r *ReconcileIstio) deleteRemoteIstios(config *istiov1beta1.Istio, logger logr.Logger) {
	remoteIstios := remoteistioCtrl.GetRemoteIstiosByOwnerReference(r.mgr, config, logger)
	for _, remoteIstio := range remoteIstios {
		err := r.Client.Delete(context.Background(), &remoteIstio)
		if err != nil {
			logger.Error(err, "could not delete remote istio resource", "name", remoteIstio.Name)
		}
	}
}

// autoSetIstioRevisions takes care of having only one unrevisioned control plane in each namespace
func (r *ReconcileIstio) autoSetIstioRevisions(config *istiov1beta1.Istio) error {
	yes, err := IsControlPlaneShouldBeRevisioned(r.Client, config)
	if err != nil {
		return err
	}

	if yes {
		if config.Spec.Global != nil && *config.Spec.Global == true {
			return errors.New("there is already an another unrevisioned control plane")
		}
		config.Spec.Global = util.BoolPointer(false)
	}

	return nil
}

func IsControlPlaneShouldBeRevisioned(c client.Client, config *istiov1beta1.Istio) (bool, error) {
	// revision turned on
	if config.IsRevisionUsed() {
		return true, nil
	}

	cps := &istiov1beta1.IstioList{}
	err := c.List(context.Background(), cps, client.InNamespace(config.Namespace))
	if err != nil {
		return false, emperror.Wrap(err, "could not list istio resources")
	}

	sort.Sort(istiov1beta1.SortableIstioItems(cps.Items))

	var oldest *istiov1beta1.Istio

	// collect unrevisioned
	unrevisioned := 0
	for _, cp := range cps.Items {
		if !cp.IsRevisionUsed() {
			if oldest == nil {
				oldest = cp.DeepCopy()
			}
			unrevisioned++
		}
	}

	if unrevisioned > 0 && config.Name != oldest.Name {
		return true, nil
	}

	return false, nil
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
		var actualConfig istiov1beta1.Istio
		err := c.Get(context.TODO(), types.NamespacedName{
			Namespace: config.Namespace,
			Name:      config.Name,
		}, &actualConfig)
		if err != nil {
			return emperror.Wrap(err, "could not get config for updating status")
		}
		actualConfig.Status.Status = status
		actualConfig.Status.ErrorMessage = errorMessage
		err = c.Status().Update(context.Background(), &actualConfig)
		if k8serrors.IsNotFound(err) {
			err = c.Update(context.Background(), &actualConfig)
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

func RemoveFinalizers(c client.Client) error {
	var istios istiov1beta1.IstioList

	err := c.List(context.Background(), &istios)
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

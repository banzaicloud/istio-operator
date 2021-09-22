/*
Copyright 2021 Cisco Systems, Inc. and/or its affiliates.

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

package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"github.com/gogo/protobuf/jsonpb"
	"istio.io/api/mesh/v1alpha1"
	istionetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	istiosecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlBuilder "sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sigs.k8s.io/yaml"

	clusterregistryv1alpha1 "github.com/banzaicloud/cluster-registry/api/v1alpha1"
	servicemeshv1alpha1 "github.com/banzaicloud/istio-operator/v2/api/v1alpha1"
	"github.com/banzaicloud/istio-operator/v2/internal/components"
	"github.com/banzaicloud/istio-operator/v2/internal/components/base"
	"github.com/banzaicloud/istio-operator/v2/internal/components/cni"
	discovery_component "github.com/banzaicloud/istio-operator/v2/internal/components/discovery"
	"github.com/banzaicloud/istio-operator/v2/internal/components/meshexpansion"
	"github.com/banzaicloud/istio-operator/v2/internal/components/sidecarinjector"
	"github.com/banzaicloud/istio-operator/v2/internal/models"
	"github.com/banzaicloud/istio-operator/v2/internal/util"
	"github.com/banzaicloud/istio-operator/v2/pkg/k8sutil"
	pkgUtil "github.com/banzaicloud/istio-operator/v2/pkg/util"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
	"github.com/banzaicloud/operator-tools/pkg/utils"
)

const (
	istioControlPlaneFinalizerID               = "istio-controlplane.servicemesh.cisco.com"
	meshExpansionGatewayRemovalRequeueDuration = time.Second * 30
	readerServiceAccountName                   = "istio-reader"
	// nolint:gosec
	readerSecretType = "k8s.cisco.com/istio-reader-secret"
)

// IstioControlPlaneReconciler reconciles a IstioControlPlane object
type IstioControlPlaneReconciler struct {
	client.Client
	Log                      logr.Logger
	Scheme                   *runtime.Scheme
	ResourceReconciler       reconciler.ResourceReconciler
	ClusterRegistry          models.ClusterRegistryConfiguration
	APIServerEndpointAddress string

	watchersInitOnce sync.Once
	builder          *ctrlBuilder.Builder
	ctrl             controller.Controller
}

// +kubebuilder:rbac:groups="",resources=nodes;replicationcontrollers,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;deletecollection
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups="",resources=configmaps;endpoints;secrets;services;serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations;mutatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apps",resources=replicasets,verbs=get;list;watch
// +kubebuilder:rbac:groups="apps",resources=deployments;daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="authentication.k8s.io",resources=tokenreviews,verbs=create
// +kubebuilder:rbac:groups="authorization.k8s.io",resources=subjectaccessreviews,verbs=create
// +kubebuilder:rbac:groups="autoscaling",resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="certificates.k8s.io",resources=certificatesigningrequests;certificatesigningrequests/approval;certificatesigningrequests/status,verbs=update;create;get;delete;watch
// +kubebuilder:rbac:groups="certificates.k8s.io",resources=signers,resourceNames=kubernetes.io/legacy-unknown,verbs=approve
// +kubebuilder:rbac:groups="coordination.k8s.io",resources=leases,verbs=get;list;create;update
// +kubebuilder:rbac:groups="discovery.k8s.io",resources=endpointslices,verbs=get;list;watch
// +kubebuilder:rbac:groups="extensions",resources=ingresses,verbs=get;list;watch
// +kubebuilder:rbac:groups="extensions",resources=ingresses/status,verbs=*
// +kubebuilder:rbac:groups="multicluster.x-k8s.io",resources=serviceexports,verbs=get;watch;list;create;delete
// +kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses;ingressclasses,verbs=get;list;watch
// +kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses/status,verbs=*
// +kubebuilder:rbac:groups="networking.x-k8s.io",resources=*,verbs=get;list;watch;update
// +kubebuilder:rbac:groups="policy",resources=podsecuritypolicies;poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterroles;clusterrolebindings;roles;rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.istio.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=security.istio.io;telemetry.istio.io;authentication.istio.io;config.istio.io;rbac.istio.io,resources=*,verbs=get;watch;list;update
// +kubebuilder:rbac:groups=servicemesh.cisco.com,resources=istiocontrolplanes;istiomeshes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=servicemesh.cisco.com,resources=istiocontrolplanes/status;istiomeshes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=servicemesh.cisco.com,resources=peeristiocontrolplanes,verbs=list;watch
// +kubebuilder:rbac:groups=clusterregistry.k8s.cisco.com,resources=clusters,verbs=list;watch

func (r *IstioControlPlaneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("istiocontrolplane", req.NamespacedName)

	icp := &servicemeshv1alpha1.IstioControlPlane{}
	err := r.Get(ctx, req.NamespacedName, icp)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if icp.Spec.Version == "" {
		err = errors.New("please set spec.version in your istiocontrolplane CR to be reconciled by this operator")
		logger.Error(err, "", "name", icp.Name, "namespace", icp.Namespace)

		return reconcile.Result{
			Requeue: false,
		}, nil
	}

	if !isIstioVersionSupported(icp.Spec.Version) {
		err = errors.New("intended Istio version is unsupported by this version of the operator")
		logger.Error(err, "", "version", icp.Spec.Version)

		return reconcile.Result{
			Requeue: false,
		}, nil
	}

	result, err := r.reconcile(ctx, icp, logger)
	if err != nil {
		updateErr := components.UpdateStatus(ctx, r.Client, icp, components.ConvertConfigStateToReconcileStatus(servicemeshv1alpha1.ConfigState_ReconcileFailed), err.Error())
		if updateErr != nil {
			logger.Error(updateErr, "failed to update state")

			return result, errors.WithStack(err)
		}

		if result.Requeue {
			return result, nil
		}

		return result, err
	}

	updateErr := components.UpdateStatus(ctx, r.Client, icp, components.ConvertConfigStateToReconcileStatus(servicemeshv1alpha1.ConfigState_Available), "")
	if updateErr != nil {
		logger.Error(updateErr, "failed to update state")

		return result, errors.WithStack(err)
	}

	err = util.RemoveFinalizer(r.Client, icp, istioControlPlaneFinalizerID)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (r *IstioControlPlaneReconciler) reconcile(ctx context.Context, icp *servicemeshv1alpha1.IstioControlPlane, logger logr.Logger) (ctrl.Result, error) {
	logger.Info("reconciling")

	// Get a config to talk to the apiserver
	k8sConfig, err := config.GetConfig()
	if err != nil {
		logger.Error(err, "unable to set up kube client config")
	}

	err = util.AddFinalizer(r.Client, icp, istioControlPlaneFinalizerID)
	if err != nil {
		return ctrl.Result{}, err
	}

	istioMesh, err := r.getRelatedIstioMesh(ctx, r.Client, icp, logger)
	if err != nil {
		return ctrl.Result{}, err
	}

	if icp.GetSpec().GetMode() == servicemeshv1alpha1.ModeType_ACTIVE {
		baseComponent, err := NewComponentReconciler(r, base.NewComponentReconciler, r.Log.WithName("base"))
		if err != nil {
			return ctrl.Result{}, err
		}

		result, err := baseComponent.Reconcile(icp)
		if err != nil {
			return result, err
		}

		r.watchersInitOnce.Do(func() {
			err = r.watchIstioCRs()
			if err != nil {
				logger.Error(err, "unable to watch Istio Custom Resources")
			}
		})
	}

	componentReconcilers := []components.ComponentReconciler{}

	err = setDynamicDefaults(ctx, r.Client, icp, k8sConfig, logger, r.ClusterRegistry.ClusterAPI.Enabled)
	if err != nil {
		return ctrl.Result{}, err
	}

	meshNetworks, err := r.getMeshNetworks(ctx, icp)
	if err != nil {
		return ctrl.Result{}, err
	}

	discoveryReconciler, err := NewComponentReconciler(r, func(helmReconciler *components.HelmReconciler) components.ComponentReconciler {
		return discovery_component.NewChartReconciler(helmReconciler, servicemeshv1alpha1.IstioControlPlaneProperties{
			Mesh:         istioMesh,
			MeshNetworks: meshNetworks,
		}, r.Log)
	}, r.Log.WithName("discovery"))
	if err != nil {
		return ctrl.Result{}, err
	}
	componentReconcilers = append(componentReconcilers, discoveryReconciler)

	cniReconciler, err := NewComponentReconciler(r, cni.NewChartReconciler, r.Log.WithName("cni"))
	if err != nil {
		return ctrl.Result{}, err
	}
	componentReconcilers = append(componentReconcilers, cniReconciler)

	meshExpansionReconciler, err := NewComponentReconciler(r, meshexpansion.NewChartReconciler, r.Log.WithName("meshexpansion"))
	if err != nil {
		return ctrl.Result{}, err
	}
	componentReconcilers = append(componentReconcilers, meshExpansionReconciler)

	sidecarInjectorReconciler, err := NewComponentReconciler(r, sidecarinjector.NewChartReconciler, r.Log.WithName("sidecarInjector"))
	if err != nil {
		return ctrl.Result{}, err
	}
	componentReconcilers = append(componentReconcilers, sidecarInjectorReconciler)

	var result ctrl.Result
	for _, r := range componentReconcilers {
		result, err = r.Reconcile(icp)
		if err != nil {
			return result, err
		}
	}

	err = r.reconcileIstiodEndpoint(ctx, icp)
	if err != nil {
		return result, err
	}

	err = r.reconcileClusterReaderSecret(ctx, icp, k8sConfig)
	if err != nil {
		return result, err
	}

	// icp is marked for deletion
	if !icp.DeletionTimestamp.IsZero() {
		err = r.waitForMeshExpansionGatewayRemoval(ctx, icp)
		if err != nil {
			result.Requeue = true
			result.RequeueAfter = meshExpansionGatewayRemovalRequeueDuration

			return result, err
		}

		return result, nil
	}

	err = r.setSidecarInjectorChecksumToStatus(ctx, icp)
	if err != nil {
		return result, err
	}

	r.setControlPlaneNameToStatus(icp)

	err = r.setMeshConfigToStatus(ctx, icp)
	if err != nil {
		return result, err
	}

	err = r.setIstiodAddressesToStatus(ctx, icp)
	if err != nil {
		return result, err
	}

	err = r.setMeshExpansionGWAddressToStatus(ctx, icp)
	if err != nil {
		logger.Info(fmt.Sprintf("mesh expansion gateway is pending: %s", err.Error()))
		_ = components.UpdateStatus(ctx, r.Client, icp, components.ConvertConfigStateToReconcileStatus(servicemeshv1alpha1.ConfigState_ReconcileFailed), err.Error())
		result.Requeue = true
		result.RequeueAfter = pendingGatewayRequeueDuration

		return result, err
	}

	return result, nil
}

func (r *IstioControlPlaneReconciler) GetClient() client.Client {
	return r.Client
}

func (r *IstioControlPlaneReconciler) GetScheme() *runtime.Scheme {
	return r.Scheme
}

func (r *IstioControlPlaneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.builder = ctrl.NewControllerManagedBy(mgr)

	ctrl, err := r.builder.
		For(&servicemeshv1alpha1.IstioControlPlane{
			TypeMeta: metav1.TypeMeta{
				Kind:       "IstioControlPlane",
				APIVersion: servicemeshv1alpha1.SchemeBuilder.GroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(util.ObjectChangePredicate{})).
		Owns(&appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: appsv1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(util.ObjectChangePredicate{})).
		Owns(&appsv1.DaemonSet{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DaemonSet",
				APIVersion: appsv1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(util.ObjectChangePredicate{})).
		Owns(&corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(util.ObjectChangePredicate{})).
		Owns(&corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(util.ObjectChangePredicate{})).
		Owns(&corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(util.ObjectChangePredicate{})).
		Owns(&corev1.Endpoints{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Endpoints",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(util.ObjectChangePredicate{})).
		Owns(&corev1.ServiceAccount{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ServiceAccount",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(util.ObjectChangePredicate{})).
		Owns(&policyv1beta1.PodSecurityPolicy{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PodSecurityPolicy",
				APIVersion: policyv1beta1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(util.ObjectChangePredicate{})).
		Owns(&policyv1beta1.PodDisruptionBudget{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PodDisruptionBudget",
				APIVersion: policyv1beta1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(util.ObjectChangePredicate{})).
		Owns(&rbacv1.Role{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Role",
				APIVersion: rbacv1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(util.ObjectChangePredicate{})).
		Owns(&rbacv1.RoleBinding{
			TypeMeta: metav1.TypeMeta{
				Kind:       "RoleBinding",
				APIVersion: rbacv1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(util.ObjectChangePredicate{})).
		Owns(&autoscalingv1.HorizontalPodAutoscaler{
			TypeMeta: metav1.TypeMeta{
				Kind:       "HorizontalPodAutoscaler",
				APIVersion: autoscalingv1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(util.ObjectChangePredicate{
			CalculateOptions: []util.CalculateOption{
				util.IgnoreMetadataAnnotations("autoscaling.alpha.kubernetes.io"),
				patch.IgnoreStatusFields(),
			},
		})).
		Build(r)
	if err != nil {
		return err
	}

	r.ctrl = ctrl

	types := []client.Object{
		&rbacv1.ClusterRole{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ClusterRole",
				APIVersion: rbacv1.SchemeGroupVersion.String(),
			},
		},
		&rbacv1.ClusterRoleBinding{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ClusterRoleBinding",
				APIVersion: rbacv1.SchemeGroupVersion.String(),
			},
		},
		&admissionregistrationv1.MutatingWebhookConfiguration{
			TypeMeta: metav1.TypeMeta{
				Kind:       "MutatingWebhookConfiguration",
				APIVersion: admissionregistrationv1.SchemeGroupVersion.String(),
			},
		},
		&admissionregistrationv1.ValidatingWebhookConfiguration{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ValidatingWebhookConfiguration",
				APIVersion: admissionregistrationv1.SchemeGroupVersion.String(),
			},
		},
	}

	for _, t := range types {
		err := r.ctrl.Watch(&source.Kind{Type: t}, handler.EnqueueRequestsFromMapFunc(reconciler.EnqueueByOwnerAnnotationMapper()), util.ObjectChangePredicate{})
		if err != nil {
			return err
		}
	}

	err = r.ctrl.Watch(
		&source.Kind{
			Type: &servicemeshv1alpha1.IstioMeshGateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       "IstioMeshGateway",
					APIVersion: servicemeshv1alpha1.SchemeBuilder.GroupVersion.String(),
				},
			},
		},
		handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []reconcile.Request {
			var imgw *servicemeshv1alpha1.IstioMeshGateway
			var ok bool
			if imgw, ok = obj.(*servicemeshv1alpha1.IstioMeshGateway); !ok {
				return nil
			}

			icp := imgw.GetSpec().GetIstioControlPlane()
			if icp == nil {
				return nil
			}

			// only act on mesh expansion gateway
			sel := labels.SelectorFromValidatedSet((&servicemeshv1alpha1.IstioControlPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      icp.Name,
					Namespace: icp.Namespace,
				},
			}).MeshExpansionGatewayLabels())
			if !sel.Matches(labels.Set(imgw.GetLabels())) {
				return nil
			}

			r.Log.V(1).Info("trigger reconcile by mesh expansion gateway change")

			return []reconcile.Request{
				{
					NamespacedName: client.ObjectKey{
						Name:      icp.GetName(),
						Namespace: icp.GetNamespace(),
					},
				},
			}
		}),
		predicate.Or(
			util.ObjectChangePredicate{},
			util.IMGWAddressChangePredicate{},
		),
	)
	if err != nil {
		return err
	}

	err = r.ctrl.Watch(
		&source.Kind{
			Type: &servicemeshv1alpha1.PeerIstioControlPlane{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PeerIstioControlPlane",
					APIVersion: servicemeshv1alpha1.SchemeBuilder.GroupVersion.String(),
				},
			},
		},
		handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []reconcile.Request {
			var picp *servicemeshv1alpha1.PeerIstioControlPlane
			var ok bool
			if picp, ok = obj.(*servicemeshv1alpha1.PeerIstioControlPlane); !ok {
				return nil
			}

			return []reconcile.Request{
				{
					NamespacedName: client.ObjectKey{
						Name:      picp.GetStatus().IstioControlPlaneName,
						Namespace: picp.GetNamespace(),
					},
				},
			}
		}),
		predicate.Or(
			util.ObjectChangePredicate{},
			util.PICPStatusChangePredicate{},
		),
	)
	if err != nil {
		return err
	}

	err = r.ctrl.Watch(
		&source.Kind{
			Type: &servicemeshv1alpha1.IstioMesh{
				TypeMeta: metav1.TypeMeta{
					Kind:       "IstioMesh",
					APIVersion: servicemeshv1alpha1.SchemeBuilder.GroupVersion.String(),
				},
			},
		},
		handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []reconcile.Request {
			var imesh *servicemeshv1alpha1.IstioMesh
			var ok bool
			if imesh, ok = obj.(*servicemeshv1alpha1.IstioMesh); !ok {
				return nil
			}

			icps := &servicemeshv1alpha1.IstioControlPlaneList{}
			err := r.Client.List(context.Background(), icps)
			if err != nil {
				r.Log.Error(err, "could not list Istio control plane resources")

				return nil
			}

			resources := make([]reconcile.Request, 0)
			for _, icp := range icps.Items {
				if imesh.GetName() == icp.GetSpec().GetMeshID() {
					resources = append(resources, reconcile.Request{
						NamespacedName: client.ObjectKey{
							Name:      icp.GetName(),
							Namespace: icp.GetNamespace(),
						},
					})
				}
			}

			return resources
		}),
		util.ObjectChangePredicate{},
	)
	if err != nil {
		return err
	}

	if r.ClusterRegistry.ClusterAPI.Enabled {
		err = r.ctrl.Watch(
			&source.Kind{
				Type: &clusterregistryv1alpha1.Cluster{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Cluster",
						APIVersion: clusterregistryv1alpha1.SchemeBuilder.GroupVersion.String(),
					},
				},
			},
			handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []reconcile.Request {
				var cluster *clusterregistryv1alpha1.Cluster
				var ok bool
				if cluster, ok = obj.(*clusterregistryv1alpha1.Cluster); !ok {
					return nil
				}

				if cluster.Status.Type != clusterregistryv1alpha1.ClusterTypeLocal {
					return nil
				}

				icps := &servicemeshv1alpha1.IstioControlPlaneList{}
				err := r.Client.List(context.Background(), icps)
				if err != nil {
					r.Log.Error(err, "could not list Istio control plane resources")

					return nil
				}

				resources := make([]reconcile.Request, len(icps.Items))
				for i, icp := range icps.Items {
					resources[i] = reconcile.Request{
						NamespacedName: client.ObjectKey{
							Name:      icp.GetName(),
							Namespace: icp.GetNamespace(),
						},
					}
				}

				r.Log.V(1).Info("trigger reconcile by cluster change")

				return resources
			}),
			predicate.Or(
				util.ObjectChangePredicate{},
				util.ClusterTypeChangePredicate{},
			),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *IstioControlPlaneReconciler) getMeshNetworks(ctx context.Context, icp *servicemeshv1alpha1.IstioControlPlane) (*v1alpha1.MeshNetworks, error) {
	networks := make(map[string]*v1alpha1.Network)

	type ControlPlane interface {
		GetName() string
		GetStatus() servicemeshv1alpha1.IstioControlPlaneStatus
		GetSpec() *servicemeshv1alpha1.IstioControlPlaneSpec
	}

	cps := make([]ControlPlane, 0)
	cps = append(cps, icp)

	picpList := &servicemeshv1alpha1.PeerIstioControlPlaneList{}
	err := r.GetClient().List(ctx, picpList, client.InNamespace(icp.GetNamespace()))
	if err != nil {
		return nil, errors.WithStackIf(err)
	}

	for _, picp := range picpList.Items {
		picp := picp
		if picp.Status.IstioControlPlaneName == icp.GetName() {
			cps = append(cps, &picp)
		}
	}

	for _, cp := range cps {
		gateways := make([]*v1alpha1.Network_IstioNetworkGateway, 0)
		for _, address := range cp.GetStatus().GatewayAddress {
			gateways = append(gateways, &v1alpha1.Network_IstioNetworkGateway{
				Gw: &v1alpha1.Network_IstioNetworkGateway_Address{
					Address: address,
				},
				Port: 15443, // nolint:gomnd
			})
		}

		networkName := cp.GetSpec().GetNetworkName()
		if networks[networkName] == nil {
			networks[networkName] = &v1alpha1.Network{}
		}
		networks[networkName].Endpoints = append(networks[networkName].Endpoints, &v1alpha1.Network_NetworkEndpoints{
			Ne: &v1alpha1.Network_NetworkEndpoints_FromRegistry{
				FromRegistry: cp.GetSpec().GetClusterID(),
			},
		})
		networks[networkName].Gateways = append(networks[networkName].Gateways, gateways...)
	}

	return &v1alpha1.MeshNetworks{
		Networks: networks,
	}, nil
}

func (r *IstioControlPlaneReconciler) getRelatedIstioMesh(ctx context.Context, c client.Client, icp *servicemeshv1alpha1.IstioControlPlane, logger logr.Logger) (*servicemeshv1alpha1.IstioMesh, error) {
	mesh := &servicemeshv1alpha1.IstioMesh{}

	err := c.Get(ctx, client.ObjectKey{
		Name:      icp.GetSpec().GetMeshID(),
		Namespace: icp.GetNamespace(),
	}, mesh)
	if k8serrors.IsNotFound(err) {
		return mesh, nil
	}
	if err != nil {
		updateErr := components.UpdateStatus(ctx, c, icp, components.ConvertConfigStateToReconcileStatus(servicemeshv1alpha1.ConfigState_ReconcileFailed), err.Error())
		if updateErr != nil {
			logger.Error(updateErr, "failed to update Istio control plane state")

			return nil, errors.WithStack(err)
		}

		return nil, errors.WrapIf(err, "could not get related Istio control plane")
	}

	return mesh, nil
}

// reconcileIstiodEndpoint creates the k8s Endpoint resource for the headless istiod service
// on PASSIVE Istio Control planes to be able to connect to istiod pods on active clusters
func (r *IstioControlPlaneReconciler) reconcileIstiodEndpoint(ctx context.Context, icp *servicemeshv1alpha1.IstioControlPlane) error {
	if !icp.DeletionTimestamp.IsZero() || icp.GetSpec().GetMode() != servicemeshv1alpha1.ModeType_PASSIVE {
		// In active mode the k8s endpoint controller takes care of creating/updating the Endpoint
		// resource based on the istiod service with selector, so istio operator does nothing
		return nil
	}

	serviceName := icp.WithRevision("istiod")
	serviceNamespace := icp.GetNamespace()

	istiodEndpointAddresses, err := pkgUtil.GetIstiodEndpointAddresses(ctx, r.Client, icp.GetName(), icp.GetSpec().GetNetworkName(), serviceNamespace)
	if err != nil {
		return errors.WithStackIf(err)
	}
	if len(istiodEndpointAddresses) == 0 {
		return errors.New("no valid istiod address found")
	}

	istiodEndpointPorts, err := pkgUtil.GetIstiodEndpointPorts(ctx, r.Client, serviceName, serviceNamespace)
	if err != nil {
		return errors.WithStackIf(err)
	}

	endpoints := k8sutil.CreateK8sEndpoints(serviceName, serviceNamespace, istiodEndpointAddresses, istiodEndpointPorts)
	pkgUtil.SetICPMetadataOnObject(endpoints, icp)

	_, err = r.ResourceReconciler.ReconcileResource(endpoints, reconciler.StatePresent)
	if err != nil {
		return errors.WithStackIf(err)
	}

	return nil
}

func (r *IstioControlPlaneReconciler) reconcileClusterReaderSecret(ctx context.Context, icp *servicemeshv1alpha1.IstioControlPlane, kubeConfig *rest.Config) error {
	var err error
	state := reconciler.StateAbsent
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      icp.WithRevision(strings.ToLower(icp.GetSpec().GetClusterID())),
			Namespace: icp.GetNamespace(),
		},
	}

	if icp.DeletionTimestamp.IsZero() && icp.GetSpec().GetMode() == servicemeshv1alpha1.ModeType_PASSIVE {
		state = reconciler.StatePresent
		secret, err = pkgUtil.GetReaderSecretForCluster(
			ctx,
			r.Client,
			kubeConfig,
			icp.GetSpec().GetClusterID(),
			types.NamespacedName{
				Name:      secret.GetName(),
				Namespace: secret.GetNamespace(),
			},
			types.NamespacedName{
				Name:      icp.WithRevision(readerServiceAccountName),
				Namespace: icp.GetNamespace(),
			},
			r.APIServerEndpointAddress,
			r.ClusterRegistry.ClusterAPI.Enabled,
		)
		if err != nil {
			return errors.WithStackIf(err)
		}

		secret.Type = readerSecretType
		pkgUtil.SetICPMetadataOnObject(secret, icp)
	}

	_, err = r.ResourceReconciler.ReconcileResource(secret, state)
	if err != nil {
		return errors.WithStackIf(err)
	}

	return errors.WithStackIf(err)
}

func (r *IstioControlPlaneReconciler) waitForMeshExpansionGatewayRemoval(ctx context.Context, icp *servicemeshv1alpha1.IstioControlPlane) error {
	if icp.DeletionTimestamp.IsZero() || !utils.PointerToBool(icp.GetSpec().GetMeshExpansion().GetEnabled()) {
		return nil
	}

	l := &servicemeshv1alpha1.IstioMeshGatewayList{}
	err := r.Client.List(ctx, l, client.InNamespace(icp.GetNamespace()), client.MatchingLabels(icp.MeshExpansionGatewayLabels()))
	if err != nil {
		return err
	}

	if len(l.Items) > 0 {
		return errors.New("mesh expansion gateway still exists")
	}

	return nil
}

func (r *IstioControlPlaneReconciler) setIstiodAddressesToStatus(ctx context.Context, icp *servicemeshv1alpha1.IstioControlPlane) error {
	if icp.GetSpec().GetMode() != servicemeshv1alpha1.ModeType_ACTIVE {
		// istiod pods should only be present on ACTIVE clusters so it only make sense set pod IPs on those clusters
		icp.Status.IstiodAddresses = nil

		return nil
	}

	endpoints, err := k8sutil.GetEndpoints(ctx, r.Client, icp.WithRevision("istiod"), icp.GetNamespace())
	if err != nil {
		return errors.WithStackIf(err)
	}
	icp.Status.IstiodAddresses = k8sutil.GetIPsForEndpoints(endpoints)

	return nil
}

func (r *IstioControlPlaneReconciler) setMeshExpansionGWAddressToStatus(ctx context.Context, icp *servicemeshv1alpha1.IstioControlPlane) error {
	if icp.DeletionTimestamp.IsZero() && !utils.PointerToBool(icp.GetSpec().GetMeshExpansion().GetEnabled()) {
		icp.Status.GatewayAddress = nil

		return nil
	}

	l := &servicemeshv1alpha1.IstioMeshGatewayList{}
	err := r.Client.List(ctx, l, client.InNamespace(icp.GetNamespace()), client.MatchingLabels(
		utils.MergeLabels(icp.RevisionLabels(), map[string]string{
			"app": "istio-meshexpansion-gateway",
		}),
	))
	if err != nil {
		return err
	}

	if len(l.Items) == 0 {
		return errors.New("could not find mesh expansion gateway")
	}

	if len(l.Items) > 1 {
		return errors.New("multiple mesh expansion gateways were found")
	}

	imgw := l.Items[0]
	if imgw.GetStatus().Status != servicemeshv1alpha1.ConfigState_Available {
		return errors.New(imgw.GetStatus().ErrorMessage)
	}

	icp.Status.GatewayAddress = imgw.GetStatus().GatewayAddress

	return nil
}

func (r *IstioControlPlaneReconciler) setControlPlaneNameToStatus(icp *servicemeshv1alpha1.IstioControlPlane) {
	icp.Status.IstioControlPlaneName = icp.GetName()
}

func (r *IstioControlPlaneReconciler) setMeshConfigToStatus(ctx context.Context, icp *servicemeshv1alpha1.IstioControlPlane) error {
	configmaps := &corev1.ConfigMapList{}
	err := r.Client.List(ctx, configmaps, client.InNamespace(icp.GetNamespace()), client.MatchingLabels(utils.MergeLabels(icp.RevisionLabels(), map[string]string{"istio": "meshconfig"})))
	if err != nil {
		return errors.WithStackIf(err)
	}

	if len(configmaps.Items) != 1 {
		return nil
	}

	var mc v1alpha1.MeshConfig

	mcYAML := configmaps.Items[0].Data["mesh"]
	mcJSON, err := yaml.YAMLToJSON([]byte(mcYAML))
	if err != nil {
		return errors.WithStackIf(err)
	}

	err = jsonpb.UnmarshalString(string(mcJSON), &mc)
	if err != nil {
		return errors.WithStackIf(err)
	}

	icp.Status.MeshConfig = &mc

	cs := icp.Status.GetChecksums()
	if cs == nil {
		cs = &servicemeshv1alpha1.StatusChecksums{}
	}
	cs.MeshConfig = fmt.Sprintf("%x", sha256.Sum256([]byte(mcYAML)))
	icp.Status.Checksums = cs

	return nil
}

func (r *IstioControlPlaneReconciler) setSidecarInjectorChecksumToStatus(ctx context.Context, icp *servicemeshv1alpha1.IstioControlPlane) error {
	configmaps := &corev1.ConfigMapList{}
	err := r.Client.List(ctx, configmaps, client.InNamespace(icp.GetNamespace()), client.MatchingLabels(utils.MergeLabels(icp.RevisionLabels(), map[string]string{"istio": "sidecar-injector"})))
	if err != nil {
		return err
	}

	if len(configmaps.Items) == 1 {
		cm := configmaps.Items[0]
		jm, err := json.Marshal(cm.Data)
		if err != nil {
			return err
		}

		cs := icp.Status.GetChecksums()
		if cs == nil {
			cs = &servicemeshv1alpha1.StatusChecksums{}
		}
		cs.SidecarInjector = fmt.Sprintf("%x", sha256.Sum256(jm))
		icp.Status.Checksums = cs
	}

	return nil
}

func (r *IstioControlPlaneReconciler) watchIstioCRs() error {
	if r.ctrl == nil {
		return errors.New("ctrl is not set")
	}

	eventHandler := &handler.EnqueueRequestForOwner{
		OwnerType: &servicemeshv1alpha1.IstioControlPlane{
			TypeMeta: metav1.TypeMeta{
				Kind:       "IstioControlPlane",
				APIVersion: servicemeshv1alpha1.SchemeBuilder.GroupVersion.String(),
			},
		},
		IsController: true,
	}

	types := []client.Object{
		&istionetworkingv1alpha3.EnvoyFilter{
			TypeMeta: metav1.TypeMeta{
				Kind:       "EnvoyFilter",
				APIVersion: istionetworkingv1alpha3.SchemeGroupVersion.String(),
			},
		},
		&istiosecurityv1beta1.PeerAuthentication{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PeerAuthentication",
				APIVersion: istiosecurityv1beta1.SchemeGroupVersion.String(),
			},
		},
	}

	for _, t := range types {
		err := r.ctrl.Watch(&source.Kind{Type: t}, eventHandler, util.ObjectChangePredicate{})
		if err != nil {
			return err
		}
	}

	return nil
}

func RemoveFinalizers(c client.Client) error {
	var icps servicemeshv1alpha1.IstioControlPlaneList
	err := c.List(context.Background(), &icps)
	if err != nil {
		return errors.WrapIf(err, "could not list Istio control plane resources")
	}

	for _, istio := range icps.Items {
		istio := istio
		err = util.RemoveFinalizer(c, &istio, istioControlPlaneFinalizerID)
		if err != nil {
			return err
		}
	}

	var imgws servicemeshv1alpha1.IstioMeshGatewayList
	err = c.List(context.Background(), &imgws)
	if err != nil {
		return errors.WrapIf(err, "could not list istio mesh gateway resources")
	}

	for _, imgw := range imgws.Items {
		imgw := imgw
		err = util.RemoveFinalizer(c, &imgw, istioMeshGatewayFinalizerID)
		if err != nil {
			return err
		}
	}

	return nil
}

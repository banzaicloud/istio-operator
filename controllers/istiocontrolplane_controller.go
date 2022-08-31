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
	"sort"
	"strings"
	"sync"
	"time"

	"emperror.dev/errors"
	"github.com/golang/protobuf/jsonpb"
	"istio.io/api/mesh/v1alpha1"
	istionetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	istiosecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
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

	servicemeshv1alpha1 "github.com/banzaicloud/istio-operator/api/v2/v1alpha1"
	"github.com/banzaicloud/istio-operator/v2/internal/components"
	"github.com/banzaicloud/istio-operator/v2/internal/components/base"
	"github.com/banzaicloud/istio-operator/v2/internal/components/cni"
	discovery_component "github.com/banzaicloud/istio-operator/v2/internal/components/discovery"
	"github.com/banzaicloud/istio-operator/v2/internal/components/meshexpansion"
	"github.com/banzaicloud/istio-operator/v2/internal/components/resourcesyncrule"
	"github.com/banzaicloud/istio-operator/v2/internal/components/sidecarinjector"
	"github.com/banzaicloud/istio-operator/v2/internal/models"
	"github.com/banzaicloud/istio-operator/v2/internal/util"
	"github.com/banzaicloud/istio-operator/v2/pkg/k8sutil"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/banzaicloud/operator-tools/pkg/logger"
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
	"github.com/banzaicloud/operator-tools/pkg/utils"
	clusterregistryv1alpha1 "github.com/cisco-open/cluster-registry-controller/api/v1alpha1"
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
	Log                      logger.Logger
	Scheme                   *runtime.Scheme
	ResourceReconciler       reconciler.ResourceReconciler
	ClusterRegistry          models.ClusterRegistryConfiguration
	APIServerEndpointAddress string
	SupportedIstioVersion    string
	Version                  string
	Recorder                 record.EventRecorder

	watchersInitOnce sync.Once
	builder          *ctrlBuilder.Builder
	ctrl             controller.Controller
}

// +kubebuilder:rbac:groups="",resources=nodes;replicationcontrollers,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;deletecollection;delete;patch;update
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups="",resources=configmaps;endpoints;secrets;services;serviceaccounts;resourcequotas,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations;mutatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apps",resources=replicasets,verbs=get;list;watch
// +kubebuilder:rbac:groups="apps",resources=deployments;daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="authentication.k8s.io",resources=tokenreviews,verbs=create
// +kubebuilder:rbac:groups="authorization.k8s.io",resources=subjectaccessreviews,verbs=create
// +kubebuilder:rbac:groups="autoscaling",resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="certificates.k8s.io",resources=certificatesigningrequests;certificatesigningrequests/approval;certificatesigningrequests/status,verbs=update;create;get;delete;watch
// +kubebuilder:rbac:groups="certificates.k8s.io",resources=signers,verbs=approve
// +kubebuilder:rbac:groups="coordination.k8s.io",resources=leases,verbs=get;list;create;update
// +kubebuilder:rbac:groups="discovery.k8s.io",resources=endpointslices,verbs=get;list;watch
// +kubebuilder:rbac:groups="extensions",resources=ingresses,verbs=get;list;watch
// +kubebuilder:rbac:groups="extensions",resources=ingresses/status,verbs=*
// +kubebuilder:rbac:groups="multicluster.x-k8s.io",resources=serviceexports,verbs=get;watch;list;create;delete
// +kubebuilder:rbac:groups="multicluster.x-k8s.io",resources=serviceimports,verbs=get;watch;list
// +kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses;ingressclasses,verbs=get;list;watch
// +kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses/status,verbs=*
// +kubebuilder:rbac:groups="gateway.networking.k8s.io",resources=*,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups="networking.x-k8s.io",resources=*,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups="policy",resources=podsecuritypolicies;poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterroles;clusterrolebindings;roles;rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.istio.io;security.istio.io;telemetry.istio.io;authentication.istio.io;config.istio.io;rbac.istio.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="extensions.istio.io",resources=*,verbs=get;list;watch
// +kubebuilder:rbac:groups=servicemesh.cisco.com,resources=istiocontrolplanes;peeristiocontrolplanes;istiomeshes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=servicemesh.cisco.com,resources=istiocontrolplanes/status;peeristiocontrolplanes/status;istiomeshes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=clusterregistry.k8s.cisco.com,resources=clusters,verbs=list;watch
// +kubebuilder:rbac:groups=clusterregistry.k8s.cisco.com,resources=resourcesyncrules;clusterfeatures,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="gateway.networking.k8s.io",resources=gatewayclasses,verbs=create;update;patch;delete

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

	if !IsIstioVersionSupported(icp.Spec.Version) {
		err = errors.New("intended Istio version is unsupported by this version of the operator")
		logger.Error(err, "", "version", icp.Spec.Version)

		return reconcile.Result{
			Requeue: false,
		}, nil
	}

	if requeueNeeded, err := k8sutil.IsReqeueNeededCosNamespaceTermination(ctx, r.GetClient(), icp); requeueNeeded && err == nil {
		logger.Info("namespace is terminating, requeue needed")

		return ctrl.Result{
			RequeueAfter: nsTerminationRequeueDuration,
		}, nil
	} else if err != nil {
		return ctrl.Result{}, err
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
	if updateErr != nil && !k8serrors.IsNotFound(updateErr) {
		logger.Error(updateErr, "failed to update state")

		return result, errors.WithStack(err)
	}

	err = util.RemoveFinalizer(ctx, r.Client, icp, istioControlPlaneFinalizerID, true)
	if err != nil {
		return result, errors.WithStack(err)
	}

	return result, nil
}

func (r *IstioControlPlaneReconciler) reconcile(ctx context.Context, icp *servicemeshv1alpha1.IstioControlPlane, logger logger.Logger) (ctrl.Result, error) {
	logger.Info("reconciling")

	// Get a config to talk to the apiserver
	k8sConfig, err := config.GetConfig()
	if err != nil {
		logger.Error(err, "unable to set up kube client config")
	}

	err = util.AddFinalizer(ctx, r.Client, icp, istioControlPlaneFinalizerID)
	if err != nil {
		return ctrl.Result{}, err
	}

	istioMesh, err := r.getRelatedIstioMesh(ctx, r.Client, icp, logger)
	if err != nil {
		return ctrl.Result{}, err
	}

	if icp.GetSpec().GetMode() == servicemeshv1alpha1.ModeType_ACTIVE {
		baseComponent, err := NewComponentReconciler(r, func(helmReconciler *components.HelmReconciler) components.ComponentReconciler {
			return base.NewComponentReconciler(helmReconciler, r.Log.WithName("base"), r.SupportedIstioVersion)
		}, r.Log.WithName("base"))
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

	// set cluster ID to status as it is not always in the stored spec
	icp.GetStatus().ClusterID = icp.Spec.ClusterID

	meshNetworks, err := r.getMeshNetworks(ctx, icp)
	if err != nil {
		return ctrl.Result{}, err
	}

	trustedCACertificates, err := r.getCACertificatesFromPeers(ctx, icp)
	if err != nil {
		return ctrl.Result{}, err
	}

	discoveryReconciler, err := NewComponentReconciler(r, func(helmReconciler *components.HelmReconciler) components.ComponentReconciler {
		return discovery_component.NewChartReconciler(helmReconciler, servicemeshv1alpha1.IstioControlPlaneProperties{
			Mesh:                         istioMesh,
			MeshNetworks:                 meshNetworks,
			TrustedRootCACertificatePEMs: trustedCACertificates,
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

	resourceSyncRuleReconciler, err := NewComponentReconciler(r, func(helmReconciler *components.HelmReconciler) components.ComponentReconciler {
		return resourcesyncrule.NewChartReconciler(helmReconciler, r.ClusterRegistry.ResourceSyncRules.Enabled)
	}, r.Log.WithName("resourcesyncrule"))
	if err != nil {
		return ctrl.Result{}, err
	}
	componentReconcilers = append(componentReconcilers, resourceSyncRuleReconciler)

	var result ctrl.Result
	for _, r := range componentReconcilers {
		result, err = r.Reconcile(icp)
		if err != nil {
			return result, err
		}
	}

	err = r.deleteIstioRootCAConfigmapsOnPassive(ctx, icp, logger)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.reconcileNamespaceInjectionLabels(ctx, icp)
	if err != nil {
		return result, err
	}

	err = r.setInjectionNamespacesToStatus(ctx, icp)
	if err != nil {
		return result, err
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

		if err := r.removeFinalizerFromRelatedMeshGateways(ctx, icp); err != nil {
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

	err = r.setIstioCARootCertToStatus(ctx, icp)
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

	objectChangePredicate := util.ObjectChangePredicate{Logger: r.Log}

	ctrl, err := r.builder.
		For(&servicemeshv1alpha1.IstioControlPlane{
			TypeMeta: metav1.TypeMeta{
				Kind:       "IstioControlPlane",
				APIVersion: servicemeshv1alpha1.SchemeBuilder.GroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(objectChangePredicate)).
		Owns(&appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: appsv1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(objectChangePredicate)).
		Owns(&appsv1.DaemonSet{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DaemonSet",
				APIVersion: appsv1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(objectChangePredicate)).
		Owns(&corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(objectChangePredicate)).
		Owns(&corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(objectChangePredicate)).
		Owns(&corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(objectChangePredicate)).
		Owns(&corev1.Endpoints{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Endpoints",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(objectChangePredicate)).
		Owns(&corev1.ServiceAccount{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ServiceAccount",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(objectChangePredicate)).
		Owns(&policyv1.PodDisruptionBudget{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PodDisruptionBudget",
				APIVersion: policyv1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(objectChangePredicate)).
		Owns(&rbacv1.Role{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Role",
				APIVersion: rbacv1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(objectChangePredicate)).
		Owns(&rbacv1.RoleBinding{
			TypeMeta: metav1.TypeMeta{
				Kind:       "RoleBinding",
				APIVersion: rbacv1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(objectChangePredicate)).
		Owns(&autoscalingv1.HorizontalPodAutoscaler{
			TypeMeta: metav1.TypeMeta{
				Kind:       "HorizontalPodAutoscaler",
				APIVersion: autoscalingv1.SchemeGroupVersion.String(),
			},
		}, ctrlBuilder.WithPredicates(util.ObjectChangePredicate{
			Logger: r.Log,
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

	if r.ClusterRegistry.ResourceSyncRules.Enabled {
		types = append(types, []client.Object{
			&clusterregistryv1alpha1.ResourceSyncRule{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ResourceSyncRule",
					APIVersion: clusterregistryv1alpha1.SchemeBuilder.GroupVersion.String(),
				},
			},
			&clusterregistryv1alpha1.ClusterFeature{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ClusterFeature",
					APIVersion: clusterregistryv1alpha1.SchemeBuilder.GroupVersion.String(),
				},
			},
		}...)
	}

	for _, t := range types {
		err := r.ctrl.Watch(&source.Kind{Type: t}, handler.EnqueueRequestsFromMapFunc(reconciler.EnqueueByOwnerAnnotationMapper()), objectChangePredicate)
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
			objectChangePredicate,
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
			util.ObjectChangePredicate{
				Logger: r.Log,
				CalculateOptions: []patch.CalculateOption{
					util.IgnoreMetadataAnnotations(patch.LastAppliedConfig),
				},
			},
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
		objectChangePredicate,
	)
	if err != nil {
		return err
	}

	err = r.ctrl.Watch(
		&source.Kind{
			Type: &corev1.Namespace{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Namespace",
					APIVersion: corev1.SchemeGroupVersion.String(),
				},
			},
		},
		handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []reconcile.Request {
			var ok bool
			var revision string
			if revision, ok = obj.GetLabels()[servicemeshv1alpha1.RevisionedAutoInjectionLabel]; !ok {
				resources, err := r.getICPReconcileRequests(context.Background())
				if err != nil {
					r.Log.Error(err, "")

					return nil
				}

				return resources
			}

			nn := servicemeshv1alpha1.NamespacedNameFromRevision(revision)
			if nn.Namespace != "" {
				r.Log.Info("trigger reconcile by namespace change")

				return []reconcile.Request{
					{
						NamespacedName: nn,
					},
				}
			}

			return nil
		}),
		util.NamespaceRevisionLabelChange{},
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

				resources, err := r.getICPReconcileRequests(context.Background())
				if err != nil {
					r.Log.Error(err, "could not list Istio control plane resources")

					return nil
				}

				r.Log.V(1).Info("trigger reconcile by cluster change")

				return resources
			}),
			predicate.Or(
				objectChangePredicate,
				util.ClusterTypeChangePredicate{},
			),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

type ControlPlane interface {
	GetName() string
	GetStatus() *servicemeshv1alpha1.IstioControlPlaneStatus
	GetSpec() *servicemeshv1alpha1.IstioControlPlaneSpec
}

type SortableControlPlanes []ControlPlane

func (list SortableControlPlanes) Len() int {
	return len(list)
}

func (list SortableControlPlanes) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func (list SortableControlPlanes) Less(i, j int) bool {
	return list[i].GetName() > list[j].GetName()
}

func (r *IstioControlPlaneReconciler) getCACertificatesFromPeers(ctx context.Context, icp *servicemeshv1alpha1.IstioControlPlane) ([]string, error) {
	certData := make([]string, 0)

	cps := make(SortableControlPlanes, 0)

	picpList := &servicemeshv1alpha1.PeerIstioControlPlaneList{}
	err := r.GetClient().List(ctx, picpList, client.InNamespace(icp.GetNamespace()))
	if err != nil {
		return nil, errors.WithStackIf(err)
	}

	for _, picp := range picpList.Items {
		picp := picp
		if picp.GetStatus().IstioControlPlaneName == icp.GetName() && picp.Spec.Mode == servicemeshv1alpha1.ModeType_ACTIVE {
			cps = append(cps, &picp)
		}
	}

	sort.Sort(cps)

	for _, cp := range cps {
		if cp.GetStatus().CaRootCertificate != "" {
			certData = append(certData, cp.GetStatus().CaRootCertificate)
		}
	}

	return certData, nil
}

func (r *IstioControlPlaneReconciler) getMeshNetworks(ctx context.Context, icp *servicemeshv1alpha1.IstioControlPlane) (*v1alpha1.MeshNetworks, error) {
	networks := make(map[string]*v1alpha1.Network)

	cps := make(SortableControlPlanes, 0)
	cps = append(cps, icp)

	picpList := &servicemeshv1alpha1.PeerIstioControlPlaneList{}
	err := r.GetClient().List(ctx, picpList, client.InNamespace(icp.GetNamespace()))
	if err != nil {
		return nil, errors.WithStackIf(err)
	}

	for _, picp := range picpList.Items {
		picp := picp
		if picp.GetStatus().IstioControlPlaneName == icp.GetName() {
			cps = append(cps, &picp)
		}
	}

	sort.Sort(cps)

	for _, cp := range cps {
		gateways := make([]*v1alpha1.Network_IstioNetworkGateway, 0)
		sort.Strings(cp.GetStatus().GatewayAddress)
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
				FromRegistry: cp.GetStatus().ClusterID,
			},
		})
		networks[networkName].Gateways = append(networks[networkName].Gateways, gateways...)
	}

	return &v1alpha1.MeshNetworks{
		Networks: networks,
	}, nil
}

func (r *IstioControlPlaneReconciler) getRelatedIstioMesh(ctx context.Context, c client.Client, icp *servicemeshv1alpha1.IstioControlPlane, logger logger.Logger) (*servicemeshv1alpha1.IstioMesh, error) {
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

	istiodEndpointAddresses, err := k8sutil.GetIstiodEndpointAddresses(ctx, r.Client, icp.GetName(), icp.GetSpec().GetNetworkName(), serviceNamespace)
	if err != nil {
		return errors.WithStackIf(err)
	}
	if len(istiodEndpointAddresses) == 0 {
		return errors.New("no valid istiod address found")
	}

	istiodEndpointPorts, err := k8sutil.GetIstiodEndpointPorts(ctx, r.Client, serviceName, serviceNamespace)
	if err != nil {
		return errors.WithStackIf(err)
	}

	endpoints := k8sutil.CreateK8sEndpoints(serviceName, serviceNamespace, istiodEndpointAddresses, istiodEndpointPorts)
	k8sutil.SetICPMetadataOnObject(endpoints, icp)

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

	if icp.DeletionTimestamp.IsZero() {
		state = reconciler.StatePresent
		secret, err = k8sutil.GetReaderSecretForCluster(
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
		k8sutil.SetICPMetadataOnObject(secret, icp)
	}

	_, err = r.ResourceReconciler.ReconcileResource(secret, state)
	if err != nil {
		return errors.WithStackIf(err)
	}

	return errors.WithStackIf(err)
}

func (r *IstioControlPlaneReconciler) removeFinalizerFromRelatedMeshGateways(ctx context.Context, icp *servicemeshv1alpha1.IstioControlPlane) error {
	r.Log.Info("remove finalizers from related meshgateways")
	var imgws servicemeshv1alpha1.IstioMeshGatewayList
	err := r.Client.List(ctx, &imgws)
	if err != nil {
		return errors.WrapIf(err, "could not list istio mesh gateway resources")
	}

	for _, imgw := range imgws.Items {
		imgw := imgw
		if imgw.GetSpec().GetIstioControlPlane().Name != icp.Name || imgw.GetSpec().GetIstioControlPlane().Namespace != icp.Namespace {
			continue
		}
		err = util.RemoveFinalizer(ctx, r.Client, &imgw, istioMeshGatewayFinalizerID, false)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (r *IstioControlPlaneReconciler) waitForMeshExpansionGatewayRemoval(ctx context.Context, icp *servicemeshv1alpha1.IstioControlPlane) error {
	getMeshExpansionEnabled := icp.GetSpec().GetMeshExpansion().GetEnabled().GetValue()
	if icp.DeletionTimestamp.IsZero() || !utils.PointerToBool(&getMeshExpansionEnabled) {
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
		icp.GetStatus().IstiodAddresses = nil

		return nil
	}

	endpoints, err := k8sutil.GetEndpoints(ctx, r.Client, icp.WithRevision("istiod"), icp.GetNamespace())
	if err != nil {
		return errors.WithStackIf(err)
	}
	icp.GetStatus().IstiodAddresses = k8sutil.GetIPsForEndpoints(endpoints)

	return nil
}

func (r *IstioControlPlaneReconciler) setMeshExpansionGWAddressToStatus(ctx context.Context, icp *servicemeshv1alpha1.IstioControlPlane) error {
	getMeshExpansionEnabled := icp.GetSpec().GetMeshExpansion().GetEnabled().GetValue()
	if icp.DeletionTimestamp.IsZero() && !utils.PointerToBool(&getMeshExpansionEnabled) {
		icp.GetStatus().GatewayAddress = nil

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

	icp.GetStatus().GatewayAddress = imgw.GetStatus().GatewayAddress

	return nil
}

func (r *IstioControlPlaneReconciler) setControlPlaneNameToStatus(icp *servicemeshv1alpha1.IstioControlPlane) {
	icp.GetStatus().IstioControlPlaneName = icp.GetName()
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

	icp.GetStatus().MeshConfig = &mc

	cs := icp.GetStatus().GetChecksums()
	if cs == nil {
		cs = &servicemeshv1alpha1.StatusChecksums{}
	}
	cs.MeshConfig = fmt.Sprintf("%x", sha256.Sum256([]byte(mcYAML)))
	icp.GetStatus().Checksums = cs

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

		cs := icp.GetStatus().GetChecksums()
		if cs == nil {
			cs = &servicemeshv1alpha1.StatusChecksums{}
		}
		cs.SidecarInjector = fmt.Sprintf("%x", sha256.Sum256(jm))
		icp.GetStatus().Checksums = cs
	}

	return nil
}

func (r *IstioControlPlaneReconciler) setInjectionNamespacesToStatus(ctx context.Context, icp *servicemeshv1alpha1.IstioControlPlane) error {
	namespaces := &corev1.NamespaceList{}
	err := r.GetClient().List(ctx, namespaces, client.MatchingLabels(icp.RevisionLabels()))
	if err != nil {
		return errors.WrapIf(err, "could not list namespaces")
	}

	names := []string{}
	for _, ns := range namespaces.Items {
		names = append(names, ns.GetName())
	}

	sort.Strings(names)

	icp.GetStatus().InjectionNamespaces = names

	return nil
}

// deleteIstioRootCAConfigmapsOnPassive deletes the istio-ca-root-cert-<revision> configmaps from passive clusters
// to gracefully handle the ACTIVE --> PASSIVE ICP mode switch and letting the cluster registry controller to recreate
// the configmap for the passive cluster
// NOTE: if cluster registry controller is not used, these configmaps need to be recreated manually
func (r *IstioControlPlaneReconciler) deleteIstioRootCAConfigmapsOnPassive(ctx context.Context, icp *servicemeshv1alpha1.IstioControlPlane, logger logger.Logger) error {
	if icp.GetSpec().GetMode() == servicemeshv1alpha1.ModeType_ACTIVE {
		return nil
	}

	configmaps := &corev1.ConfigMapList{}
	selectors := labels.NewSelector()

	crOwnershipFilter, err := labels.NewRequirement(clusterregistryv1alpha1.OwnershipAnnotation, selection.DoesNotExist, nil)
	if err != nil {
		return errors.WithStackIf(err)
	}
	selectors = selectors.Add(*crOwnershipFilter)

	istioConfigFilter, err := labels.NewRequirement("istio.io/config", selection.Equals, []string{"true"})
	if err != nil {
		return errors.WithStackIf(err)
	}
	selectors = selectors.Add(*istioConfigFilter)

	for key, value := range icp.RevisionLabels() {
		revisionLabelFilter, err := labels.NewRequirement(key, selection.Equals, []string{value})
		if err != nil {
			return errors.WithStackIf(err)
		}
		selectors = selectors.Add(*revisionLabelFilter)
	}

	err = r.GetClient().List(ctx, configmaps, client.MatchingLabelsSelector{
		Selector: selectors,
	})
	if err != nil {
		return errors.WrapIf(err, "could not list root ca configmaps")
	}

	for _, cm := range configmaps.Items {
		cm := cm
		annotations := cm.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		annotations[clusterregistryv1alpha1.OwnershipAnnotation] = "set-to-trigger-resync"
		cm.SetAnnotations(annotations)
		err = r.GetClient().Update(ctx, &cm)
		if err != nil {
			return errors.WrapIf(err, "could not update root ca configmap")
		}

		err = r.GetClient().Delete(ctx, &cm, client.PropagationPolicy(metav1.DeletePropagationForeground))
		if err != nil {
			return errors.WrapIf(err, "could not delete root ca configmap")
		}

		logger.Info(fmt.Sprintf("deleted root ca configmap %s.%s", cm.GetName(), cm.GetNamespace()))
	}

	return nil
}

func (r *IstioControlPlaneReconciler) setIstioCARootCertToStatus(ctx context.Context, icp *servicemeshv1alpha1.IstioControlPlane) error {
	if icp.GetSpec().GetMode() != servicemeshv1alpha1.ModeType_ACTIVE {
		icp.GetStatus().CaRootCertificate = ""

		return nil
	}

	secret := &corev1.Secret{}
	err := r.GetClient().Get(ctx, client.ObjectKey{
		Name:      "istio-ca-secret",
		Namespace: icp.GetNamespace(),
	}, secret)
	if k8serrors.IsNotFound(err) {
		icp.GetStatus().CaRootCertificate = ""

		return nil
	}
	if err != nil {
		return err
	}

	icp.GetStatus().CaRootCertificate = string(secret.Data["ca-cert.pem"])

	return nil
}

func (r *IstioControlPlaneReconciler) getICPReconcileRequests(ctx context.Context) ([]reconcile.Request, error) {
	icps := &servicemeshv1alpha1.IstioControlPlaneList{}
	err := r.Client.List(ctx, icps)
	if err != nil {
		return nil, errors.WrapIf(err, "could not list Istio control plane resources")
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

	return resources, nil
}

func (r *IstioControlPlaneReconciler) getNamespaceInjectionSourcePICP(ctx context.Context, cp client.ObjectKey) (*servicemeshv1alpha1.PeerIstioControlPlane, error) {
	picpList := &servicemeshv1alpha1.PeerIstioControlPlaneList{}
	err := r.GetClient().List(ctx, picpList, client.InNamespace(cp.Namespace))
	if err != nil {
		return nil, errors.WithStackIf(err)
	}

	var sourceICP *servicemeshv1alpha1.PeerIstioControlPlane
	for _, picp := range picpList.Items {
		picp := picp
		if v, ok := picp.GetAnnotations()[servicemeshv1alpha1.NamespaceInjectionSourceAnnotation]; ok && v == "true" && picp.GetStatus().IstioControlPlaneName == cp.Name { // nolint:goconst
			sourceICP = &picp
		}
	}

	return sourceICP, nil
}

func (r *IstioControlPlaneReconciler) reconcileNamespaceInjectionLabels(ctx context.Context, icp *servicemeshv1alpha1.IstioControlPlane) error {
	if a, ok := icp.GetAnnotations()[servicemeshv1alpha1.NamespaceInjectionSourceAnnotation]; ok && a == "true" {
		return nil
	}

	sourceICP, err := r.getNamespaceInjectionSourcePICP(ctx, client.ObjectKeyFromObject(icp))
	if err != nil {
		return err
	}

	if sourceICP == nil {
		return nil
	}

	r.Log.Info("sync namespace injection labels")

	namespaces := make(map[string]struct{})
	for _, name := range sourceICP.GetStatus().InjectionNamespaces {
		namespaces[name] = struct{}{}
	}

	localNamespacesWithInjectionLabel := make(map[string]struct{})
	localNamespaces := &corev1.NamespaceList{}
	err = r.GetClient().List(ctx, localNamespaces, client.MatchingLabels(icp.RevisionLabels()))
	if err != nil {
		return errors.WrapIf(err, "could not list namespaces")
	}
	for _, ns := range localNamespaces.Items {
		ns := ns
		if _, ok := namespaces[ns.GetName()]; !ok {
			labels := ns.GetLabels()
			delete(labels, servicemeshv1alpha1.RevisionedAutoInjectionLabel)
			ns.SetLabels(labels)
			r.Log.Info("remove injection label from namespace", "namespace", ns.GetName(), "label", servicemeshv1alpha1.RevisionedAutoInjectionLabel)
			err = r.GetClient().Update(ctx, &ns)
			if err != nil {
				errMsg := "could not remove injection label from namespace"
				r.Recorder.Event(
					&ns,
					corev1.EventTypeWarning,
					"IstioInjectionLabelRemovalError",
					errMsg,
				)

				return errors.WrapIfWithDetails(err, errMsg, "namespace", ns.GetName())
			}
			r.Recorder.Eventf(
				&ns,
				corev1.EventTypeNormal,
				"IstioInjectionLabelRemoval",
				"%s label removed from namespace %s, because the namespace either "+
					"does not exist or does not have %s label in the cluster %s, where the ICP %s "+
					"is present with the %s annotation",
				servicemeshv1alpha1.RevisionedAutoInjectionLabel,
				ns.GetName(),
				servicemeshv1alpha1.RevisionedAutoInjectionLabel,
				sourceICP.GetSpec().GetClusterID(),
				icp.GetName(),
				servicemeshv1alpha1.NamespaceInjectionSourceAnnotation,
			)
		} else {
			localNamespacesWithInjectionLabel[ns.GetName()] = struct{}{}
		}
	}

	for name := range namespaces {
		if _, ok := localNamespacesWithInjectionLabel[name]; ok {
			continue
		}
		ns := &corev1.Namespace{}
		err = r.GetClient().Get(ctx, client.ObjectKey{
			Name: name,
		}, ns)
		if k8serrors.IsNotFound(err) {
			continue
		}
		if err != nil {
			return errors.WrapIfWithDetails(err, "could not get namespace", "namespace", name)
		}
		labels := utils.MergeLabels(ns.GetLabels(), icp.RevisionLabels())
		delete(labels, servicemeshv1alpha1.DeprecatedAutoInjectionLabel)
		ns.SetLabels(labels)
		r.Log.Info("add injection label to namespace", "namespace", ns.GetName(), "label", servicemeshv1alpha1.RevisionedAutoInjectionLabel)
		err = r.GetClient().Update(ctx, ns)
		if err != nil {
			errMsg := "could not update namespace"
			r.Recorder.Event(
				ns,
				corev1.EventTypeWarning,
				"IstioInjectionLabelAdditionError",
				errMsg,
			)

			return errors.WrapIfWithDetails(err, errMsg, "namespace", name)
		}
		r.Recorder.Eventf(
			ns,
			corev1.EventTypeNormal,
			"IstioInjectionLabelAddition",
			"%s label added to namespace %s, because the namespace "+
				"has %s label in the cluster %s, where the ICP %s is present "+
				"with the %s annotation",
			servicemeshv1alpha1.RevisionedAutoInjectionLabel,
			name,
			servicemeshv1alpha1.RevisionedAutoInjectionLabel,
			sourceICP.GetSpec().GetClusterID(),
			icp.GetName(),
			servicemeshv1alpha1.NamespaceInjectionSourceAnnotation,
		)
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
		err := r.ctrl.Watch(&source.Kind{Type: t}, eventHandler, util.ObjectChangePredicate{Logger: r.Log})
		if err != nil {
			return err
		}
	}

	return nil
}

func RemoveFinalizers(ctx context.Context, c client.Client) error {
	var icps servicemeshv1alpha1.IstioControlPlaneList
	err := c.List(context.Background(), &icps)
	if err != nil {
		return errors.WrapIf(err, "could not list Istio control plane resources")
	}

	for _, istio := range icps.Items {
		istio := istio
		err = util.RemoveFinalizer(ctx, c, &istio, istioControlPlaneFinalizerID, false)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	var imgws servicemeshv1alpha1.IstioMeshGatewayList
	err = c.List(context.Background(), &imgws)
	if err != nil {
		return errors.WrapIf(err, "could not list istio mesh gateway resources")
	}

	for _, imgw := range imgws.Items {
		imgw := imgw
		err = util.RemoveFinalizer(ctx, c, &imgw, istioMeshGatewayFinalizerID, false)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

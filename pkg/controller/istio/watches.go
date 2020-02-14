/*
Copyright 2020 Banzai Cloud.

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
	"github.com/pkg/errors"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	v1alpha1 "github.com/banzaicloud/istio-client-go/pkg/authentication/v1alpha1"
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/crds"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
)

func (r *ReconcileIstio) initWatches(watchCreatedResourcesEvents bool) error {
	if r.ctrl == nil {
		return errors.New("controller is not set")
	}

	err := r.watchIstioConfig()
	if err != nil {
		return err
	}

	err = r.watchRemoteIstioConfig()
	if err != nil {
		return err
	}

	err = r.watchIstioCoreDNSService()
	if err != nil {
		return err
	}

	if !watchCreatedResourcesEvents {
		return nil
	}

	createdResourceTypes := []runtime.Object{
		&corev1.ServiceAccount{TypeMeta: metav1.TypeMeta{Kind: "ServiceAccount", APIVersion: corev1.SchemeGroupVersion.String()}},
		&rbacv1.Role{TypeMeta: metav1.TypeMeta{Kind: "ClusterRole", APIVersion: rbacv1.SchemeGroupVersion.String()}},
		&rbacv1.RoleBinding{TypeMeta: metav1.TypeMeta{Kind: "ClusterRoleBinding", APIVersion: rbacv1.SchemeGroupVersion.String()}},
		&rbacv1.ClusterRole{TypeMeta: metav1.TypeMeta{Kind: "ClusterRole", APIVersion: rbacv1.SchemeGroupVersion.String()}},
		&rbacv1.ClusterRoleBinding{TypeMeta: metav1.TypeMeta{Kind: "ClusterRoleBinding", APIVersion: rbacv1.SchemeGroupVersion.String()}},
		&corev1.ConfigMap{TypeMeta: metav1.TypeMeta{Kind: "ConfigMap", APIVersion: corev1.SchemeGroupVersion.String()}},
		&corev1.Service{TypeMeta: metav1.TypeMeta{Kind: "Service", APIVersion: corev1.SchemeGroupVersion.String()}},
		&appsv1.Deployment{TypeMeta: metav1.TypeMeta{Kind: "Deployment", APIVersion: appsv1.SchemeGroupVersion.String()}},
		&appsv1.DaemonSet{TypeMeta: metav1.TypeMeta{Kind: "DaemonSet", APIVersion: appsv1.SchemeGroupVersion.String()}},
		&autoscalingv2beta1.HorizontalPodAutoscaler{TypeMeta: metav1.TypeMeta{Kind: "HorizontalPodAutoscaler", APIVersion: autoscalingv2beta1.SchemeGroupVersion.String()}},
		&admissionregistrationv1beta1.MutatingWebhookConfiguration{TypeMeta: metav1.TypeMeta{Kind: "MutatingWebhookConfiguration", APIVersion: admissionregistrationv1beta1.SchemeGroupVersion.String()}},
		&corev1.Namespace{TypeMeta: metav1.TypeMeta{Kind: "Namespace", APIVersion: corev1.SchemeGroupVersion.String()}},
		&istiov1beta1.MeshGateway{TypeMeta: metav1.TypeMeta{Kind: "MeshGateway", APIVersion: istiov1beta1.SchemeGroupVersion.String()}},
	}

	// Watch for changes to resources managed by the operator
	for _, resource := range createdResourceTypes {
		err = r.watchResource(resource)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ReconcileIstio) watchIstioConfig() error {
	err := r.ctrl.Watch(
		&source.Kind{
			Type: &istiov1beta1.Istio{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Istio",
					APIVersion: istiov1beta1.SchemeGroupVersion.String(),
				},
			},
		},
		&handler.EnqueueRequestForObject{},
		k8sutil.GetWatchPredicateForIstio(),
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *ReconcileIstio) watchRemoteIstioConfig() error {
	err := r.ctrl.Watch(
		&source.Kind{
			Type: &istiov1beta1.RemoteIstio{
				TypeMeta: metav1.TypeMeta{
					Kind:       "RemoteIstio",
					APIVersion: istiov1beta1.SchemeGroupVersion.String(),
				},
			},
		},
		&handler.EnqueueRequestsFromMapFunc{
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
		},
		k8sutil.GetWatchPredicateForRemoteIstioAvailability(),
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *ReconcileIstio) watchIstioCoreDNSService() error {
	err := r.ctrl.Watch(
		&source.Kind{
			Type: &corev1.Service{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Service",
					APIVersion: corev1.SchemeGroupVersion.String(),
				},
			},
		},
		&handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &istiov1beta1.Istio{},
		},
		k8sutil.GetWatchPredicateForIstioService("istiocoredns"),
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *ReconcileIstio) watchResource(resource runtime.Object) error {
	err := r.ctrl.Watch(
		&source.Kind{
			Type: resource,
		},
		&handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &istiov1beta1.Istio{},
		},
		k8sutil.GetWatchPredicateForOwnedResources(&istiov1beta1.Istio{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Istio",
				APIVersion: istiov1beta1.SchemeGroupVersion.String(),
			},
		}, true, r.mgr.GetScheme(), log),
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *ReconcileIstio) watchCRDs(nn types.NamespacedName) error {
	err := r.ctrl.Watch(
		&source.Kind{
			Type: &extensionsobj.CustomResourceDefinition{
				TypeMeta: metav1.TypeMeta{
					Kind:       "CustomResourceDefinition",
					APIVersion: extensionsobj.SchemeGroupVersion.String(),
				},
			},
		},
		&handler.EnqueueRequestsFromMapFunc{
			ToRequests: handler.ToRequestsFunc(func(object handler.MapObject) []reconcile.Request {
				return []reconcile.Request{
					{
						NamespacedName: nn,
					},
				}
			}),
		},
		crds.GetWatchPredicateForCRDs(),
	)

	return err
}

func (r *ReconcileIstio) watchMeshPolicy(nn types.NamespacedName) error {
	err := r.ctrl.Watch(
		&source.Kind{
			Type: &v1alpha1.MeshPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       "MeshPolicy",
					APIVersion: v1alpha1.SchemeGroupVersion.String(),
				},
			},
		},
		&handler.EnqueueRequestsFromMapFunc{
			ToRequests: handler.ToRequestsFunc(func(object handler.MapObject) []reconcile.Request {
				return []reconcile.Request{
					{
						NamespacedName: nn,
					},
				}
			}),
		},
		k8sutil.GetWatchPredicateForOwnedResources(&istiov1beta1.Istio{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Istio",
				APIVersion: istiov1beta1.SchemeGroupVersion.String(),
			},
		}, true, r.mgr.GetScheme(), log),
	)

	return err
}

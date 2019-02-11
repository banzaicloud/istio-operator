package k8sutil

import (
	"context"
	"errors"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"istio.io/api/pkg/kube/apis/config/v1alpha2"
	"istio.io/api/pkg/kube/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/autoscaling/v2beta1"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

type UnstructuredResource struct {
	Name      string
	Namespace string
	Gvr       schema.GroupVersionResource
	Resource  *unstructured.Unstructured
}

func ReconcileResource(log logr.Logger, client runtimeClient.Client, namespace string, name string, desired runtime.Object) error {
	log = log.WithValues("type", reflect.TypeOf(desired))
	var current = desired.DeepCopyObject()
	key, err := runtimeClient.ObjectKeyFromObject(current)
	if err != nil {
		return emperror.With(err)
	}
	err = client.Get(context.TODO(), key, current)
	if err != nil && !apierrors.IsNotFound(err) {
		return emperror.WrapWith(err, "getting resource failed", "name", name, "type", reflect.TypeOf(desired))
	}
	if apierrors.IsNotFound(err) {
		if err := client.Create(context.TODO(), desired); err != nil {
			return emperror.WrapWith(err, "creating resource failed", "name", name, "type", reflect.TypeOf(desired))
		}
		log.Info("resource created", "name", name)
	}
	if err == nil {
		switch desired.(type) {
		default:
			return emperror.With(errors.New("unexpected resource type"), "type", reflect.TypeOf(desired))
		case *apiv1.Namespace:
			ns := desired.(*apiv1.Namespace)
			ns.ResourceVersion = current.(*apiv1.Namespace).ResourceVersion
			desired = ns
		case *apiv1.ServiceAccount:
			sa := desired.(*apiv1.ServiceAccount)
			sa.ResourceVersion = current.(*apiv1.ServiceAccount).ResourceVersion
			desired = sa
		case *rbacv1.ClusterRole:
			cr := desired.(*rbacv1.ClusterRole)
			cr.ResourceVersion = current.(*rbacv1.ClusterRole).ResourceVersion
			desired = cr
		case *rbacv1.ClusterRoleBinding:
			crb := desired.(*rbacv1.ClusterRoleBinding)
			crb.ResourceVersion = current.(*rbacv1.ClusterRoleBinding).ResourceVersion
			desired = crb
		case *apiv1.ConfigMap:
			cm := desired.(*apiv1.ConfigMap)
			cm.ResourceVersion = current.(*apiv1.ConfigMap).ResourceVersion
			desired = cm
		case *apiv1.Service:
			svc := desired.(*apiv1.Service)
			svc.ResourceVersion = current.(*apiv1.Service).ResourceVersion
			svc.Spec.ClusterIP = current.(*apiv1.Service).Spec.ClusterIP
			desired = svc
		case *appsv1.Deployment:
			deploy := desired.(*appsv1.Deployment)
			deploy.ResourceVersion = current.(*appsv1.Deployment).ResourceVersion
			desired = deploy
		case *v2beta1.HorizontalPodAutoscaler:
			hpa := desired.(*v2beta1.HorizontalPodAutoscaler)
			hpa.ResourceVersion = current.(*v2beta1.HorizontalPodAutoscaler).ResourceVersion
			desired = hpa
		case *v1alpha3.Gateway:
			gw := desired.(*v1alpha3.Gateway)
			gw.ResourceVersion = current.(*v1alpha3.Gateway).ResourceVersion
			desired = gw
		case *v1alpha2.AttributeManifest:
			am := desired.(*v1alpha2.AttributeManifest)
			am.ResourceVersion = current.(*v1alpha2.AttributeManifest).ResourceVersion
			desired = am
		}
		if err := client.Update(context.TODO(), desired); err != nil {
			return emperror.WrapWith(err, "updating resource failed", "name", name, "type", reflect.TypeOf(desired))
		}
		log.Info("resource updated", "name", name)
	}
	return nil
}

func ReconcileDynamicResource(log logr.Logger, client dynamic.Interface, desired UnstructuredResource) error {
	log = log.WithValues("type", reflect.TypeOf(desired))
	_, err := client.Resource(desired.Gvr).Namespace(desired.Namespace).Get(desired.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return emperror.WrapWith(err, "getting resource failed", "name", desired.Name, "type", reflect.TypeOf(desired))
	}
	if apierrors.IsNotFound(err) {
		if _, err := client.Resource(desired.Gvr).Namespace(desired.Namespace).Create(desired.Resource, metav1.CreateOptions{}); err != nil {
			return emperror.WrapWith(err, "creating resource failed", "name", desired.Name, "type", reflect.TypeOf(desired))
		}
		log.Info("resource created", "name", desired.Name)
	}
	if err == nil {
		if _, err := client.Resource(desired.Gvr).Update(desired.Resource, metav1.UpdateOptions{}); err != nil {
			return emperror.WrapWith(err, "updating resource failed", "name", desired.Name, "type", reflect.TypeOf(desired))
		}
		log.Info("resource updated", "name", desired.Name)
	}
	return nil
}

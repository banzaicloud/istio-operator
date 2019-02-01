package k8sutil

import (
	"context"
	"errors"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

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
			desired = svc
		case *appsv1.Deployment:
			deploy := desired.(*appsv1.Deployment)
			deploy.ResourceVersion = current.(*appsv1.Deployment).ResourceVersion
			desired = deploy
		}
		if err := client.Update(context.TODO(), desired); err != nil {
			return emperror.WrapWith(err, "updating resource failed", "name", name, "type", reflect.TypeOf(desired))
		}
		log.Info("resource updated", "name", name)
	}
	return nil
}

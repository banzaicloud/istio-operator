package k8sutil

import (
	"github.com/goph/emperror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	apiv1 "k8s.io/api/core/v1"
	"github.com/go-logr/logr"
	"context"
	"k8s.io/apimachinery/pkg/types"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"errors"
	"reflect"
)

func ReconcileResource(log logr.Logger, client client.Client, namespace string, name string, desired runtime.Object) error {
	log = log.WithValues("type", reflect.TypeOf(desired))
	var current = desired.DeepCopyObject()
	err := client.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, current)
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
		}
		if err := client.Update(context.TODO(), desired); err != nil {
			return emperror.WrapWith(err, "updating resource failed", "name", name, "type", reflect.TypeOf(desired))
		}
		log.Info("resource updated", "name", name)
	}
	return nil
}

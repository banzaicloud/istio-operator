package k8sutil

import (
	"github.com/goph/emperror"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	"context"
	"k8s.io/apimachinery/pkg/types"
)

func ReconcileServiceAccount(log logr.Logger, client client.Client, desired *apiv1.ServiceAccount) error {
	var current *apiv1.ServiceAccount
	err := client.Get(context.TODO(), types.NamespacedName{desired.Namespace, desired.Name}, current)
	if err != nil && !apierrors.IsNotFound(err) {
		return emperror.WrapWith(err, "getting service account failed", "name", desired.Name)
	}
	if apierrors.IsNotFound(err) {
		if err := client.Create(context.TODO(), desired); err != nil {
			return emperror.WrapWith(err, "creating service account failed", "name", desired.Name)
		}
		log.Info("service account created", "name", desired.Name)
	}
	if err == nil {
		desired.ResourceVersion = current.ResourceVersion
		if err := client.Update(context.TODO(), desired); err != nil {
			return emperror.WrapWith(err, "updating service account failed", "name", desired.Name)
		}
		log.Info("service account updated", "name", desired.Name)
	}
	return nil
}

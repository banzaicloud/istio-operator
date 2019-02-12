package config

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
)

type ComponentReconciler interface {
	Reconcile(log logr.Logger) error
}

type Resource func(owner *istiov1beta1.Config) runtime.Object

type Reconciler struct {
	client.Client
	Owner     *istiov1beta1.Config
	Resources []Resource
}

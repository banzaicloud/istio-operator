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

package config

import (
	"context"
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller")

// Add creates a new Config Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileConfig{
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		config: mgr.GetConfig(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("config-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Config
	err = c.Watch(&source.Kind{Type: &istiov1beta1.Config{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	//
	//// TODO(user): Modify this to be the types you create
	//// Uncomment watch a Deployment created by Config - change this for objects you create
	//err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
	//	IsController: true,
	//	OwnerType:    &istiov1beta1.Config{},
	//})
	//if err != nil {
	//	return err
	//}

	return nil
}

var _ reconcile.Reconciler = &ReconcileConfig{}

// ReconcileConfig reconciles a Config object
type ReconcileConfig struct {
	client.Client
	scheme *runtime.Scheme
	config *rest.Config
}

type ReconcileComponent func(log logr.Logger, istio *istiov1beta1.Config) error

// Reconcile reads that state of the cluster for a Config object and makes changes based on the state read
// and what is in the Config.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=istio.banzaicloud.io,resources=configs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=istio.banzaicloud.io,resources=configs/status,verbs=get;update;patch
func (r *ReconcileConfig) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Istio")
	// Fetch the Config instance
	instance := &istiov1beta1.Config{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	reconcilers := []ReconcileComponent{
		r.ReconcileCitadel,
		r.ReconcileGalley,
		r.ReconcilePilot,
		r.ReconcileMixer,
		r.ReconcileGateways,
		r.ReconcileSidecarInjector,
	}

	for _, rec := range reconcilers {
		err = rec(reqLogger, instance)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

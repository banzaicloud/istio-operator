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

package endpoints

import (
	"context"

	"github.com/banzaicloud/istio-operator/pkg/remoteclusters"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("endpoints")

// Add creates a new controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, cm *remoteclusters.Manager) error {
	return add(mgr, newReconciler(mgr, cm))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, cm *remoteclusters.Manager) reconcile.Reconciler {
	return &ReconcileEndpoints{
		Client:            mgr.GetClient(),
		scheme:            mgr.GetScheme(),
		remoteClustersMgr: cm,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("endpoints-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Pod
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForObject{}, getWatchPredicate())
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileEndpoints{}

// ReconcileEndpoints reconciles a Pod object
type ReconcileEndpoints struct {
	client.Client
	scheme *runtime.Scheme

	remoteClustersMgr *remoteclusters.Manager
}

// Reconcile reads that state of the cluster for a Pod object and makes changes based on the state read
// and what is in the Pod.Spec
// +kubebuilder:rbac:groups=apps,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=pods/status,verbs=get;update;patch
func (r *ReconcileEndpoints) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Pod instance
	pod := &corev1.Pod{}
	err := r.Get(context.TODO(), request.NamespacedName, pod)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	ready := 0
	for _, status := range pod.Status.ContainerStatuses {
		if status.Ready {
			ready++
		}
	}

	if pod.Status.Phase != corev1.PodRunning || len(pod.Spec.Containers) != ready {
		return reconcile.Result{}, nil
	}

	log.Info("pod event detected", "podName", pod.Name, "podIP", pod.Status.PodIP, "podStatus", pod.Status.Phase)

	for _, cluster := range r.remoteClustersMgr.GetAll() {
		remoteConfig := cluster.GetRemoteConfig()
		if remoteConfig != nil {
			for i, svc := range remoteConfig.Spec.EnabledServices {
				ls, err := labels.Parse(svc.LabelSelector)
				if err != nil {
					return reconcile.Result{}, err
				}
				if ls.Matches(labels.Set(pod.Labels)) {
					svc.IPs = []string{pod.Status.PodIP}
					remoteConfig.Spec.EnabledServices[i] = svc
				}
			}

			log.Info("updating endpoints", "cluster", cluster.GetName())
			err = cluster.ReconcileEnabledServiceEndpoints(remoteConfig)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	}

	return reconcile.Result{}, nil
}

func getWatchPredicate() predicate.Funcs {
	return predicate.Funcs{
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
		CreateFunc: func(e event.CreateEvent) bool {
			if _, ok := e.Meta.GetLabels()["istio"]; ok {
				return true
			}
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if _, ok := e.MetaNew.GetLabels()["istio"]; ok {
				return true
			}
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}
}

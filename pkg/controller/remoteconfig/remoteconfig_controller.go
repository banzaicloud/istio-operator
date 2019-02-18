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

package remoteconfig

import (
	"context"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	operatorv1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/remoteclusters"
	pkgerrors "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("remote-istio-controller")

// Add creates a new RemoteConfig Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, cm *remoteclusters.Manager) error {
	return add(mgr, newReconciler(mgr, cm))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, cm *remoteclusters.Manager) reconcile.Reconciler {
	return &ReconcileRemoteConfig{
		Client:            mgr.GetClient(),
		scheme:            mgr.GetScheme(),
		remoteClustersMgr: cm,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("remoteconfig-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to RemoteConfig
	err = c.Watch(&source.Kind{Type: &operatorv1beta1.RemoteConfig{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileRemoteConfig{}

// ReconcileRemoteConfig reconciles a RemoteConfig object
type ReconcileRemoteConfig struct {
	client.Client
	scheme *runtime.Scheme

	remoteClustersMgr *remoteclusters.Manager
}

// Reconcile reads that state of the cluster for a RemoteConfig object and makes changes based on the state read
// and what is in the RemoteConfig.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=operator.istio.io,resources=remoteconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.istio.io,resources=remoteconfigs/status,verbs=get;update;patch
func (r *ReconcileRemoteConfig) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Info("reconciling remote istio deployment")

	remoteConfig := &operatorv1beta1.RemoteConfig{}
	err := r.Get(context.TODO(), request.NamespacedName, remoteConfig)
	if err != nil {
		if !errors.IsNotFound(err) {
			// Error reading the object - requeue the request.
			log.Info("remoteconfig error - requeue")
			return reconcile.Result{}, err
		}

		// Handle delete with a finalizer
		return reconcile.Result{}, nil
	}

	finalizer := "remote-istio-operator.finializer.banzaicloud.io"
	if remoteConfig.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(remoteConfig.ObjectMeta.Finalizers, finalizer) {
			remoteConfig.ObjectMeta.Finalizers = append(remoteConfig.ObjectMeta.Finalizers, finalizer)
			if err := r.Update(context.Background(), remoteConfig); err != nil {
				return reconcile.Result{}, err
			}
		}
	} else {
		if containsString(remoteConfig.ObjectMeta.Finalizers, finalizer) {
			log.Info("running finalizer stuff", "instance", remoteConfig)
			cluster, err := r.remoteClustersMgr.Get(remoteConfig.Spec.ClusterName)
			if err == nil {
				err = cluster.RemoveConfig()
				if err != nil {
					return reconcile.Result{}, err
				}
			}
		}

		remoteConfig.ObjectMeta.Finalizers = removeString(remoteConfig.ObjectMeta.Finalizers, finalizer)
		if err := r.Update(context.Background(), remoteConfig); err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	}

	remoteConfig, err = r.populateEnabledServicePodIPs(remoteConfig)
	if err != nil {
		return reconcile.Result{}, err
	}

	cluster, _ := r.remoteClustersMgr.Get(remoteConfig.Spec.ClusterName)
	if cluster == nil {
		cluster, err = r.getRemoteCluster(remoteConfig)
	}
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileRemoteCluster(remoteConfig, cluster)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileRemoteConfig) populateEnabledServicePodIPs(remoteConfig *istiov1beta1.RemoteConfig) (*istiov1beta1.RemoteConfig, error) {
	var pods corev1.PodList

	for i, svc := range remoteConfig.Spec.Config.EnabledServices {
		o := &client.ListOptions{}
		err := o.SetLabelSelector(svc.LabelSelector)
		if err != nil {
			return remoteConfig, err
		}
		err = r.List(context.TODO(), o, &pods)
		if err != nil {
			return remoteConfig, err
		}
		for _, pod := range pods.Items {
			ready := 0
			for _, status := range pod.Status.ContainerStatuses {
				if status.Ready {
					ready++
				}
			}
			if len(pod.Spec.Containers) == ready {
				svc.IPs = append(svc.IPs, pod.Status.PodIP)
				remoteConfig.Spec.Config.EnabledServices[i] = svc
			}
		}
	}

	return remoteConfig, nil
}

func (r *ReconcileRemoteConfig) getRemoteCluster(remoteConfig *istiov1beta1.RemoteConfig) (*remoteclusters.Cluster, error) {
	k8sconfig, err := r.getK8SConfigForCluster(remoteConfig.ObjectMeta.Namespace, remoteConfig.Spec.ClusterName)
	if err != nil {
		return nil, err
	}

	log.Info("k8s config found", "cluster", remoteConfig.Spec.ClusterName)

	cluster, err := remoteclusters.NewCluster(remoteConfig.Spec.ClusterName, k8sconfig, log)
	if err != nil {
		return nil, err
	}
	err = r.remoteClustersMgr.Add(cluster)
	if err != nil {
		return nil, err
	}

	return cluster, nil
}

func (r *ReconcileRemoteConfig) reconcileRemoteCluster(remoteConfig *istiov1beta1.RemoteConfig, cluster *remoteclusters.Cluster) error {
	err := cluster.Reconcile(remoteConfig)
	if err != nil {
		return err
	}

	return nil
}

func (r *ReconcileRemoteConfig) getK8SConfigForCluster(namespace string, name string) ([]byte, error) {
	var secret corev1.Secret
	err := r.Get(context.TODO(), client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, &secret)
	if err != nil {
		return nil, err
	}

	for _, config := range secret.Data {
		return config, nil
	}

	return nil, pkgerrors.New("could not found k8s config")
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

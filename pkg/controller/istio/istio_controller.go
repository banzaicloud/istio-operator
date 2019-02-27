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

package istio

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/crds"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil/objectmatch"
	"github.com/banzaicloud/istio-operator/pkg/resources"
	"github.com/banzaicloud/istio-operator/pkg/resources/citadel"
	"github.com/banzaicloud/istio-operator/pkg/resources/common"
	"github.com/banzaicloud/istio-operator/pkg/resources/galley"
	"github.com/banzaicloud/istio-operator/pkg/resources/gateways"
	"github.com/banzaicloud/istio-operator/pkg/resources/mixer"
	"github.com/banzaicloud/istio-operator/pkg/resources/pilot"
	"github.com/banzaicloud/istio-operator/pkg/resources/sidecarinjector"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

const finalizerID = "istio-operator.finializer.banzaicloud.io"

var log = logf.Log.WithName("controller")

// Add creates a new Config Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	dynamic, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		return emperror.Wrap(err, "failed to create dynamic client")
	}
	crd, err := crds.New(mgr.GetConfig(), crds.InitCrds())
	if err != nil {
		return emperror.Wrap(err, "unable to set up crd operator")
	}
	return add(mgr, newReconciler(mgr, dynamic, crd))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, d dynamic.Interface, crd *crds.CrdOperator) reconcile.Reconciler {
	return &ReconcileConfig{
		Client:      mgr.GetClient(),
		dynamic:     d,
		crdOperator: crd,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("config-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	err = initWatches(c)
	if err != nil {
		return err
	}
	return nil
}

var _ reconcile.Reconciler = &ReconcileConfig{}

// ReconcileConfig reconciles a Config object
type ReconcileConfig struct {
	client.Client
	dynamic     dynamic.Interface
	crdOperator *crds.CrdOperator
}

type ReconcileComponent func(log logr.Logger, istio *istiov1beta1.Istio) error

// Reconcile reads that state of the cluster for a Config object and makes changes based on the state read
// and what is in the Config.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=istio.banzaicloud.io,resources=istios,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=istio.banzaicloud.io,resources=istios/status,verbs=get;update;patch
func (r *ReconcileConfig) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger := log.WithValues("trigger", request.Namespace+"/"+request.Name)
	logger.Info("Reconciling Istio")
	// Fetch the Config instance
	config := &istiov1beta1.Istio{}
	err := r.Get(context.TODO(), request.NamespacedName, config)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	// Set default values where not set
	istiov1beta1.SetDefaults(config)
	result, err := r.reconcile(logger, config)
	if err != nil {
		updateErr := r.updateStatus(config, istiov1beta1.ReconcileFailed, err.Error())
		if updateErr != nil {
			log.Error(updateErr, "failed to update state")
			return result, errors.WithStack(err)
		}
		return reconcile.Result{}, emperror.Wrap(err, "could not reconcile istio")
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileConfig) reconcile(logger logr.Logger, config *istiov1beta1.Istio) (reconcile.Result, error) {

	if config.Status.Status == "" {
		err := r.updateStatus(config, istiov1beta1.Created, "")
		if err != nil {
			return reconcile.Result{}, errors.WithStack(err)
		}
	}

	// add finalizer strings and update
	if config.ObjectMeta.DeletionTimestamp.IsZero() {
		if !util.ContainsString(config.ObjectMeta.Finalizers, finalizerID) {
			config.ObjectMeta.Finalizers = append(config.ObjectMeta.Finalizers, finalizerID)
			if err := r.Update(context.Background(), config); err != nil {
				return reconcile.Result{}, emperror.Wrap(err, "could not add finalizer to config")
			}
			return reconcile.Result{}, nil
		}
	} else {
		// Deletion timestamp set, config is marked for deletion
		if util.ContainsString(config.ObjectMeta.Finalizers, finalizerID) {
			if config.Status.Status == istiov1beta1.Reconciling && config.Status.ErrorMessage == "" {
				log.Info("cannot remove Istio while reconciling")
				return reconcile.Result{}, nil
			}
			config.ObjectMeta.Finalizers = util.RemoveString(config.ObjectMeta.Finalizers, finalizerID)
			if err := r.Update(context.Background(), config); err != nil {
				return reconcile.Result{}, emperror.Wrap(err, "could not remove finalizer from config")
			}
		}

		log.Info("Istio removed")

		return reconcile.Result{}, nil
	}

	if config.Status.Status == istiov1beta1.Reconciling {
		return reconcile.Result{}, errors.New("cannot trigger reconcile while already reconciling")
	}

	err := r.updateStatus(config, istiov1beta1.Reconciling, "")
	if err != nil {
		return reconcile.Result{}, errors.WithStack(err)
	}

	log.Info("reconciling CRDs")
	err = r.crdOperator.Reconcile(config, logger)
	if err != nil {
		log.Error(err, "unable to reconcile CRDs")
		return reconcile.Result{}, err
	}

	reconcilers := []resources.ComponentReconciler{
		common.New(r.Client, config),
		citadel.New(citadel.Configuration{
			DeployMeshPolicy: true,
			SelfSignedCA:     true,
		}, r.Client, r.dynamic, config),
		galley.New(r.Client, config),
		pilot.New(r.Client, r.dynamic, config),
		gateways.New(r.Client, config),
		mixer.New(r.Client, r.dynamic, config),
		sidecarinjector.New(r.Client, config),
	}

	for _, rec := range reconcilers {
		err = rec.Reconcile(log)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	err = r.updateStatus(config, istiov1beta1.Available, "")
	if err != nil {
		return reconcile.Result{}, errors.WithStack(err)
	}
	log.Info("reconcile finished")
	return reconcile.Result{}, nil
}

func (r *ReconcileConfig) updateStatus(config *istiov1beta1.Istio, status istiov1beta1.ConfigState, errorMessage string) error {
	typeMeta := config.TypeMeta
	config.Status.Status = status
	config.Status.ErrorMessage = errorMessage
	err := r.Status().Update(context.Background(), config)
	if err != nil {
		if !k8serrors.IsConflict(err) {
			return emperror.Wrapf(err, "could not update Istio state to '%s'", status)
		}
		err := r.Get(context.TODO(), types.NamespacedName{
			Namespace: config.Namespace,
			Name:      config.Name,
		}, config)
		if err != nil {
			return emperror.Wrap(err, "could not get config for updating status")
		}
		config.Status.Status = status
		config.Status.ErrorMessage = errorMessage
		err = r.Status().Update(context.Background(), config)
		if err != nil {
			return emperror.Wrapf(err, "could not update Istio state to '%s'", status)
		}
	}
	// update loses the typeMeta of the config that's used later when setting ownerrefs
	config.TypeMeta = typeMeta
	log.Info("Istio state updated", "status", status)
	return nil
}

func watchPredicateForConfig() predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			old := e.ObjectOld.(*istiov1beta1.Istio)
			new := e.ObjectNew.(*istiov1beta1.Istio)
			if !reflect.DeepEqual(old.Spec, new.Spec) ||
				old.GetDeletionTimestamp() != new.GetDeletionTimestamp() ||
				old.GetGeneration() != new.GetGeneration() {
				return true
			}
			return false
		},
	}
}

func initWatches(c controller.Controller) error {
	// Watch for changes to Config
	err := c.Watch(&source.Kind{Type: &istiov1beta1.Istio{}}, &handler.EnqueueRequestForObject{}, watchPredicateForConfig())
	if err != nil {
		return err
	}

	// Watch for changes to resources managed by the operator
	for _, t := range []runtime.Object{
		&corev1.ServiceAccount{},
		&rbacv1.ClusterRole{},
		&rbacv1.ClusterRoleBinding{},
		&corev1.ConfigMap{},
		&corev1.Service{},
		&appsv1.Deployment{},
		&autoscalingv2beta1.HorizontalPodAutoscaler{},
		&admissionregistrationv1beta1.MutatingWebhookConfiguration{},
	} {
		err = c.Watch(&source.Kind{Type: t}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &istiov1beta1.Istio{},
		}, predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				return false
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return true
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				objectsEquals, err := objectmatch.Match(e.ObjectOld, e.ObjectNew)
				if err != nil {
					log.Error(err, "could not match objects", "kind", e.ObjectOld.GetObjectKind())
				} else if objectsEquals {
					return false
				}
				return true
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

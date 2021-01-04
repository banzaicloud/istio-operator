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

package meshgateway

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/go-logr/logr"
	"github.com/gofrs/uuid"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources/gateways"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

var log = logf.Log.WithName("controller")

// Add creates a new MeshGateway Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	dynamic, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		return emperror.Wrap(err, "failed to create dynamic client")
	}

	return add(mgr, newReconciler(mgr, dynamic))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, d dynamic.Interface) reconcile.Reconciler {
	return &ReconcileMeshGateway{Client: mgr.GetClient(), dynamic: d, scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("meshgateway-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to MeshGateway
	err = c.Watch(&source.Kind{Type: &istiov1beta1.MeshGateway{TypeMeta: metav1.TypeMeta{Kind: "MeshGateway", APIVersion: "istio.banzaicloud.io/v1beta1"}}}, &handler.EnqueueRequestForObject{}, k8sutil.GetWatchPredicateForMeshGateway())
	if err != nil {
		return err
	}

	// Watch for changes to resources created by the controller
	for _, t := range []runtime.Object{
		&corev1.ServiceAccount{TypeMeta: metav1.TypeMeta{Kind: "ServiceAccount", APIVersion: "v1"}},
		&rbacv1.Role{TypeMeta: metav1.TypeMeta{Kind: "Role", APIVersion: "v1"}},
		&rbacv1.RoleBinding{TypeMeta: metav1.TypeMeta{Kind: "RoleBinding", APIVersion: "v1"}},
		&rbacv1.ClusterRole{TypeMeta: metav1.TypeMeta{Kind: "ClusterRole", APIVersion: "v1"}},
		&rbacv1.ClusterRoleBinding{TypeMeta: metav1.TypeMeta{Kind: "ClusterRoleBinding", APIVersion: "v1"}},
		&corev1.Service{TypeMeta: metav1.TypeMeta{Kind: "Service", APIVersion: "v1"}},
		&appsv1.Deployment{TypeMeta: metav1.TypeMeta{Kind: "Deployment", APIVersion: "v1"}},
		&autoscalingv2beta2.HorizontalPodAutoscaler{TypeMeta: metav1.TypeMeta{Kind: "HorizontalPodAutoscaler", APIVersion: "v2beta2"}},
	} {
		err = c.Watch(&source.Kind{Type: t}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &istiov1beta1.MeshGateway{},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileMeshGateway{}

// ReconcileMeshGateway reconciles a MeshGateway object
type ReconcileMeshGateway struct {
	client.Client
	dynamic dynamic.Interface
	scheme  *runtime.Scheme
}

// +kubebuilder:rbac:groups=istio.banzaicloud.io,resources=meshgateways;meshgateways/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=istio.banzaicloud.io,resources=meshgateways/status,verbs=get;update;patch

// Reconcile reads that state of the cluster for a MeshGateway object and makes changes based on the state read
// and what is in the MeshGateway.Spec
func (r *ReconcileMeshGateway) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger := log.WithValues("trigger", request.Namespace+"/"+request.Name, "correlationID", uuid.Must(uuid.NewV4()).String())

	// Fetch the MeshGateway instance
	instance := &istiov1beta1.MeshGateway{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	err = r.setDefaultLabels(instance)
	if err != nil {
		return reconcile.Result{}, errors.WithStack(err)
	}

	instance.SetDefaults()

	istio, err := r.getRelatedIstioCR(instance)
	if err != nil {
		updateErr := updateStatus(r.Client, instance, istiov1beta1.ReconcileFailed, err.Error(), logger)
		if updateErr != nil {
			logger.Error(updateErr, "failed to update state")
			return reconcile.Result{}, errors.WithStack(err)
		}
		return reconcile.Result{
			Requeue: false,
		}, errors.WithStack(err)
	}
	istio.SetDefaults()

	if !istio.Spec.Version.IsSupported() {
		return reconcile.Result{}, nil
	}

	instance.Spec.Labels = util.MergeStringMaps(instance.Spec.Labels, istio.RevisionLabels())

	err = updateStatus(r.Client, instance, istiov1beta1.Reconciling, "", logger)
	if err != nil {
		return reconcile.Result{}, errors.WithStack(err)
	}

	reconciler := gateways.New(r.Client, r.dynamic, istio, instance, r.scheme)

	err = reconciler.Reconcile(log)
	if err == nil {
		instance.Status.GatewayAddress, err = reconciler.GetGatewayAddress()
		if err != nil {
			log.Info(fmt.Sprintf("gateway address pending: %s", err.Error()))
			updateErr := updateStatus(r.Client, instance, istiov1beta1.ReconcileFailed, errors.Cause(err).Error(), logger)
			if updateErr != nil {
				logger.Error(updateErr, "failed to update state")
				return reconcile.Result{}, errors.WithStack(err)
			}
			return reconcile.Result{
				RequeueAfter: time.Second * 30,
			}, nil
		}
	} else {
		updateErr := updateStatus(r.Client, instance, istiov1beta1.ReconcileFailed, errors.Cause(err).Error(), logger)
		if updateErr != nil {
			logger.Error(updateErr, "failed to update state")
			return reconcile.Result{}, errors.WithStack(err)
		}
		return reconcile.Result{}, emperror.Wrap(err, "could not reconcile mesh gateway")
	}

	err = updateStatus(r.Client, instance, istiov1beta1.Available, "", logger)
	if err != nil {
		return reconcile.Result{}, errors.WithStack(err)
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileMeshGateway) setDefaultLabels(instance *istiov1beta1.MeshGateway) error {
	i := instance.DeepCopy()
	i.SetDefaultLabels()
	if !reflect.DeepEqual(i.Spec.Labels, instance.Spec.Labels) {
		gvk := instance.DeepCopy().GetObjectKind().GroupVersionKind()
		instance.Spec.Labels = i.Spec.Labels
		err := r.Update(context.Background(), instance)
		if err != nil {
			return errors.WithStack(err)
		}
		instance.SetGroupVersionKind(gvk)
	}

	return nil
}

func (r *ReconcileMeshGateway) getRelatedIstioCR(instance *istiov1beta1.MeshGateway) (*istiov1beta1.Istio, error) {
	istio := &istiov1beta1.Istio{}

	// try to get specified Istio CR
	if instance.Spec.IstioControlPlane != nil {
		err := r.Client.Get(context.Background(), client.ObjectKey{
			Name:      instance.Spec.IstioControlPlane.Name,
			Namespace: instance.Spec.IstioControlPlane.Namespace,
		}, istio)
		if err != nil {
			return nil, emperror.Wrap(err, "could not get related Istio CR")
		}

		return istio, nil
	}

	// get the oldest otherwise for backward compatibility
	var configs istiov1beta1.IstioList
	err := r.Client.List(context.TODO(), &configs)
	if err != nil {
		return nil, emperror.Wrap(err, "could not list istio resources")
	}
	if len(configs.Items) == 0 {
		return nil, errors.New("no Istio CRs were found")
	}

	sort.Sort(istiov1beta1.SortableIstioItems(configs.Items))

	config := configs.Items[0]
	gvk := config.GroupVersionKind()
	gvk.Version = istiov1beta1.SchemeGroupVersion.Version
	gvk.Group = istiov1beta1.SchemeGroupVersion.Group
	gvk.Kind = "Istio"
	config.SetGroupVersionKind(gvk)

	return &config, nil
}

func updateStatus(c client.Client, instance *istiov1beta1.MeshGateway, status istiov1beta1.ConfigState, errorMessage string, logger logr.Logger) error {
	typeMeta := instance.TypeMeta
	instance.Status.Status = status
	instance.Status.ErrorMessage = errorMessage
	err := c.Status().Update(context.Background(), instance)
	if k8serrors.IsNotFound(err) {
		err = c.Update(context.Background(), instance)
	}
	if err != nil {
		if !k8serrors.IsConflict(err) {
			return emperror.Wrapf(err, "could not update mesh gateway state to '%s'", status)
		}
		var actualInstance istiov1beta1.MeshGateway
		err := c.Get(context.TODO(), types.NamespacedName{
			Namespace: instance.Namespace,
			Name:      instance.Name,
		}, &actualInstance)
		if err != nil {
			return emperror.Wrap(err, "could not get resource for updating status")
		}
		actualInstance.Status.Status = status
		actualInstance.Status.ErrorMessage = errorMessage
		err = c.Status().Update(context.Background(), &actualInstance)
		if k8serrors.IsNotFound(err) {
			err = c.Update(context.Background(), &actualInstance)
		}
		if err != nil {
			return emperror.Wrapf(err, "could not update mesh gateway state to '%s'", status)
		}
	}

	// update loses the typeMeta of the instace that's used later when setting ownerrefs
	instance.TypeMeta = typeMeta
	logger.Info("mesh gateway state updated", "status", status)
	return nil
}

package istio

import (
	"context"

	istiov1alpha1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1alpha1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
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

var log = logf.Log.WithName("controller_istio")

// Add creates a new Istio Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	reconciler, err := newReconciler(mgr)
	if err != nil {
		return err
	}
	return add(mgr, reconciler)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) (reconcile.Reconciler, error) {
	crdC, err := apiextensionsclient.NewForConfig(mgr.GetConfig())
	if err != nil {
		return nil, err
	}
	return &ReconcileIstio{
		client:    mgr.GetClient(),
		crdClient: crdC,
		scheme:    mgr.GetScheme(),
	}, nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("istio-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Istio
	err = c.Watch(&source.Kind{Type: &istiov1alpha1.Istio{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	//// TODO(user): Modify this to be the types you create that are owned by the primary resource
	//// Watch for changes to secondary resource Pods and requeue the owner Istio
	//err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
	//	IsController: true,
	//	OwnerType:    &istiov1alpha1.Istio{},
	//})
	//if err != nil {
	//	return err
	//}

	return nil
}

var _ reconcile.Reconciler = &ReconcileIstio{}

// ReconcileIstio reconciles a Istio object
type ReconcileIstio struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client    client.Client
	crdClient apiextensionsclient.Interface
	scheme    *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Istio object and makes changes based on the state read
// and what is in the Istio.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileIstio) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Istio")

	// Fetch the Istio instance
	istio := &istiov1alpha1.Istio{}
	err := r.client.Get(context.TODO(), request.NamespacedName, istio)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue

			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	err = r.ReconcileCrds(reqLogger, istio)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcileGalley(reqLogger, istio)
	if err != nil {
		return reconcile.Result{}, err
	}

	// TODO: install galley,mixer,pilot,etc...

	// TODO: install custom resources

	return reconcile.Result{}, nil
}

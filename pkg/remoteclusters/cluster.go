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

package remoteclusters

import (
	"context"
	"sync"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/config"
	"github.com/banzaicloud/istio-operator/pkg/controller/meshgateway"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
)

type Cluster struct {
	name           string
	operatorConfig config.Configuration
	k8sConfig      []byte
	log            logr.Logger

	stop          context.Context
	stopper       context.CancelFunc
	initClient    sync.Once
	initInformers sync.Once

	mgr manager.Manager

	restConfig        *rest.Config
	ctrlRuntimeClient client.Client
	dynamicClient     dynamic.Interface
	istioConfig       *istiov1beta1.Istio
	remoteConfig      *istiov1beta1.RemoteIstio
	ctrl              controller.Controller
	cl                client.Client
}

func NewCluster(name string, cfg config.Configuration, ctrl controller.Controller, cl client.Client, k8sConfig []byte, log logr.Logger) (*Cluster, error) {
	ctx, cancel := context.WithCancel(context.Background())

	cluster := &Cluster{
		name:      name,
		k8sConfig: k8sConfig,
		log:       log.WithValues("cluster", name),
		stop:      ctx,
		stopper:   cancel,
		ctrl:      ctrl,
		cl:        cl,
	}

	restConfig, err := cluster.getRestConfig(k8sConfig)
	if err != nil {
		return nil, emperror.Wrap(err, "could not get k8s rest config")
	}
	cluster.restConfig = restConfig

	return cluster, nil
}

func (c *Cluster) GetName() string {
	return c.name
}

func (c *Cluster) initK8sInformers() error {
	for _, f := range []func() error{
		c.namespaceInformer,
		c.configMapInformer,
		c.meshGatewayInformer,
	} {
		err := f()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Cluster) namespaceInformer() error {
	if c.remoteConfig == nil {
		return errors.New("remoteconfig must be set")
	}

	namespaceInformer, err := c.mgr.GetCache().GetInformerForKind(context.Background(), corev1.SchemeGroupVersion.WithKind("Namespace"))
	if err != nil {
		return emperror.Wrap(err, "could not get informer for Namespace")
	}

	err = c.ctrl.Watch(&source.Informer{
		Informer: namespaceInformer,
	}, handler.EnqueueRequestsFromMapFunc(
		func(object client.Object) []reconcile.Request {
			return []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Name:      c.remoteConfig.Name,
						Namespace: c.remoteConfig.Namespace,
					},
				},
			}
		},
	), predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return true
		},
	})

	if err != nil {
		return emperror.Wrap(err, "could not set informer for Namespace")
	}

	return nil
}

func (c *Cluster) configMapInformer() error {
	if c.remoteConfig == nil {
		return errors.New("remoteconfig must be set")
	}

	configmapInformer, err := c.mgr.GetCache().GetInformerForKind(context.Background(), corev1.SchemeGroupVersion.WithKind("ConfigMap"))
	if err != nil {
		return emperror.Wrap(err, "could not get informer for ConfigMap")
	}

	err = c.ctrl.Watch(&source.Informer{
		Informer: configmapInformer,
	}, handler.EnqueueRequestsFromMapFunc(
		func(object client.Object) []reconcile.Request {
			return []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Name:      c.remoteConfig.Name,
						Namespace: c.remoteConfig.Namespace,
					},
				},
			}
		}), predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			if labels.SelectorFromValidatedSet(caRootConfigMapLabels).Matches(labels.Set(e.Object.GetLabels())) {
				return true
			}

			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if labels.SelectorFromValidatedSet(caRootConfigMapLabels).Matches(labels.Set(e.ObjectOld.GetLabels())) {
				return true
			}

			return false
		},
	})

	if err != nil {
		return emperror.Wrap(err, "could not set informer for ConfigMap")
	}

	return nil
}

func (c *Cluster) meshGatewayInformer() error {
	if c.remoteConfig == nil {
		return errors.New("remoteconfig must be set")
	}

	mgwInformer, err := c.mgr.GetCache().GetInformerForKind(context.Background(), istiov1beta1.SchemeGroupVersion.WithKind("MeshGateway"))
	if err != nil {
		return emperror.Wrap(err, "could not get informer for MeshGateway")
	}

	err = c.ctrl.Watch(&source.Informer{
		Informer: mgwInformer,
	}, handler.EnqueueRequestsFromMapFunc(
		func(object client.Object) []reconcile.Request {
			return []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Name:      c.remoteConfig.Name,
						Namespace: c.remoteConfig.Namespace,
					},
				},
			}
		},
	), k8sutil.GetWatchPredicateForMeshGateway())

	if err != nil {
		return emperror.Wrap(err, "could not set informer for MeshGateway")
	}

	return nil
}

func (c *Cluster) initK8SClients() error {
	// already initialized
	if c.ctrlRuntimeClient != nil {
		return nil
	}

	// add mesh gateway controller to the manager
	meshgateway.Add(c.mgr, c.operatorConfig)

	c.ctrlRuntimeClient = c.mgr.GetClient()

	dynamicClient, err := dynamic.NewForConfig(c.restConfig)
	if err != nil {
		return emperror.Wrap(err, "could not get dynamic client")
	}
	c.dynamicClient = dynamicClient

	return nil
}

func (c *Cluster) Reconcile(remoteConfig *istiov1beta1.RemoteIstio, istio *istiov1beta1.Istio) error {
	c.log.Info("reconciling remote istio")

	var ReconcilerFuncs []func(remoteConfig *istiov1beta1.RemoteIstio, istio *istiov1beta1.Istio) error

	c.remoteConfig = remoteConfig

	err := c.startManager(c.restConfig)
	if err != nil {
		return err
	}

	if err := c.reconcileCRDs(remoteConfig, istio); err != nil {
		return emperror.Wrapf(err, "could not reconcile")
	}

	// init k8s clients
	c.initClient.Do(func() {
		err = c.initK8SClients()
	})
	if err != nil {
		return emperror.Wrap(err, "could not init k8s clients")
	}

	// init k8s informers
	c.initInformers.Do(func() {
		err = c.initK8sInformers()
	})
	if err != nil {
		return emperror.Wrap(err, "could not init k8s informers")
	}

	ReconcilerFuncs = append(ReconcilerFuncs,
		c.reconcileConfig,
		c.reconcileSignCert,
		c.reconcileCARootToNamespaces,
		c.reconcileEnabledServices,
		c.ReconcileEnabledServiceEndpoints,
		c.reconcileNamespaceInjectionLabels,
		c.reconcileComponents,
	)

	for _, f := range ReconcilerFuncs {
		if err := f(remoteConfig, istio); err != nil {
			return emperror.Wrapf(err, "could not reconcile")
		}
	}

	return nil
}

func (c *Cluster) GetRemoteConfig() *istiov1beta1.RemoteIstio {
	return c.remoteConfig
}

func (c *Cluster) GetClient() (client.Client, error) {
	err := c.initK8SClients()
	if err != nil {
		return nil, err
	}

	return c.ctrlRuntimeClient, nil
}

func (c *Cluster) RemoveRemoteIstioComponents() error {
	if c.istioConfig == nil {
		return nil
	}
	c.log.Info("removing istio from remote cluster by removing installed CRDs")

	istiocrd, err := c.istiocrd()
	if err != nil {
		return emperror.Wrap(err, "could not get istio CRD")
	}
	err = c.ctrlRuntimeClient.Delete(context.TODO(), istiocrd)
	if err != nil && !k8serrors.IsNotFound(err) {
		emperror.Wrap(err, "could not remove istio CRD from remote cluster")
	}

	meshgatewaycrd, err := c.meshgatewaycrd()
	if err != nil {
		return emperror.Wrap(err, "could not get meshgateway CRD")
	}
	err = c.ctrlRuntimeClient.Delete(context.TODO(), meshgatewaycrd)
	if err != nil && k8serrors.IsNotFound(err) {
		return emperror.Wrap(err, "could not remove mesh gateway CRD from remote cluster")
	}

	return nil
}

func (c *Cluster) Shutdown() {
	c.log.Info("shutdown remote cluster manager")
	c.stopper()
}

func (c *Cluster) getRestConfig(kubeconfig []byte) (*rest.Config, error) {
	clusterConfig, err := clientcmd.Load(kubeconfig)
	if err != nil {
		return nil, emperror.Wrap(err, "could not load kubeconfig")
	}

	rest, err := clientcmd.NewDefaultClientConfig(*clusterConfig, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, emperror.Wrap(err, "could not create k8s rest config")
	}

	return rest, nil
}

func (c *Cluster) startManager(config *rest.Config) error {
	mgr, err := manager.New(config, manager.Options{
		MetricsBindAddress: "0", // disable metrics
		MapperProvider:     k8sutil.NewCachedRESTMapper,
	})
	if err != nil {
		return emperror.Wrap(err, "could not create manager")
	}

	c.mgr = mgr
	go func() {
		c.mgr.Start(c.stop)
	}()

	c.mgr.GetCache().WaitForCacheSync(c.stop)

	return nil
}

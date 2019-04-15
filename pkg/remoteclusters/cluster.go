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

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
)

type Cluster struct {
	name   string
	config []byte
	log    logr.Logger

	restConfig        *rest.Config
	ctrlRuntimeClient client.Client
	dynamicClient     dynamic.Interface
	istioConfig       *istiov1beta1.Istio
	remoteConfig      *istiov1beta1.RemoteIstio
}

func NewCluster(name string, config []byte, log logr.Logger) (*Cluster, error) {
	cluster := &Cluster{
		name:   name,
		config: config,
		log:    log.WithValues("cluster", name),
	}

	err := cluster.initK8SClients()
	if err != nil {
		return nil, emperror.Wrap(err, "could not re-init k8s clients")
	}

	return cluster, nil
}

func (c *Cluster) GetName() string {
	return c.name
}

func (c *Cluster) initK8SClients() error {
	restConfig, err := c.getRestConfig(c.config)
	if err != nil {
		return emperror.Wrap(err, "could not get k8s rest config")
	}
	c.restConfig = restConfig

	ctrlRuntimeClient, err := c.getCtrlRuntimeClient(restConfig)
	if err != nil {
		return emperror.Wrap(err, "could not get control-runtime client")
	}
	c.ctrlRuntimeClient = ctrlRuntimeClient

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return emperror.Wrap(err, "could not get dynamic client")
	}
	c.dynamicClient = dynamicClient

	return nil
}

func (c *Cluster) Reconcile(remoteConfig *istiov1beta1.RemoteIstio, istio *istiov1beta1.Istio) error {
	c.log.Info("reconciling remote istio")

	var ReconcilerFuncs []func(remoteConfig *istiov1beta1.RemoteIstio, istio *istiov1beta1.Istio) error

	ReconcilerFuncs = append(ReconcilerFuncs,
		c.reconcileConfigCrd,
		c.reconcileConfig,
		c.reconcileSignCert,
		c.reconcileEnabledServices,
		c.ReconcileEnabledServiceEndpoints,
		c.reconcileComponents,
		c.getIngressGatewayAddress,
	)

	for _, f := range ReconcilerFuncs {
		if err := f(remoteConfig, istio); err != nil {
			return emperror.Wrapf(err, "could not reconcile")
		}
	}

	c.remoteConfig = remoteConfig

	return nil
}

func (c *Cluster) GetRemoteConfig() *istiov1beta1.RemoteIstio {
	return c.remoteConfig
}

func (c *Cluster) RemoveConfig() error {
	if c.istioConfig == nil {
		return nil
	}
	c.log.Info("removing istio from remote cluster by removing its config")
	err := c.ctrlRuntimeClient.Delete(context.TODO(), c.istioConfig)
	if k8serrors.IsNotFound(err) {
		err = nil
	}

	return emperror.Wrap(err, "could not remove istio config from remote cluster")
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

func (c *Cluster) getCtrlRuntimeClient(config *rest.Config) (client.Client, error) {
	writeObj, err := client.New(config, client.Options{})
	if err != nil {
		return nil, emperror.Wrap(err, "could not create control-runtime client")
	}

	return client.DelegatingClient{
		Reader: &client.DelegatingReader{
			CacheReader:  writeObj,
			ClientReader: writeObj,
		},
		Writer:       writeObj,
		StatusClient: writeObj,
	}, nil
}

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
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
)

var ReconcilerFuncs []func(remoteConfig *istiov1beta1.RemoteConfig) error

type Cluster struct {
	name   string
	config []byte
	log    logr.Logger

	restConfig        *rest.Config
	ctrlRuntimeClient client.Client
	dynamicClient     dynamic.Interface
	istioConfig       *istiov1beta1.Config
	remoteConfig      *istiov1beta1.RemoteConfig
}

func NewCluster(name string, config []byte, log logr.Logger) (*Cluster, error) {
	cluster := &Cluster{
		name:   name,
		config: config,
		log:    log.WithValues("clusterName", name),
	}

	err := cluster.initK8SClients()
	if err != nil {
		return nil, err
	}

	return cluster, nil
}

func (c *Cluster) GetName() string {
	return c.name
}

func (c *Cluster) initK8SClients() error {
	restConfig, err := c.getRestConfig(c.config)
	if err != nil {
		return err
	}
	c.restConfig = restConfig

	ctrlRuntimeClient, err := c.getCtrlRuntimeClient(restConfig)
	if err != nil {
		return err
	}
	c.ctrlRuntimeClient = ctrlRuntimeClient

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return err
	}
	c.dynamicClient = dynamicClient

	return nil
}

func (c *Cluster) Reconcile(remoteConfig *istiov1beta1.RemoteConfig) error {
	c.log.Info("reconciling remote istio")

	ReconcilerFuncs = append(ReconcilerFuncs,
		c.reconcileConfigCrd,
		c.reconcileConfig,
		c.reconcileEnabledServices,
		c.ReconcileEnabledServiceEndpoints,
		c.reconcileDeployment,
	)

	for _, f := range ReconcilerFuncs {
		if err := f(remoteConfig); err != nil {
			return err
		}
	}

	c.remoteConfig = remoteConfig

	return nil
}

func (c *Cluster) GetRemoteConfig() *istiov1beta1.RemoteConfig {
	return c.remoteConfig
}

func (c *Cluster) RemoveConfig() error {
	c.log.Info("removing istio from remote cluster by removing its config")
	return c.ctrlRuntimeClient.Delete(context.TODO(), c.istioConfig)
}

func (c *Cluster) getRestConfig(kubeconfig []byte) (*rest.Config, error) {
	clusterConfig, err := clientcmd.Load(kubeconfig)
	if err != nil {
		return nil, err
	}

	rest, err := clientcmd.NewDefaultClientConfig(*clusterConfig, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, err
	}

	return rest, nil
}

func (c *Cluster) getCtrlRuntimeClient(config *rest.Config) (client.Client, error) {
	writeObj, err := client.New(config, client.Options{})
	if err != nil {
		return nil, err
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

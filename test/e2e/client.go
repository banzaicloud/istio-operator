/*
Copyright 2021 Banzai Cloud.

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

package e2e

import (
	"context"
	"time"

	"emperror.dev/errors"
	"github.com/spf13/viper"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/banzaicloud/istio-operator/pkg/apis"
	utilclient "github.com/banzaicloud/istio-operator/test/e2e/util/client"
)

func init() {
	scheme := utilclient.GetScheme()
	err := apis.AddToSchemes.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}
}

func GetK8sClient() (client.Client, error) {
	return utilclient.NewClientFromKubeConfigAndContext(viper.GetString("kubeconfig"), viper.GetString("kubecontext"))
}

func GetK8sDynamicClient() (dynamic.Interface, error) {
	config, err := utilclient.GetConfigWithContext(viper.GetString("kubeconfig"), viper.GetString("kubecontext"))
	if err != nil {
		return nil, errors.WrapIf(err, "could not get k8s config")
	}

	return dynamic.NewForConfig(config)
}

func getClient() client.Client {
	k8sClient, err := GetK8sClient()
	if err != nil {
		panic(err)
	}
	return k8sClient
}

func getDynamicClient() dynamic.Interface {
	dynamicClient, err := GetK8sDynamicClient()
	if err != nil {
		panic(err)
	}
	return dynamicClient
}

// waitForClientReady waits until the client is ready to be used, by checking if CRDs can be listed.
func waitForClientReady(c client.Client, timeout time.Duration, interval time.Duration) error {
	var crdList apiextensionsv1.CustomResourceDefinitionList
	err := c.List(context.TODO(), &crdList)
	if err == nil {
		return nil
	}

	// TODO use waitforcondition
	timer := time.After(timeout)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-timer:
			return err
		case <-ticker.C:
			err = c.List(context.TODO(), &crdList)
			if err == nil {
				return nil
			}
		}
	}
}

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
	"fmt"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	//"sigs.k8s.io/controller-runtime/pkg/reconcile"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/config"
	"github.com/banzaicloud/istio-operator/pkg/crds"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

var c client.Client

const timeout = time.Second * 5

func TestReconcile(t *testing.T) {
	// TODO randomize!
	const namespace = "default"
	// TODO randomize?
	const instanceName = "istio"
	instance := mkMinimalIstio(namespace, instanceName)

	logf.SetLogger(util.CreateLogger(true, true))
	log := logf.Log.WithName(t.Name())

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(k8sConfig, manager.Options{
		MetricsBindAddress: "0",
		Logger:             log,
	})
	require.NoError(t, err)
	c = mgr.GetClient()

	dynamic, err := dynamic.NewForConfig(mgr.GetConfig())
	require.NoError(t, err)

	crd, err := crds.New(mgr, istiov1beta1.SupportedIstioVersion)
	require.NoError(t, err)
	err = crd.LoadCRDs()
	require.NoError(t, err)

	log.Info("Creating reconciler")
	recFn, requests := SetupTestReconcile(log, newReconciler(mgr, config.Configuration{}, dynamic, crd))
	log.Info("Creating controller")
	err = newController(mgr, recFn)
	require.NoError(t, err)

	log.Info("Starting test manager")
	stopMgr, mgrStopped := StartTestManager(mgr, t)

	defer func() {
		log.Info("Stopping manager")
		close(stopMgr)
		log.Info("Waiting for manager to stop")
		mgrStopped.Wait()
		log.Info("Manager stopped")
	}()

	log.Info("Sleeping...")
	//time.Sleep(15 * time.Second)
	log.Info("Listing all resources")
	listAllResources(t, c)

	log.Info("Creating Istio instance")
	// Create the Config object and expect the Reconcile and Deployment to be created
	err = c.Create(context.TODO(), instance)
	assert.NoError(t, err)
	defer func() {
		log.Info("Deleting Istio instance")
		err := c.Delete(context.TODO(), instance)
		if err != nil {
			t.Log(err)
		}
	}()

	g := gomega.NewGomegaWithT(t)
	var expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: instanceName, Namespace: namespace}}
	log.Info("Waiting for request to arrive")
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	// default        service/istiod-istio-sample-v19x
	assert.Eventually(t,
		ServiceExists(t, context.TODO(), c, namespace, fmt.Sprintf("istiod-%s", instanceName)),
		timeout, 100*time.Millisecond)

	// default              deployment.apps/istiod-istio-sample-v19x
	assert.Eventually(t,
		DeploymentExists(t, context.TODO(), c, namespace, fmt.Sprintf("istiod-%s", instanceName)),
		timeout, 100*time.Millisecond)

	// default     horizontalpodautoscaler.autoscaling/istiod-autoscaler-istio-sample-v19x
	assert.Eventually(t,
		HPAExists(t, context.TODO(), c, namespace, fmt.Sprintf("istiod-autoscaler-%s", instanceName)),
		timeout, 100*time.Millisecond)

	if t.Failed() {
		log.Info("Test failed, listing resources")
		listAllResources(t, c)
	}
}

func newIstio() *istiov1beta1.Istio {
	istio := istiov1beta1.Istio{}
	istio.SetGroupVersionKind(istiov1beta1.SchemeGroupVersion.WithKind("Istio"))
	return &istio
}

func mkMinimalIstio(namespace, name string) *istiov1beta1.Istio {
	istio := istiov1beta1.Istio{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
	istio.SetGroupVersionKind(istiov1beta1.SchemeGroupVersion.WithKind("Istio"))
	istio.SetDefaults()

	istio.Spec.Version = istiov1beta1.IstioVersion(istiov1beta1.SupportedIstioVersion)

	return &istio
}

// ResourceExists checks if a resource exists in the cluster
func ResourceExists(t *testing.T, ctx context.Context, kubeClient client.Client, item runtime.Object, namespace, name string) func() bool {
	return func() bool {
		err := kubeClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, item)
		if err != nil && errors.IsNotFound(err) {
			return false
		}
		assert.NoError(t, err)
		return true
	}
}

func DeploymentExists(t *testing.T, ctx context.Context, kubeClient client.Client, namespace, name string) func() bool {
	var deployment appsv1.Deployment
	return ResourceExists(t, ctx, kubeClient, &deployment, namespace, name)
}

func ServiceExists(t *testing.T, ctx context.Context, kubeClient client.Client, namespace, name string) func() bool {
	var svc corev1.Service
	return ResourceExists(t, ctx, kubeClient, &svc, namespace, name)
}

func HPAExists(t *testing.T, ctx context.Context, kubeClient client.Client, namespace, name string) func() bool {
	var hpa autoscalingv2beta2.HorizontalPodAutoscaler
	return ResourceExists(t, ctx, kubeClient, &hpa, namespace, name)
}

func listAllResources(t *testing.T, client client.Client) {
	istios := listIstios(t, client)
	fmt.Printf("istios: %d\n", len(istios))
	for i, istio := range istios {
		fmt.Printf("istio %d: %s %s\n", i, istio.Namespace, istio.Name)
	}
	deployments := listDeployments(t, client)
	fmt.Printf("deployments: %d\n", len(deployments))
	for i, deploy := range deployments {
		fmt.Printf("deployment %d: %s %s\n", i, deploy.Namespace, deploy.Name)
	}
	services := listServices(t, client)
	fmt.Printf("services: %d\n", len(services))
	for i, svc := range services {
		fmt.Printf("service %d: %s %s\n", i, svc.Namespace, svc.Name)
	}
	hpas := listHorizontalPodAutoscalers(t, client)
	fmt.Printf("hpas: %d\n", len(hpas))
	for i, hpa := range hpas {
		fmt.Printf("hpa %d: %s %s\n", i, hpa.Namespace, hpa.Name)
	}
}

func listIstios(t *testing.T, client client.Client) []istiov1beta1.Istio {
	var istioList istiov1beta1.IstioList
	err := client.List(context.TODO(), &istioList)
	assert.NoError(t, err)
	return istioList.Items
}

func listDeployments(t *testing.T, client client.Client) []appsv1.Deployment {
	var deploymentList appsv1.DeploymentList
	err := client.List(context.TODO(), &deploymentList)
	assert.NoError(t, err)
	return deploymentList.Items
}

func listServices(t *testing.T, client client.Client) []corev1.Service {
	var serviceList corev1.ServiceList
	err := client.List(context.TODO(), &serviceList)
	assert.NoError(t, err)
	return serviceList.Items
}

func listHorizontalPodAutoscalers(t *testing.T, client client.Client) []autoscalingv2beta2.HorizontalPodAutoscaler {
	var hpaList autoscalingv2beta2.HorizontalPodAutoscalerList
	err := client.List(context.TODO(), &hpaList)
	assert.NoError(t, err)
	return hpaList.Items
}

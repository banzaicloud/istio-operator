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

package e2e

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	"emperror.dev/errors"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil/mgw"
	"github.com/banzaicloud/istio-operator/test/e2e/util"
	"github.com/banzaicloud/istio-operator/test/e2e/util/resources"
)

var istioGVR = schema.GroupVersionResource{
	Group:    "istio.banzaicloud.io",
	Version:  "v1beta1",
	Resource: "istios",
}

var meshGatewayGVR = schema.GroupVersionResource{
	Group:    "istio.banzaicloud.io",
	Version:  "v1beta1",
	Resource: "meshgateways",
}

type IstioTestEnv struct {
	t *testing.T

	c     client.Client
	istio *istiov1beta1.Istio

	clusterStateBefore *ClusterState
}

func NewIstioTestEnv(t *testing.T, c client.Client, istio *istiov1beta1.Istio) IstioTestEnv {
	return IstioTestEnv{t, c, istio, nil}
}

func (e *IstioTestEnv) Start() {
	var err error
	e.clusterStateBefore, err = listAllResources(e.c)
	require.NoError(e.t, err)
	require.NotNil(e.t, e.clusterStateBefore)

	e.t.Log("Creating Istio resource")
	err = testEnv.Client.Create(context.TODO(), e.istio)
	assert.NoError(e.t, err)
}

func (e *IstioTestEnv) Close() {
	g := gomega.NewWithT(e.t)

	e.t.Log("Deleting Istio resource")
	err := e.c.Delete(context.TODO(), e.istio)
	g.Expect(err).NotTo(HaveOccurred())

	e.t.Log("Waiting for Istio resource to be deleted")
	// TODO use g.Eventually() and remove util.WaitForCondition
	err = util.WaitForCondition(10*time.Second, 100*time.Millisecond, func() (bool, error) {
		return !IstioExists(context.TODO(), e.t, e.c, e.istio.Namespace, e.istio.Name)(), nil
	})
	g.Expect(err).NotTo(HaveOccurred(), "waiting for Istio resource to be deleted")

	e.t.Log("Waiting for the cluster to be cleaned up")
	WaitForCleanup(e.t, g, *e.clusterStateBefore, 120*time.Second, 100*time.Millisecond)
}

func WaitForCleanup(t *testing.T, g *WithT, expectedClusterState ClusterState, timeout time.Duration, interval time.Duration) {
	t.Log("Waiting for cleanup")
	err := util.WaitForCondition(timeout, interval, func() (bool, error) {
		currentClusterState, err := listAllResources(testEnv.Client)
		if err != nil {
			return false, err
		}
		return clusterIsClean(expectedClusterState, *currentClusterState), nil
	})
	if err != nil {
		// The err can be a timeout, in which case it's helpful to show the resources which were not cleaned up
		t.Log("Got error while waiting for cluster cleanup. Rechecking to give more detail.", err)
		clusterStateAfter, err := listAllResources(testEnv.Client)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(*clusterStateAfter).To(Equal(expectedClusterState))
	} else {
		t.Log("Cleaned up")
	}
}

func (e *IstioTestEnv) WaitForIstioReconcile() {
	g := gomega.NewWithT(e.t)

	e.t.Log("Waiting for Istio resource to be reconciled")
	err := WaitForStatus(istioGVR, e.istio.Namespace, e.istio.Name, string(istiov1beta1.Available), 300*time.Second, 1000*time.Millisecond)
	g.Expect(err).NotTo(HaveOccurred(), "waiting for Istio resource to be reconciled")
	e.t.Log("Istio resource is reconciled")
}

func WaitForStatus(gvr schema.GroupVersionResource, namespace, name string, expectedStatus string, timeout time.Duration, interval time.Duration) error {
	return util.WaitForCondition(timeout, interval, func() (bool, error) {
		status, err := GetStatus(context.TODO(), testEnv.Dynamic, gvr, namespace, name)
		if err != nil {
			return false, err
		}
		return status == expectedStatus, nil
	})
}

// TODO use Ginkgo?
func TestIstioOperator(t *testing.T) {
	g := gomega.NewWithT(t)

	// TODO randomize!
	const namespace = "default"
	// TODO randomize?
	const instanceName = "istio"
	instance := mkMinimalIstio(namespace, instanceName)

	istioTestEnv := NewIstioTestEnv(t, testEnv.Client, instance)
	istioTestEnv.Start()
	defer func() {
		if t.Failed() {
			t.Log("Test failed, not cleaning up")
		} else {
			istioTestEnv.Close()
		}
	}()

	istioTestEnv.WaitForIstioReconcile()

	// TODO extract this and related code
	clusterStateBeforeTests, err := listAllResources(testEnv.Client)
	g.Expect(err).NotTo(HaveOccurred())

	t.Run("Istio resource stays reconciled (Available)", func(t *testing.T) {
		g := gomega.NewWithT(t)
		defer func() {
			if t.Failed() {
				t.Log("Test failed, not cleaning up")
			} else {
				WaitForCleanup(t, g, *clusterStateBeforeTests, 1*time.Second, 100*time.Millisecond)
			}
		}()

		isAvailableConsistently, err := IstioResourceIsAvailableConsistently(t, namespace, instanceName, 5*time.Second, 100*time.Millisecond)
		if !isAvailableConsistently || err != nil {
			// TODO this is temporary, until the exact issue can be tracked down
			t.Logf("Not available consistently (%v). Re-checking with a longer timeout", err)
			isAvailableConsistently2, err2 := IstioResourceIsAvailableConsistently(t, namespace, instanceName, 5*time.Minute, 100*time.Millisecond)
			t.Logf("Result: %v %v", isAvailableConsistently2, err2)
			t.Log("Failing because of the failure with a short timeout")
			t.Fail()
		}
		g.Expect(isAvailableConsistently, err).Should(BeTrue())

		// TODO Check that the expected CRDs, deployments, services, etc. are present
	})

	t.Run("MeshGateway sets up working ingress", func(t *testing.T) {
		g := gomega.NewWithT(t)
		resourcesFile := filepath.Join("testdata", t.Name()+".yaml")

		resources.MustCreateResources(t, testEnv.Client, resourcesFile)
		defer func() {
			if t.Failed() {
				t.Log("Test failed, not cleaning up")
			} else {
				resources.MustDeleteResources(t, testEnv.Client, resourcesFile)
				WaitForCleanup(t, g, *clusterStateBeforeTests, 60*time.Second, 100*time.Millisecond)
			}
		}()

		meshGatewayAddress, err := GetMeshGatewayAddress("istio-system", "mgw01", 30*time.Second, 100*time.Millisecond)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(URLIsAccessible(t, fmt.Sprintf("http://%s:80/get", meshGatewayAddress), 30*time.Second, 100*time.Millisecond)).To(Succeed())
	})

	t.Log("Test done")
}

func URLIsAccessible(t *testing.T, url string, timeout time.Duration, interval time.Duration) error {
	t.Log("Checking url", url)

	var response *http.Response
	var body []byte
	err := util.WaitForCondition(timeout, interval, func() (bool, error) {
		// TODO set up timeout, etc.
		var err error
		response, err = http.Get(url)
		if err != nil {
			return false, err
		}

		body, err = ioutil.ReadAll(response.Body)
		if err != nil {
			return false, err
		}

		// TODO also check that the response body has the expected contents
		return response.StatusCode == http.StatusOK, nil
	})

	if response != nil {
		t.Logf("Final response: url=%v proto=%v statusCode=%v status=%v header=%v body=%s",
			url, response.Proto, response.StatusCode, response.Status, response.Header, string(body))
	} else {
		t.Logf("No response: url=%v", url)
	}

	return err
}

func GetMeshGatewayAddress(mgw01Namespace string, mgw01Name string, timeout time.Duration, interval time.Duration) (string, error) {
	mgw01ObjectKey := client.ObjectKey{Namespace: mgw01Namespace, Name: mgw01Name}

	var meshGatewayAddresses []string
	err := util.WaitForCondition(timeout, interval, func() (bool, error) {
		var err error
		status, err := GetStatus(context.TODO(), testEnv.Dynamic, meshGatewayGVR, mgw01Namespace, mgw01Name)
		if err != nil {
			return false, err
		}
		if status != string(istiov1beta1.Available) {
			return false, nil
		}

		meshGatewayAddresses, err = mgw.GetMeshGatewayAddress(testEnv.Client, mgw01ObjectKey)
		if err != nil {
			return false, err
		}
		if len(meshGatewayAddresses) < 1 {
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		return "", err
	}

	return meshGatewayAddresses[0], nil
}

// TODO simplify this a bit
func IstioResourceIsAvailableConsistently(t *testing.T, namespace, name string, timeout time.Duration, interval time.Duration) (bool, error) {
	// check that it stays reconciled. Give it some slack because it is regularly re-reconciled, so there will be
	// blips of "Reconciling".
	t.Log(time.Now(), "Checking that Istio resource stays reconciled")
	timer := time.After(timeout)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	configStateOccurrences := make(map[istiov1beta1.ConfigState]int)
	var prevState istiov1beta1.ConfigState
y:
	for {
		select {
		case <-ticker.C:
			status, err := GetStatus(context.TODO(), testEnv.Dynamic, istioGVR, namespace, name)
			if err != nil {
				return false, err
			}
			configState := istiov1beta1.ConfigState(status)
			configStateOccurrences[configState] += 1
			if prevState != configState {
				if prevState == "" {
					t.Logf("%v Initial state: %v", time.Now(), configState)
				} else {
					t.Logf("%v Observed transition to %v state", time.Now(), configState)
				}
				prevState = configState
			}
		case <-timer:
			break y
		}
	}

	t.Logf("%v Observed Istio resource states: %v", time.Now(), configStateOccurrences)
	occurrencesOfAvailable := configStateOccurrences[istiov1beta1.Available]
	if occurrencesOfAvailable <= 0 {
		return false, errors.NewWithDetails(
			"expected occurrences of 'Available' to be greater than 0",
			"occurrencesOfAvailable",
			occurrencesOfAvailable,
		)
	}

	sum := 0
	for _, n := range configStateOccurrences {
		sum += n
	}
	otherConfigStates := float64(sum - occurrencesOfAvailable)
	maxOtherConfigStates := 0.1 * float64(sum)
	if otherConfigStates >= maxOtherConfigStates {
		return false, errors.NewWithDetails(
			"expected otherConfigStates to be less than maxOtherConfigStates",
			"otherConfigStates",
			otherConfigStates,
			"maxOtherConfigStates",
			maxOtherConfigStates,
		)
	}

	return true, nil
}

func GetStatus(ctx context.Context, d dynamic.Interface, gvr schema.GroupVersionResource, namespace, name string) (string, error) {
	obj, err := d.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	status := obj.Object["status"]
	if status == nil {
		// The resource was probably just created, so an error wouldn't be helpful
		return "", nil
	}

	statusObj, ok := status.(map[string]interface{})
	if !ok {
		return "", errors.NewWithDetails("unexpected type for status", "status", status)
	}

	statusStatus, ok := statusObj["Status"]
	if !ok {
		return "", errors.NewWithDetails("status.Status is not set", "status", status)
	}

	statusStatusString, ok := statusStatus.(string)
	if !ok {
		return "", errors.NewWithDetails("status.Status is not a string", "status.Status", statusStatus)
	}
	return statusStatusString, nil
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

// TODO Remove `t` arg from this and similar functions (the assertion should happen closer to the test function,
//  probably in the test function itself)
// ResourceExists checks if a resource exists in the cluster
func ResourceExists(ctx context.Context, t *testing.T, kubeClient client.Client, item runtime.Object, namespace, name string) func() bool {
	return func() bool {
		err := kubeClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, item)
		if err != nil && k8sapierrors.IsNotFound(err) {
			return false
		}
		assert.NoError(t, err)
		return true
	}
}

func IstioExists(ctx context.Context, t *testing.T, kubeClient client.Client, namespace, name string) func() bool {
	var istio istiov1beta1.Istio
	return ResourceExists(ctx, t, kubeClient, &istio, namespace, name)
}

func DeploymentExists(ctx context.Context, t *testing.T, kubeClient client.Client, namespace, name string) func() bool {
	var deployment appsv1.Deployment
	return ResourceExists(ctx, t, kubeClient, &deployment, namespace, name)
}

func ServiceExists(ctx context.Context, t *testing.T, kubeClient client.Client, namespace, name string) func() bool {
	var svc corev1.Service
	return ResourceExists(ctx, t, kubeClient, &svc, namespace, name)
}

func HPAExists(ctx context.Context, t *testing.T, kubeClient client.Client, namespace, name string) func() bool {
	var hpa autoscalingv2beta2.HorizontalPodAutoscaler
	return ResourceExists(ctx, t, kubeClient, &hpa, namespace, name)
}

// TODO this should be a `map[metav1.GroupVersionKind][]types.NamespacedName`
type ClusterState struct {
	Istios      []types.NamespacedName
	Deployments []types.NamespacedName
	Pods        []types.NamespacedName
	Services    []types.NamespacedName
	Hpas        []types.NamespacedName
}

func clusterIsClean(before ClusterState, after ClusterState) bool {
	return reflect.DeepEqual(before, after)
}

// TODO add more resource types
func listAllResources(client client.Client) (*ClusterState, error) {
	istios, err := listIstios(client)
	if err != nil {
		return nil, err
	}
	deployments, err := listDeployments(client)
	if err != nil {
		return nil, err
	}
	pods, err := listPods(client)
	if err != nil {
		return nil, err
	}
	services, err := listServices(client)
	if err != nil {
		return nil, err
	}
	hpas, err := listHorizontalPodAutoscalers(client)
	if err != nil {
		return nil, err
	}
	return &ClusterState{istios, deployments, pods, services, hpas}, nil
}

// TODO rewrite these using the dynamic client
func listIstios(client client.Client) ([]types.NamespacedName, error) {
	var istioList istiov1beta1.IstioList
	err := client.List(context.TODO(), &istioList)
	if err != nil {
		return nil, err
	}

	result := make([]types.NamespacedName, len(istioList.Items))
	for i, item := range istioList.Items {
		result[i] = types.NamespacedName{Namespace: item.Namespace, Name: item.Name}
	}
	sortNamespacedNames(result)
	return result, nil
}

func listDeployments(client client.Client) ([]types.NamespacedName, error) {
	var deploymentList appsv1.DeploymentList
	err := client.List(context.TODO(), &deploymentList)
	if err != nil {
		return nil, err
	}

	result := make([]types.NamespacedName, len(deploymentList.Items))
	for i, item := range deploymentList.Items {
		result[i] = types.NamespacedName{Namespace: item.Namespace, Name: item.Name}
	}
	sortNamespacedNames(result)
	return result, nil
}

func listPods(client client.Client) ([]types.NamespacedName, error) {
	var podList corev1.PodList
	err := client.List(context.TODO(), &podList)
	if err != nil {
		return nil, err
	}

	result := make([]types.NamespacedName, len(podList.Items))
	for i, item := range podList.Items {
		result[i] = types.NamespacedName{Namespace: item.Namespace, Name: item.Name}
	}
	sortNamespacedNames(result)
	return result, nil
}

func listServices(client client.Client) ([]types.NamespacedName, error) {
	var serviceList corev1.ServiceList
	err := client.List(context.TODO(), &serviceList)
	if err != nil {
		return nil, err
	}

	result := make([]types.NamespacedName, len(serviceList.Items))
	for i, item := range serviceList.Items {
		result[i] = types.NamespacedName{Namespace: item.Namespace, Name: item.Name}
	}
	sortNamespacedNames(result)
	return result, nil
}

func listHorizontalPodAutoscalers(client client.Client) ([]types.NamespacedName, error) {
	var hpaList autoscalingv2beta2.HorizontalPodAutoscalerList
	err := client.List(context.TODO(), &hpaList)
	if err != nil {
		return nil, err
	}

	result := make([]types.NamespacedName, len(hpaList.Items))
	for i, item := range hpaList.Items {
		result[i] = types.NamespacedName{Namespace: item.Namespace, Name: item.Name}
	}
	sortNamespacedNames(result)
	return result, nil
}

func sortNamespacedNames(nns []types.NamespacedName) {
	sort.Slice(nns, func(i, j int) bool {
		if nns[i].Namespace < nns[j].Namespace {
			return true
		}
		if nns[i].Namespace == nns[j].Namespace {
			return false
		}

		if nns[i].Name < nns[j].Name {
			return true
		}
		return false
	})
}

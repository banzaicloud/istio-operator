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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil/mgw"
	"github.com/banzaicloud/istio-operator/pkg/resources/gvr"
	"github.com/banzaicloud/istio-operator/test/e2e/util"
)

func getLoggerName(gtd ginkgo.GinkgoTestDescription) string {
	return strings.Join(gtd.ComponentTexts, "/")
}

func shouldFailFast() bool {
	return os.Getenv("E2E_TEST_FAIL_FAST") == "1"
}

func maybeCleanup(log logr.Logger, noCleanupMsg string, cleanup func()) {
	if shouldFailFast() && ginkgo.CurrentGinkgoTestDescription().Failed {
		log.Info(noCleanupMsg)
	} else {
		cleanup()
	}
}

type IstioTestEnv struct {
	log                logr.Logger
	c                  client.Client
	d                  dynamic.Interface
	istio              *istiov1beta1.Istio
	clusterStateBefore ClusterResourceList
}

func NewIstioTestEnv(log logr.Logger, c client.Client, d dynamic.Interface, istio *istiov1beta1.Istio) IstioTestEnv {
	return IstioTestEnv{log, c, d, istio, nil}
}

func (e *IstioTestEnv) Start() {
	var err error
	e.clusterStateBefore, err = listAllResources(e.d)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(e.clusterStateBefore).NotTo(gomega.BeNil())

	e.log.Info("Creating Istio resource")
	err = e.c.Create(context.TODO(), e.istio)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

func (e *IstioTestEnv) Close() {
	e.log.Info("Deleting Istio resource")
	err := e.c.Delete(context.TODO(), e.istio)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	e.log.Info("Waiting for Istio resource to be deleted")
	// TODO use g.Eventually() and remove util.WaitForCondition
	err = util.WaitForCondition(10*time.Second, 100*time.Millisecond, func() (bool, error) {
		return !IstioExists(context.TODO(), e.c, e.istio.Namespace, e.istio.Name)(), nil
	})
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "waiting for Istio resource to be deleted")

	e.log.Info("Waiting for the cluster to be cleaned up")
	WaitForCleanup(e.log, e.clusterStateBefore, 120*time.Second, 100*time.Millisecond)
}

func WaitForCleanup(log logr.Logger, expectedClusterState ClusterResourceList, timeout time.Duration, interval time.Duration) {
	log.Info("Waiting for cleanup")
	err := util.WaitForCondition(timeout, interval, func() (bool, error) {
		currentClusterState, err := listAllResources(testEnv.DynamicClient)
		if err != nil {
			return false, err
		}
		return clusterIsClean(expectedClusterState, currentClusterState), nil
	})
	if err != nil {
		// The err can be a timeout, in which case it's helpful to show the resources which were not cleaned up
		log.Error(err, "Got error while waiting for cluster cleanup. Rechecking to give more detail.")
		clusterStateAfter, err := listAllResources(testEnv.DynamicClient)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(clusterStateAfter).To(gomega.Equal(expectedClusterState))
	} else {
		log.Info("Cleaned up")
	}
}

func (e *IstioTestEnv) WaitForIstioReconcile() {
	e.log.Info("Waiting for Istio resource to be reconciled")
	err := WaitForStatus(gvr.Istio, e.istio.Namespace, e.istio.Name, string(istiov1beta1.Available), 300*time.Second, 1000*time.Millisecond)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "waiting for Istio resource to be reconciled")
	e.log.Info("Istio resource is reconciled")
}

func WaitForStatus(gvr schema.GroupVersionResource, namespace, name string, expectedStatus string, timeout time.Duration, interval time.Duration) error {
	return util.WaitForCondition(timeout, interval, func() (bool, error) {
		status, err := GetStatus(context.TODO(), testEnv.DynamicClient, gvr, namespace, name)
		if err != nil {
			return false, err
		}
		return status == expectedStatus, nil
	})
}

func URLIsAccessible(log logr.Logger, url string, timeout time.Duration, interval time.Duration) error {
	log.Info("Checking url", "url", url)

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
		log.Info("Final response",
			"url", url, "proto", response.Proto, "statusCode", response.StatusCode, "status", response.Status,
			"header", response.Header, "body", string(body))
	} else {
		log.Info("No response", "url", url)
	}

	return err
}

func GetMeshGatewayAddress(mgw01Namespace string, mgw01Name string, timeout time.Duration, interval time.Duration) (string, error) {
	mgw01ObjectKey := client.ObjectKey{Namespace: mgw01Namespace, Name: mgw01Name}

	var meshGatewayAddresses []string
	err := util.WaitForCondition(timeout, interval, func() (bool, error) {
		var err error
		status, err := GetStatus(context.TODO(), testEnv.DynamicClient, gvr.MeshGateway, mgw01Namespace, mgw01Name)
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
func IstioResourceIsAvailableConsistently(log logr.Logger, namespace, name string, timeout time.Duration, interval time.Duration) (bool, error) {
	// check that it stays reconciled. Give it some slack because it is regularly re-reconciled, so there will be
	// blips of "Reconciling".
	log.Info("Checking that Istio resource stays reconciled")
	timer := time.After(timeout)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	configStateOccurrences := make(map[istiov1beta1.ConfigState]int)
	var prevState istiov1beta1.ConfigState
y:
	for {
		select {
		case <-ticker.C:
			status, err := GetStatus(context.TODO(), testEnv.DynamicClient, gvr.Istio, namespace, name)
			if err != nil {
				return false, err
			}
			configState := istiov1beta1.ConfigState(status)
			configStateOccurrences[configState] += 1
			if prevState != configState {
				if prevState == "" {
					log.Info("Initial state", "state", configState)
				} else {
					log.Info("Observed transition to new state", "prevState", prevState, "newState", configState)
				}
				prevState = configState
			}
		case <-timer:
			break y
		}
	}

	log.Info("Observed Istio resource states", "resourceStates", configStateOccurrences)
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

	istio.Spec.Version = istiov1beta1.IstioVersion(istiov1beta1.SupportedIstioVersion)

	return &istio
}

// TODO Do not use gomega in this and similar functions (the assertion should happen closer to the test function,
//  probably in the test function itself)
// ResourceExists checks if a resource exists in the cluster
func ResourceExists(ctx context.Context, kubeClient client.Client, item runtime.Object, namespace, name string) func() bool {
	return func() bool {
		err := kubeClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, item)
		if err != nil && k8sapierrors.IsNotFound(err) {
			return false
		}
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return true
	}
}

func IstioExists(ctx context.Context, kubeClient client.Client, namespace, name string) func() bool {
	var istio istiov1beta1.Istio
	return ResourceExists(ctx, kubeClient, &istio, namespace, name)
}

func DeploymentExists(ctx context.Context, kubeClient client.Client, namespace, name string) func() bool {
	var deployment appsv1.Deployment
	return ResourceExists(ctx, kubeClient, &deployment, namespace, name)
}

func ServiceExists(ctx context.Context, kubeClient client.Client, namespace, name string) func() bool {
	var svc corev1.Service
	return ResourceExists(ctx, kubeClient, &svc, namespace, name)
}

func HPAExists(ctx context.Context, kubeClient client.Client, namespace, name string) func() bool {
	var hpa autoscalingv2beta2.HorizontalPodAutoscaler
	return ResourceExists(ctx, kubeClient, &hpa, namespace, name)
}

type groupVersionResourceString string
type ClusterResourceList map[groupVersionResourceString][]types.NamespacedName

func clusterIsClean(before ClusterResourceList, after ClusterResourceList) bool {
	return reflect.DeepEqual(before, after)
}

// TODO add more resource types
func listAllResources(d dynamic.Interface) (ClusterResourceList, error) {
	// This list should probably match the list in dump-cluster-state-and-logs.sh
	gvrs := []schema.GroupVersionResource{
		gvr.Service,
		gvr.Pod,
		gvr.Deployment,
		gvr.HorizontalPodAutoscaler,
		gvr.ClusterRole,
		gvr.ClusterRoleBinding,
		gvr.ValidatingWebhookConfiguration,
		gvr.MutatingWebhookConfiguration,
		gvr.DestinationRule,
		gvr.VirtualService,
		gvr.PeerAuthentication,
		gvr.Gateway,
		gvr.EnvoyFilter,
		gvr.Istio,
		gvr.MeshGateway,
	}

	result := make(ClusterResourceList)
	for _, gvr := range gvrs {
		items, err := listResources(d, gvr)
		if err != nil && !k8sapierrors.IsNotFound(err) {
			return nil, err
		}
		gvrString := groupVersionResourceString(gvr.String())
		if items == nil {
			items = []types.NamespacedName{}
		}
		result[gvrString] = items
	}
	return result, nil
}

func listResources(d dynamic.Interface, gvr schema.GroupVersionResource) ([]types.NamespacedName, error) {
	list, err := d.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]types.NamespacedName, len(list.Items))
	for i, item := range list.Items {
		result[i] = types.NamespacedName{
			Namespace: item.GetNamespace(),
			Name:      item.GetName(),
		}
	}

	sortNamespacedNames(result)

	return result, nil
}

func sortNamespacedNames(nns []types.NamespacedName) {
	sort.Slice(nns, func(i, j int) bool {
		if nns[i].Namespace < nns[j].Namespace {
			return true
		}
		if nns[i].Namespace > nns[j].Namespace {
			return false
		}

		if nns[i].Name < nns[j].Name {
			return true
		}
		return false
	})
}

// Note: This method is currently not being used in the E2E test.
/*
func testDataPath(description ginkgo.GinkgoTestDescription) string {
	path := filepath.Join(description.ComponentTexts...)
	return strings.ReplaceAll(path, " ", "_")
}
*/

// Get unstructured object with Kuberentes dynamic clients.
func GetUnstructuredObject(ctx context.Context, d dynamic.Interface, gvr schema.GroupVersionResource,
	resource types.NamespacedName) (*unstructured.Unstructured, error) {
	var (
		unstructuredObject *unstructured.Unstructured
		err                error
	)

	unstructuredObject, err = d.Resource(gvr).Namespace(resource.Namespace).Get(ctx, resource.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return unstructuredObject, nil
}

// Get Deployment object with Kubernetes typed clients.
func GetDeployment(ctx context.Context, c client.Client, resource types.NamespacedName) (*appsv1.Deployment, error) {
	dep := &appsv1.Deployment{}

	err := c.Get(ctx, resource, dep)
	if err != nil {
		return dep, err
	}

	return dep, nil
}

// Get Service object with Kubernetes typed clients.
func GetService(ctx context.Context, c client.Client, resource types.NamespacedName) (*corev1.Service, error) {
	svc := &corev1.Service{}

	err := c.Get(ctx, resource, svc)
	if err != nil {
		return svc, err
	}

	return svc, nil
}

// Get a container list of given Deployment object.
func GetContainersFromDeployment(dep *appsv1.Deployment) []corev1.Container {
	return dep.Spec.Template.Spec.Containers
}

// Validate if the container exists in given container list.
func ContainerExists(containerList []corev1.Container, containerName string) error {
	for _, container := range containerList {
		if container.Name == containerName {
			return nil
		}
	}

	return fmt.Errorf("%s does not exist in deployment", containerName)
}

// Wait until the Deployment is available for being acquired through API calls.
func WaitForDeployment(c client.Client, resource types.NamespacedName, timeout time.Duration,
	interval time.Duration) (*appsv1.Deployment, error) {
	dep := &appsv1.Deployment{}

	// Wait until Deployment is available
	err := util.WaitForCondition(timeout, interval, func() (bool, error) {
		var err error
		dep, err = GetDeployment(context.TODO(), c, resource)
		if err != nil {
			return false, err
		}

		return true, nil
	})
	if err != nil {
		return dep, err
	}

	return dep, nil
}

func getIstioObject(istio *istiov1beta1.Istio, namespace, name string) error {
	return testEnv.Client.Get(context.TODO(), client.ObjectKey{
		Namespace: namespace,
		Name:      name},
		istio)
}

func setMixerlessTelemetryState(istio *istiov1beta1.Istio, newState *bool) error {
	istio.Spec.MixerlessTelemetry.Enabled = newState
	// upload to cluster
	testEnv.Log.Info("Updating cluster to:", "newState", newState)
	return testEnv.Client.Update(context.TODO(), istio)
}

func waitForMixerlessTelemetryFilter(
	namespace, filterName string, filterShouldExist bool, timeout, interval time.Duration) error {
	return util.WaitForCondition(timeout, interval, func() (bool, error) {
		_, err := testEnv.DynamicClient.Resource(gvr.EnvoyFilter).Namespace(namespace).Get(
			context.TODO(),
			filterName,
			metav1.GetOptions{})
		if err != nil && !k8sapierrors.IsNotFound(err) {
			// only expect "Not Found" Error. Bail out here if we get an unexpected error
			return false, err
		}
		if k8sapierrors.IsNotFound(err) != filterShouldExist {
			// IsNotFound returns true if the error contains "Not found"
			// which is the opposite fo filterShouldExist
			return true, nil
		}
		return false, nil
	})
}

func waitForMixerlessTelemetryFilters(
	namespace, filterName1 string, filterName2 string, filterShouldExist bool, timeout, interval time.Duration) error {
	// check both filters sequentially. Usually, the first filter will wait the most.
	err := waitForMixerlessTelemetryFilter(namespace, filterName1, filterShouldExist, timeout, interval)
	if err != nil {
		return err
	}
	return waitForMixerlessTelemetryFilter(namespace, filterName2, filterShouldExist, timeout, interval)
}

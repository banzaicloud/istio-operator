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
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"

	"github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/test/e2e/util/resources"
)

var _ = Describe("E2E", func() {
	const istioResourceNamespace = "default"
	const istioResourceName = "istio"

	var (
		log                     logr.Logger
		clusterStateBeforeTests ClusterResourceList
		instance                *v1beta1.Istio
		istioTestEnv            IstioTestEnv
	)

	BeforeEach(func() {
		log = testEnv.Log.WithName(getLoggerName(CurrentGinkgoTestDescription()))

		var err error
		clusterStateBeforeTests, err = listAllResources(testEnv.DynamicClient)
		Expect(err).NotTo(HaveOccurred())

		instance = mkMinimalIstio(istioResourceNamespace, istioResourceName)
	})

	JustBeforeEach(func() {
		istioTestEnv = NewIstioTestEnv(log, testEnv.Client, testEnv.DynamicClient, instance)
		istioTestEnv.Start()
		istioTestEnv.WaitForIstioReconcile()
	})

	// TODO Move this out of this "Describe" somehow. It should automatically run after all failed tests, not just the
	//  ones in this "Describe".
	JustAfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			testEnv.ClusterStateDumper.Dump(CurrentGinkgoTestDescription())
		}
	})

	AfterEach(func() {
		maybeCleanup(log, "Test failed, not waiting for cleanup", func() {
			istioTestEnv.Close()
			WaitForCleanup(log, clusterStateBeforeTests, 60*time.Second, 100*time.Millisecond)
		})
	})

	Describe("tests with minimal istio resource", func() {
		// TODO: Unskipped when creating a PR
		Context("Istio resource", func() {
			It("should stay reconciled (Available)", func() {
				isAvailableConsistently, err := IstioResourceIsAvailableConsistently(log, istioResourceNamespace, istioResourceName, 5*time.Second, 100*time.Millisecond)
				if !isAvailableConsistently || err != nil {
					// TODO this is temporary, until the exact issue can be tracked down
					log.Info("Not available consistently. Re-checking with a longer timeout", "err", err)
					isAvailableConsistently2, err2 := IstioResourceIsAvailableConsistently(log, istioResourceNamespace, istioResourceName, 5*time.Minute, 100*time.Millisecond)
					log.Info("Result", "isAvailableConsistently2", isAvailableConsistently2, "err2", err2)
					Fail("Failing because of the failure with a short timeout")
				}
				Expect(isAvailableConsistently, err).Should(BeTrue())

				// TODO Check that the expected CRDs, deployments, services, etc. are present
			})
		})

		Context("MeshGateway", func() {
			const (
				testNamespace = "test0001"
				mgwName       = "mgw01"
			)

			var resourcesFile string

			BeforeEach(func() {
				resourcesFile = filepath.Join("testdata", testDataPath(CurrentGinkgoTestDescription())+".yaml")
			})

			JustBeforeEach(func() {
				Expect(resources.CreateResources(testEnv.Client, resourcesFile)).To(Succeed())
			})

			AfterEach(func() {
				maybeCleanup(log, "Test failed, not cleaning up", func() {
					Expect(resources.DeleteResources(testEnv.Client, resourcesFile)).To(Succeed())
				})
			})

			It("sets up working ingress", func() {
				// Construct object ObjectKey to make K8s API calls
				mgwNamespacedName := types.NamespacedName{
					Namespace: testNamespace,
					Name:      mgwName,
				}

				mgwDep, err := WaitForDeployment(istioTestEnv.c, mgwNamespacedName, 300*time.Second, 100*time.Millisecond)
				Expect(err).ShouldNot(HaveOccurred())

				const (
					mgwPodContainerAmount   int    = 1
					istioProxyContainerName string = "istio-proxy"
				)

				// Validate if there one container only in mesh-gateway pod
				containerList := GetContainersFromDeployment(mgwDep)
				Expect(len(containerList)).Should(BeIdenticalTo(mgwPodContainerAmount))

				// Check if the istio-proxy sidecar container exists in the pod
				err = ContainerExists(containerList, istioProxyContainerName)
				Expect(err).ShouldNot(HaveOccurred())

				// Check if service is created that matches mesh-gateway's name in the same namespace
				mgwSvc, err := GetService(context.TODO(), istioTestEnv.c, mgwNamespacedName)
				Expect(err).ShouldNot(HaveOccurred())

				// Validate if mgw Service label selector matches mgw Deployment's MatchLabels field
				Expect(mgwSvc.Spec.Selector).Should(BeEquivalentTo(mgwDep.Spec.Selector.MatchLabels))

				// Will face MetalLB issues outside of Linux OS
				if runtime.GOOS == "linux" {
					meshGatewayAddress, err := GetMeshGatewayAddress(testNamespace, "mgw01", 300*time.Second,
						100*time.Millisecond)
					Expect(err).NotTo(HaveOccurred())

					Expect(URLIsAccessible(log, fmt.Sprintf("http://%s:8080/get", meshGatewayAddress), 30*time.Second,
						100*time.Millisecond)).To(Succeed())
				}
			})
		})
	})
})

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
	"path/filepath"
	"time"

	"github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/test/e2e/util/resources"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// FIXME
//  ValidatingWebhookConfiguration is not deleted when an istio resource is deleted on k8s 1.20+
//  See: https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#owners-and-dependents
//
//     Note:
//     Cross-namespace owner references are disallowed by design.
//     Namespaced dependents can specify cluster-scoped or namespaced owners. A namespaced owner must
//     exist in the same namespace as the dependent. If it does not, the owner reference is treated as
//     absent, and the dependent is subject to deletion once all owners are verified absent.
//
//     Cluster-scoped dependents can only specify cluster-scoped owners. In v1.20+, if a cluster-scoped
//     dependent specifies a namespaced kind as an owner, it is treated as having an unresolveable owner
//     reference, and is not able to be garbage collected.
//
//     In v1.20+, if the garbage collector detects an invalid cross-namespace ownerReference, or a
//     cluster-scoped dependent with an ownerReference referencing a namespaced kind, a warning Event
//     with a reason of OwnerRefInvalidNamespace and an involvedObject of the invalid dependent is
//     reported. You can check for that kind of Event by running
//     kubectl get events -A --field-selector=reason=OwnerRefInvalidNamespace.
const clusterScopedToNamespaceScopedOwnerRefGCWorks = false

var _ = Describe("E2E", func() {
	const istioResourceNamespace = "default"
	const istioResourceName = "istio"

	var (
		log                     logr.Logger
		clusterStateBeforeTests *ClusterState
		instance                *v1beta1.Istio
		istioTestEnv            IstioTestEnv
	)

	BeforeEach(func() {
		log = testEnv.log.WithName(CurrentGinkgoTestDescription().FullTestText)

		var err error
		clusterStateBeforeTests, err = listAllResources(testEnv.client)
		Expect(err).NotTo(HaveOccurred())

		instance = mkMinimalIstio(istioResourceNamespace, istioResourceName)
	})

	JustBeforeEach(func() {
		istioTestEnv = NewIstioTestEnv(log, testEnv.client, testEnv.dynamic, instance)
		istioTestEnv.Start()
		istioTestEnv.WaitForIstioReconcile()
	})

	AfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			log.Info("Test failed, not waiting for cleanup")
		} else {
			// deleting the istio resource leaves a ValidatingWebhookConfiguration behind on k8s 1.20+,
			// which prevents a new istio resource to be installed, so let's operate purely by updating
			// the existing istio resource
			if clusterScopedToNamespaceScopedOwnerRefGCWorks {
				istioTestEnv.Close()
				WaitForCleanup(log, testEnv.client, *clusterStateBeforeTests, 60*time.Second, 100*time.Millisecond)
			}
		}
	})

	Describe("tests with minimal istio resource", func() {
		Context("Istio resource", func() {
			It("should stay reconciled (Available)", func() {
				isAvailableConsistently, err := IstioResourceIsAvailableConsistently(log, testEnv.dynamic, istioResourceNamespace, istioResourceName, 5*time.Second, 100*time.Millisecond)
				if !isAvailableConsistently || err != nil {
					// TODO this is temporary, until the exact issue can be tracked down
					log.Info("Not available consistently. Re-checking with a longer timeout", "err", err)
					isAvailableConsistently2, err2 := IstioResourceIsAvailableConsistently(log, testEnv.dynamic, istioResourceNamespace, istioResourceName, 5*time.Minute, 100*time.Millisecond)
					log.Info("Result", "isAvailableConsistently2", isAvailableConsistently2, "err2", err2)
					Fail("Failing because of the failure with a short timeout")
				}
				Expect(isAvailableConsistently, err).Should(BeTrue())

				// TODO Check that the expected CRDs, deployments, services, etc. are present
			})
		})

		Context("MeshGateway", func() {
			const testNamespace = "test0001"

			var resourcesFile string

			BeforeEach(func() {
				resourcesFile = filepath.Join("testdata", testDataPath(CurrentGinkgoTestDescription())+".yaml")
			})

			JustBeforeEach(func() {
				Expect(resources.CreateResources(testEnv.client, resourcesFile)).To(Succeed())
			})

			AfterEach(func() {
				if CurrentGinkgoTestDescription().Failed {
					log.Info("Test failed, not cleaning up")
				} else {
					Expect(resources.DeleteResources(testEnv.client, resourcesFile)).To(Succeed())
				}
			})

			It("sets up working ingress", func() {
				meshGatewayAddress, err := GetMeshGatewayAddress(testEnv.client, testEnv.dynamic, testNamespace, "mgw01",
					30*time.Second, 100*time.Millisecond)
				Expect(err).NotTo(HaveOccurred())

				Expect(URLIsAccessible(log, fmt.Sprintf("http://%s:8080/get", meshGatewayAddress), 30*time.Second, 100*time.Millisecond)).
					To(Succeed())
			})
		})
	})
})

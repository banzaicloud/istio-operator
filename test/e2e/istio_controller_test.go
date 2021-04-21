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
	"testing"
	"time"

	"github.com/onsi/gomega"

	"github.com/banzaicloud/istio-operator/test/e2e/util/resources"
)

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
	g.Expect(err).NotTo(gomega.HaveOccurred())

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
		g.Expect(isAvailableConsistently, err).Should(gomega.BeTrue())

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
		g.Expect(err).NotTo(gomega.HaveOccurred())

		g.Expect(URLIsAccessible(t, fmt.Sprintf("http://%s:80/get", meshGatewayAddress), 30*time.Second, 100*time.Millisecond)).To(gomega.Succeed())
	})

	t.Log("Test done")
}

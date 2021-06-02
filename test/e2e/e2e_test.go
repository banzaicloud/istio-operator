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
	"strings"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/util"
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
		clusterStateBeforeTests, err = listAllResources(testEnv.Dynamic)
		Expect(err).NotTo(HaveOccurred())

		instance = mkMinimalIstio(istioResourceNamespace, istioResourceName)
	})

	JustBeforeEach(func() {
		istioTestEnv = NewIstioTestEnv(log, testEnv.Client, testEnv.Dynamic, instance)
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
		Context("Mixerless Telemetry Stats Filter Test", func() {
			var (
				istio        v1beta1.Istio
				majorMinor   string
			)
			const timeout = 30 * time.Second
			const interval = 1 * time.Second

			JustBeforeEach(func() {
				// get the istio object created in the OUTER JustBeforeEach
				log.Info("Namespace: ", "Namespace", instance.Namespace)
				log.Info("Name: ", "Name", instance.Name)
				Expect(GetIstioObject(&istio, instance.Namespace, instance.Name)).Should(Succeed())
				version := istio.Spec.Version // get first 2 digits
				versionParts := strings.SplitN(string(version), ".", 3)
				majorMinor = fmt.Sprintf("%s.%s", versionParts[0], versionParts[1])
				log.Info("Istio Version: ", "Version", majorMinor)
			})

			It("Transitions Filter through all states", func() {
				// Names of the filters of interest
				statsName := istio.WithRevision(fmt.Sprintf("mixerless-telemetry-stats-filter-%s", majorMinor))
				tcpName := istio.WithRevision(fmt.Sprintf("mixerless-telemetry-tcp-stats-filter-%s", majorMinor))
				log.Info("Filter Names: ", "Stats", statsName, "TCP", tcpName)

				log.Info("Starting filter as true")
				Expect(GetIstioObject(&istio, instance.Namespace, instance.Name)).Should(Succeed())
				Expect(SetMixerlessTelemetryState(&istio, util.BoolPointer(true))).Should(Succeed())
				Expect(WaitForMixerlessTelemetryFilters(
					istio.Namespace, statsName, tcpName, true, timeout, interval)).Should(Succeed())

				log.Info("Transition filter true -> false")
				Expect(GetIstioObject(&istio, instance.Namespace, instance.Name)).Should(Succeed())
				Expect(SetMixerlessTelemetryState(&istio, util.BoolPointer(false))).Should(Succeed())
				Expect(WaitForMixerlessTelemetryFilters(
					istio.Namespace, statsName, tcpName, false, timeout, interval)).Should(Succeed())

				log.Info("Transition filter false -> true")
				Expect(GetIstioObject(&istio, instance.Namespace, instance.Name)).Should(Succeed())
				Expect(SetMixerlessTelemetryState(&istio, util.BoolPointer(true))).Should(Succeed())
				Expect(WaitForMixerlessTelemetryFilters(
					istio.Namespace, statsName, tcpName, true, timeout, interval)).Should(Succeed())

				log.Info("Transition filter true -> nil")
				Expect(GetIstioObject(&istio, instance.Namespace, instance.Name)).Should(Succeed())
				Expect(SetMixerlessTelemetryState(&istio, nil)).Should(Succeed())
				Expect(WaitForMixerlessTelemetryFilters(
					istio.Namespace, statsName, tcpName, false, timeout, interval)).Should(Succeed())

				log.Info("Transition filter nil -> true")
				Expect(GetIstioObject(&istio, instance.Namespace, instance.Name)).Should(Succeed())
				Expect(SetMixerlessTelemetryState(&istio, util.BoolPointer(true))).Should(Succeed())
				Expect(WaitForMixerlessTelemetryFilters(
					istio.Namespace, statsName, tcpName, true, timeout, interval)).Should(Succeed())

			})
		})

		Context("MeshGateway", func() {
			const testNamespace = "test0001"

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
				if runtime.GOOS != "linux" {
					Skip("MetalLB based test only works on Linux")
				}

				meshGatewayAddress, err := GetMeshGatewayAddress(testNamespace, "mgw01", 300*time.Second, 100*time.Millisecond)
				Expect(err).NotTo(HaveOccurred())

				Expect(URLIsAccessible(log, fmt.Sprintf("http://%s:8080/get", meshGatewayAddress), 30*time.Second, 100*time.Millisecond)).
					To(Succeed())
			})
		})
	})
})

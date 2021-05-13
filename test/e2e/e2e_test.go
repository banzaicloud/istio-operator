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
	"github.com/banzaicloud/istio-operator/pkg/util"
	"strings"
	//	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"path/filepath"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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
			// var resourcesFile string
			var (
				istio        v1beta1.Istio
				filterNil    bool
				filterBefore bool
				err          error
			)

			BeforeEach(func() {
				//
			})

			JustBeforeEach(func() {
				// get the istio object created in the OUTER JustBeforeEach
				log.Info("Namespace: ", "Namespace", instance.Namespace)
				log.Info("Name: ", "Name", instance.Name)
				filterNil, filterBefore, err = GetMixerlessTelemetryStatus(&istio, instance.Namespace, instance.Name)
				Expect(err).NotTo(HaveOccurred())
				log.Info("JustBeforeEach: ", "filterNil", filterNil, "filterBefore", filterBefore)
			})

			AfterEach(func() {
				if filterNil {
					log.Info("AfterEach: Restore Nil")
				} else {
					if filterBefore {
						log.Info("AfterEach: Restore", "filterBefore", filterBefore)
					}
				}

			})

			It("Transitions Filter through all states", func() {
				var (
					err                 error
					stateTransitions    string
					previousState       string
					expectMissingFilter bool
				)
				err = GetIstioObject(&istio, instance.Namespace, instance.Name)
				Expect(err).NotTo(HaveOccurred())

				version := istio.Spec.Version // get first 2 digits
				verParts := strings.SplitN(string(version), ".", 3)
				log.Info("Control Plane: ", "Name", istio.Name)
				log.Info("Version: ", "Version", version, "Parts", verParts)
				majMinor := fmt.Sprintf("%s.%s", verParts[0], verParts[1])
				log.Info("New Version: ", "Version", majMinor)

				// Names of the filters of interest
				statsname := fmt.Sprintf("mixerless-telemetry-stats-filter-%s-%s", majMinor, istio.Name)
				tcpname := fmt.Sprintf("mixerless-tcp-telemetry-stats-filter-%s-%s", majMinor, istio.Name)
				// logically step through True -> False -> Nil -> True -> Nil -> False -> True
				// based on our starting values of filterNil and filterBefore
				if filterNil == true {
					// starting state is Nil
					previousState = "N"
					stateTransitions = "TFTN"
					//stateTransitions = "TNFTFN"
				} else {
					if filterBefore == true {
						// Starting state is True
						previousState = "T"
						stateTransitions = "FTNT"
						//stateTransitions = "FNTNFT"

					} else {
						// starting state is non-nil false
						previousState = "F"
						stateTransitions = "TNTF"
						//stateTransitions = "TFNTNF"
					}
				}
				for i, state := range stateTransitions {
					err = GetIstioObject(&istio, instance.Namespace, instance.Name)
					Expect(err).NotTo(HaveOccurred())
					switch state {
					case 'T': // transition to true
						istio.Spec.MixerlessTelemetry.Enabled = util.BoolPointer(true)
						expectMissingFilter = false
					case 'F': // transition to false
						istio.Spec.MixerlessTelemetry.Enabled = util.BoolPointer(false)
						expectMissingFilter = true
						if previousState == "N" {
							time.Sleep(15 * time.Second)
						}
					case 'N': // transition to nil
						istio.Spec.MixerlessTelemetry.Enabled = nil
						expectMissingFilter = true
						if previousState == "F" {
							time.Sleep(15 * time.Second)
						}
					}
					// upload to cluster
					log.Info("Transitioning to:",
						"iteration", i, "state", string(state), "previous", previousState)
					err = testEnv.Client.Update(context.TODO(), &istio)
					Expect(err).NotTo(HaveOccurred())

					// make sure change was applied or removed (loop)
					log.Info("Filter Names: ", "Stats", statsname, "TCP", tcpname)
					err = WaitForMixerlessTelemetryFilter(
						log, istio.Namespace, statsname, expectMissingFilter, 300*time.Second, 5*time.Second)
					Expect(err).NotTo(HaveOccurred())

					err = WaitForMixerlessTelemetryFilter(
						log, istio.Namespace, tcpname, expectMissingFilter, 300*time.Second, 5*time.Second)
					Expect(err).NotTo(HaveOccurred())
					previousState = string(state)
				}

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
				meshGatewayAddress, err := GetMeshGatewayAddress(testNamespace, "mgw01", 120*time.Second, 100*time.Millisecond)
				Expect(err).NotTo(HaveOccurred())

				Expect(URLIsAccessible(log, fmt.Sprintf("http://%s:8080/get", meshGatewayAddress), 30*time.Second, 100*time.Millisecond)).
					To(Succeed())
			})
		})
	})
})

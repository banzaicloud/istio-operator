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
	"bytes"
	"encoding/json"
	"github.com/kr/pretty"
	"testing"
	"time"

	_ "encoding/json"
	"github.com/go-logr/logr"
	_ "github.com/kr/pretty"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/banzaicloud/istio-operator/pkg/util"
	"github.com/banzaicloud/istio-operator/test/e2e/util/clusterstate"
)

type TestEnv struct {
	Log     logr.Logger
	Client  client.Client
	Dynamic dynamic.Interface

	ClusterStateDumper *clusterstate.Dumper
}

func NewTestEnv() *TestEnv {
	log := logf.Log.WithName("TestSuite")

	return &TestEnv{
		Log:     log,
		Client:  getClient(),
		Dynamic: getDynamicClient(),

		ClusterStateDumper: clusterstate.NewDumper(log),
	}
}

func TestE2E(t *testing.T) {
	logf.SetLogger(util.CreateLogger(true, true))

	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Suite")
}

var (
	testEnv            *TestEnv
	clusterStateBefore ClusterResourceList
)

var _ = BeforeSuite(func() {
	testEnv = NewTestEnv()

	err := waitForClientReady(testEnv.Client, 10*time.Second, 100*time.Millisecond)
	Expect(err).NotTo(HaveOccurred())

	clusterStateBefore, err = listAllResources(testEnv.Dynamic)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	log := testEnv.Log

	clusterStateAfter, err := listAllResources(testEnv.Dynamic)
	Expect(err).NotTo(HaveOccurred())

	if !clusterIsClean(clusterStateBefore, clusterStateAfter) {
		pretty.Print(pretty.Diff(clusterStateBefore, clusterStateAfter))

		var prettyJSON bytes.Buffer
		var bmap, amap []byte
		var err error
		bmap, err = json.Marshal(map[string]ClusterResourceList{"ClusterStateBefore": clusterStateBefore})
		if err != nil {
			log.Error(err, "Problem with marshal")
		} else {
			err = json.Indent(&prettyJSON, bmap, "", "    ")
			if err == nil {
				log.Info("cluster resources before")
				pretty.Print(string(prettyJSON.Bytes()))
			}
		}

		amap, err = json.Marshal(map[string]ClusterResourceList{"ClusterStateAfter": clusterStateAfter})
		if err != nil {
			log.Error(err, "Problem with marshal")
		} else {
			err = json.Indent(&prettyJSON, amap, "", "    ")
			if err == nil {
				log.Info("cluster resources after")
				pretty.Print(string(prettyJSON.Bytes()))
			}
		}

		Fail("Cluster wasn't cleaned up properly")
	}
})

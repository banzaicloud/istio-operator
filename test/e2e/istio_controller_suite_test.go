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
	"os"
	"testing"
	"time"

	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/banzaicloud/istio-operator/pkg/util"
)

var testEnv *TestEnv

func TestMain(m *testing.M) {
	logf.SetLogger(util.CreateLogger(true, true))
	log := logf.Log.WithName("TestMain")

	testEnv = NewTestEnv()

	err := waitForClientReady(testEnv.Client, 10*time.Second, 100*time.Millisecond)
	if err != nil {
		panic(err)
	}

	clusterStateBefore, err := listAllResources(testEnv.Client)
	if err != nil {
		panic(err)
	}

	code := m.Run()

	clusterStateAfter, err := listAllResources(testEnv.Client)
	if err != nil {
		panic(err)
	}
	if !clusterIsClean(*clusterStateBefore, *clusterStateAfter) {
		log.Info("cluster resources before", "clusterStateBefore", clusterStateBefore)
		log.Info("cluster resources after", "clusterStateAfter", clusterStateAfter)
		panic("Cluster wasn't cleaned up properly")
	}

	os.Exit(code)
}

type TestEnv struct {
	Client  client.Client
	Dynamic dynamic.Interface
}

func NewTestEnv() *TestEnv {
	return &TestEnv{
		Client:  getClient(),
		Dynamic: getDynamicClient(),
	}
}

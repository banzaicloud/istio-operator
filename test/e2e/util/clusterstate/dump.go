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

package clusterstate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/onsi/ginkgo"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/banzaicloud/istio-operator/test/e2e/util"
)

type Dumper struct {
	log logr.Logger
	dumpDir string
}

const defaultLogDir = "logs"

func NewDumper(log logr.Logger) *Dumper {
	return &Dumper{
		log:     log,
		dumpDir: getDumpDir(log),
	}
}

func (d Dumper) Dump(currentTest ginkgo.GinkgoTestDescription) {
	testDumpDir := filepath.Join(append([]string{d.dumpDir}, currentTest.ComponentTexts...)...)
	d.log.Info("Dumping cluster state and logs", "dir", testDumpDir)

	err := dump(d.log, testDumpDir)
	if err != nil {
		panic(err)
	}
}

func getDumpDir(log logr.Logger) string {
	dumpDir := os.Getenv("E2E_LOG_DIR")
	if dumpDir == "" {
		logf.Log.Info(fmt.Sprintf("Env variable E2E_LOG_DIR is not set. Using \"%s\" as log dir", defaultLogDir))
		dumpDir = defaultLogDir
	}
	dumpDir, err := filepath.Abs(dumpDir)
	if err != nil {
		panic(err)
	}
	log.Info(fmt.Sprintf("Log dir: %s", dumpDir))
	return dumpDir
}

func dump(log logr.Logger, dir string) error {
	script := os.Getenv("E2E_TEST_DUMP_SCRIPT")
	command := fmt.Sprintf("%s \"%s\"", script, dir)
	stdout, stderr, err := util.RunInShell(command)
	if err != nil {
		log.Error(err, "Failed to dump cluster state and logs. Stderr and stdout follows", "command", command)
		log.Info("stderr:\n" + stderr.String())
		log.Info("stdout:\n" + stdout.String())
		return err
	}

	return nil
}

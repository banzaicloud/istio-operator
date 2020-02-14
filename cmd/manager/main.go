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

//go:generate go run ../../pkg/crds/generate.go
//go:generate go run ../../pkg/manifests/generate.go

package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/banzaicloud/istio-operator/pkg/apis"
	"github.com/banzaicloud/istio-operator/pkg/controller"
	"github.com/banzaicloud/istio-operator/pkg/controller/istio"
	"github.com/banzaicloud/istio-operator/pkg/controller/remoteistio"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/remoteclusters"
	"github.com/banzaicloud/istio-operator/pkg/webhook"
)

const watchNamespaceEnvVar = "WATCH_NAMESPACE"
const podNamespaceEnvVar = "POD_NAMESPACE"

func main() {
	var metricsAddr string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	var developmentMode bool
	flag.BoolVar(&developmentMode, "devel-mode", false, "Set development mode (mainly for logging)")
	var shutdownWaitDuration time.Duration
	flag.DurationVar(&shutdownWaitDuration, "shutdown-wait-duration", time.Duration(30)*time.Second, "Wait duration before shutting down")
	var waitBeforeExitDuration time.Duration
	flag.DurationVar(&waitBeforeExitDuration, "wait-before-exit-duration", time.Duration(3)*time.Second, "Wait for workers to finish before exiting and removing finalizers")
	flag.Parse()
	logf.SetLogger(logf.ZapLogger(developmentMode))
	log := logf.Log.WithName("entrypoint")

	// Get a config to talk to the apiserver
	log.Info("setting up client for manager")
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "unable to set up client config")
		os.Exit(1)
	}

	namespace, err := getWatchNamespace()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	if namespace != "" {
		log.Info("watch namespace", "namespace", namespace)
	} else {
		log.Info("watch all namespaces")
	}

	// Create a new Cmd to provide shared dependencies and start components
	log.Info("setting up manager")
	mgr, err := manager.New(cfg, manager.Options{
		MetricsBindAddress: metricsAddr,
		Namespace:          namespace,
		MapperProvider:     k8sutil.NewCachedRESTMapper,
	})
	if err != nil {
		log.Error(err, "unable to set up overall controller manager")
		os.Exit(1)
	}

	log.Info("registering components")

	// Setup Scheme for all resources
	log.Info("setting up scheme")
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "unable add APIs to scheme")
		os.Exit(1)
	}

	stop := setupSignalHandler(mgr, log, shutdownWaitDuration)

	// Setup all Controllers
	log.Info("setting up controller")
	if err := controller.AddToManager(mgr, remoteclusters.NewManager(stop)); err != nil {
		log.Error(err, "unable to register controllers to the manager")
		os.Exit(1)
	}

	log.Info("setting up webhooks")
	if err := webhook.AddToManager(mgr); err != nil {
		log.Error(err, "unable to register webhooks to the manager")
		os.Exit(1)
	}

	// Start the Cmd
	log.Info("starting the Cmd.")
	if err := mgr.Start(stop); err != nil {
		log.Error(err, "unable to run the manager")
		os.Exit(1)
	}

	// Wait a bit for the workers to stop
	time.Sleep(waitBeforeExitDuration)

	// Cleanup
	log.Info("removing finalizer from Istio resources")
	err = istio.RemoveFinalizers(mgr.GetClient())
	if err != nil {
		log.Error(err, "could not remove finalizers from Istio resources")
	}
	log.Info("removing finalizer from RemoteIstio resources")
	err = remoteistio.RemoveFinalizers(mgr.GetClient())
	if err != nil {
		log.Error(err, "could not remove finalizers from RemoteIstio resources")
	}
}

func getWatchNamespace() (string, error) {
	podNamespace, found := os.LookupEnv(podNamespaceEnvVar)
	if !found {
		return "", errors.Errorf("%s env variable must be specified and cannot be empty", podNamespaceEnvVar)
	}

	watchNamespace, found := os.LookupEnv(watchNamespaceEnvVar)
	if found {
		if watchNamespace != "" && watchNamespace != podNamespace {
			return "", errors.New("watch namespace must be either empty or equal to pod namespace")
		}

	}
	return watchNamespace, nil
}

func setupSignalHandler(mgr manager.Manager, log logr.Logger, shutdownWaitDuration time.Duration) (stopCh <-chan struct{}) {
	stop := make(chan struct{})
	c := make(chan os.Signal, 2)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Info("termination signal arrived, shutdown gracefully")
		// wait a bit for deletion requests to arrive
		log.Info("wait a bit for CR deletion events to arrive", "waitSeconds", shutdownWaitDuration)
		time.Sleep(shutdownWaitDuration)
		close(stop)
		<-c
		os.Exit(1) // second signal. Exit directly.
	}()

	return stop
}

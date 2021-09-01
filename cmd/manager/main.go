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
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	_ "github.com/shurcooL/vfsgen"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	k8sConfig "sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/banzaicloud/istio-operator/pkg/apis"
	"github.com/banzaicloud/istio-operator/pkg/config"
	"github.com/banzaicloud/istio-operator/pkg/controller"
	"github.com/banzaicloud/istio-operator/pkg/controller/istio"
	"github.com/banzaicloud/istio-operator/pkg/controller/remoteistio"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/remoteclusters"
	"github.com/banzaicloud/istio-operator/pkg/trustbundle"
	"github.com/banzaicloud/istio-operator/pkg/util"
	wh "github.com/banzaicloud/istio-operator/pkg/webhook"
	"github.com/banzaicloud/istio-operator/pkg/webhook/cert"
)

const (
	watchNamespaceEnvVar = "WATCH_NAMESPACE"
	podNamespaceEnvVar   = "POD_NAMESPACE"
)

func main() {
	var metricsAddr string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	var developmentMode bool
	flag.BoolVar(&developmentMode, "devel-mode", false, "Set development mode (mainly for logging).")
	var shutdownWaitDuration time.Duration
	flag.DurationVar(&shutdownWaitDuration, "shutdown-wait-duration", time.Duration(30)*time.Second, "Wait duration before shutting down.")
	var waitBeforeExitDuration time.Duration
	flag.DurationVar(&waitBeforeExitDuration, "wait-before-exit-duration", time.Duration(3)*time.Second, "Wait for workers to finish before exiting and removing finalizers.")
	var leaderElectionEnabled bool
	flag.BoolVar(&leaderElectionEnabled, "leader-election-enabled", true, "Enable leader election for controller manager. "+
		"Enabling this will ensure there is only one active controller manager.")
	var leaderElectionNamespace string
	flag.StringVar(&leaderElectionNamespace, "leader-election-namespace", "istio-system", "Determines the namespace in which the leader election configmap will be created.")
	var leaderElectionName string
	flag.StringVar(&leaderElectionName, "leader-election-name", "istio-operator-leader-election", "Determines the name of the leader election configmap.")
	var webhookServerPort uint
	flag.UintVar(&webhookServerPort, "webhook-server-port", 9443, "The port that the webhook server serves at.")
	var webhookCertDir string
	flag.StringVar(&webhookCertDir, "webhook-cert-dir", "/tmp/certs", "Determines the directory that contains the server key and certificate.")
	var webhookConfigurationName string
	flag.StringVar(&webhookConfigurationName, "webhook-name", "istio-operator-webhook", "Sets the name of the validating webhook resource.")
	var webhookServiceAddress string
	flag.StringVar(&webhookServiceAddress, "webhook-service-address", "istio-operator-webhook", "Address (host[:port]) for the operator webhook endpoint.")
	var verboseLogging bool
	flag.BoolVar(&verboseLogging, "verbose", false, "Enable verbose logging")
	flag.Parse()

	logf.SetLogger(util.CreateLogger(verboseLogging, developmentMode))
	log := logf.Log.WithName("entrypoint")

	operatorConfig := config.Configuration{
		WebhookServiceAddress:    webhookServiceAddress,
		WebhookConfigurationName: webhookConfigurationName,
	}

	// Get a config to talk to the apiserver
	log.Info("setting up client for manager")
	k8sConfig, err := k8sConfig.GetConfig()
	if err != nil {
		log.Error(err, "unable to set up client config")
		os.Exit(1)
	}

	// try to detect support jwt policy
	supportedJWTPolicy, err := k8sutil.DetectSupportedJWTPolicy(k8sConfig)
	if err == nil {
		operatorConfig.SupportedJWTPolicy = supportedJWTPolicy
		log.Info("supported jwt policy", "policy", operatorConfig.SupportedJWTPolicy)
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
	mgr, err := manager.New(k8sConfig, manager.Options{
		MetricsBindAddress:      metricsAddr,
		Namespace:               namespace,
		MapperProvider:          k8sutil.NewCachedRESTMapper,
		LeaderElection:          leaderElectionEnabled,
		LeaderElectionNamespace: leaderElectionNamespace,
		LeaderElectionID:        leaderElectionName,
		CertDir:                 webhookCertDir,
		Port:                    int(webhookServerPort),
		Logger:                  log,
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

	ctx := setupSignalHandler(log, shutdownWaitDuration)

	// Setup all Controllers
	log.Info("setting up controller")
	if err := controller.AddToManager(mgr, remoteclusters.NewManager(ctx), operatorConfig); err != nil {
		log.Error(err, "unable to register controllers to the manager")
		os.Exit(1)
	}

	tb := trustbundle.NewManager(mgr, logf.Log.WithName("trustbundle"))
	if err := tb.Start(); err != nil {
		log.Error(err, "unable to start trust bundle manager")
		os.Exit(1)
	}

	mgr.GetWebhookServer().Register("/validate-istio-config", &webhook.Admission{Handler: wh.NewIstioResourceValidator(mgr)})
	mgr.GetWebhookServer().Register(trustbundle.WebhookEndpointPath, tb)

	whLogger := logf.Log.WithName("wh-cert-provisioner")
	certProvisioner := cert.NewCertProvisioner(whLogger, []string{}, webhookCertDir)
	err = certProvisioner.Init()
	if err != nil {
		log.Error(err, "could not init cert provisioner")
		os.Exit(1)
	}
	mgr.Add(wh.NewValidatingWebhookCertificateProvisioner(mgr, webhookConfigurationName, certProvisioner, whLogger))

	// Start the Cmd
	log.Info("starting the Cmd.")
	if err := mgr.Start(ctx); err != nil {
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

func setupSignalHandler(log logr.Logger, shutdownWaitDuration time.Duration) (ctx context.Context) {
	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 2)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Info("termination signal arrived, shutdown gracefully")
		// wait a bit for deletion requests to arrive
		log.Info("wait a bit for CR deletion events to arrive", "waitSeconds", shutdownWaitDuration)
		time.Sleep(shutdownWaitDuration)
		cancel()
		<-c
		os.Exit(1) // second signal. Exit directly.
	}()

	return ctx
}

/*
Copyright 2021 Cisco Systems, Inc. and/or its affiliates.

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

package main

import (
	"flag"
	"os"

	istionetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	istiosecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"

	// +kubebuilder:scaffold:imports
	servicemeshv1alpha1 "github.com/banzaicloud/istio-operator/v2/api/v1alpha1"
	"github.com/banzaicloud/istio-operator/v2/controllers"
	"github.com/banzaicloud/istio-operator/v2/pkg/util"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = istionetworkingv1alpha3.AddToScheme(scheme)
	_ = istiosecurityv1beta1.AddToScheme(scheme)
	_ = apiextensionv1.AddToScheme(scheme)

	_ = servicemeshv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	var developmentMode bool
	flag.BoolVar(&developmentMode, "devel-mode", false, "Set development mode (mainly for logging).")
	var leaderElectionEnabled bool
	flag.BoolVar(&leaderElectionEnabled, "leader-election-enabled", false, "Enable leader election for controller manager. "+
		"Enabling this will ensure there is only one active controller manager.")
	var leaderElectionNamespace string
	flag.StringVar(&leaderElectionNamespace, "leader-election-namespace", "istio-system", "Determines the namespace in which the leader election configmap will be created.")
	var leaderElectionName string
	flag.StringVar(&leaderElectionName, "leader-election-name", "istio-operator-leader-election", "Determines the name of the leader election configmap.")
	var webhookServerPort uint
	flag.UintVar(&webhookServerPort, "webhook-server-port", 9443, "The port that the webhook server serves at.")
	var verboseLogging bool
	flag.BoolVar(&verboseLogging, "verbose", false, "Enable verbose logging")
	flag.Parse()

	ctrl.SetLogger(util.CreateLogger(verboseLogging, developmentMode))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                  scheme,
		MetricsBindAddress:      metricsAddr,
		Port:                    int(webhookServerPort),
		LeaderElection:          leaderElectionEnabled,
		LeaderElectionID:        leaderElectionName,
		LeaderElectionNamespace: leaderElectionNamespace,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.IstioControlPlaneReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("IstioControlPlane"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "IstioControlPlane")
		os.Exit(1)
	}
	if err = (&controllers.MeshGatewayReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("MeshGateway"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "MeshGateway")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}

	// remove finalizers
	setupLog.Info("removing finalizer from controlled resources")
	err = controllers.RemoveFinalizers(mgr.GetClient())
	if err != nil {
		setupLog.Error(err, "could not remove finalizers from controlled resources")
	}
}

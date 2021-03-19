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

package server

import (
	"os"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const podNamespaceEnvVar = "POD_NAMESPACE"

var webhooks []func(manager.Manager, logr.Logger) (*admission.Webhook, error)

func Add(mgr manager.Manager) error {
	return add(mgr)
}

func add(mgr manager.Manager) error {
	log := logf.ZapLogger(false).WithName("webhook-server")

	name := "istio-operator-webhook"
	namespace, found := os.LookupEnv(podNamespaceEnvVar)
	if !found {
		return errors.Errorf("%s env variable must be specified and cannot be empty", podNamespaceEnvVar)
	}

	svr, err := webhook.NewServer(name, &ManagerWithCustomClient{
		Manager: mgr,
		client: &WebhookUpdateClient{
			Client: mgr.GetClient(),
		},
	}, webhook.ServerOptions{
		Port:    9443,
		CertDir: "/tmp/cert",
		BootstrapOptions: &webhook.BootstrapOptions{
			ValidatingWebhookConfigName: name,
			Service: &webhook.Service{
				Namespace: namespace,
				Name:      name,
				Selectors: map[string]string{
					"control-plane": "controller-manager",
				},
			},
		},
	})
	if err != nil {
		log.Error(err, "could not create new webhook server")
		os.Exit(2)
	}

	webhooksToRegister := make([]webhook.Webhook, 0)
	for _, f := range webhooks {
		wh, err := f(mgr, log)
		if err != nil {
			continue
		}
		webhooksToRegister = append(webhooksToRegister, wh)
	}

	if err := svr.Register(webhooksToRegister...); err != nil {
		log.Error(err, "could not register webhooks")
		os.Exit(2)
	} else {
		log.Info("started")
	}

	return nil
}

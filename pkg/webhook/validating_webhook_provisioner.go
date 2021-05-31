/*
Copyright 2020 Banzai Cloud.

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

package webhook

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/webhook/cert"
)

type ValidatingWebhookCertificateProvisioner struct {
	name            string
	mgr             manager.Manager
	certProvisioner *cert.Provisioner
	log             logr.Logger
	whc             *admissionregistrationv1.ValidatingWebhookConfiguration
	trigger         chan struct{}
}

func NewValidatingWebhookCertificateProvisioner(mgr manager.Manager, name string, certProvisioner *cert.Provisioner, log logr.Logger) *ValidatingWebhookCertificateProvisioner {
	return &ValidatingWebhookCertificateProvisioner{
		name:            name,
		mgr:             mgr,
		certProvisioner: certProvisioner,
		log:             log,
		trigger:         make(chan struct{}),
		whc:             &admissionregistrationv1.ValidatingWebhookConfiguration{},
	}
}

func (m *ValidatingWebhookCertificateProvisioner) Start(ctx context.Context) error {
	err := m.getWHC()
	if err != nil {
		return err
	}

	err = m.startInformer()
	if err != nil {
		return err
	}

	dnsNames := make([]string, 0)
	for _, wh := range m.whc.Webhooks {
		dnsNames = append(dnsNames, fmt.Sprintf("%s.%s.svc", wh.ClientConfig.Service.Name, wh.ClientConfig.Service.Namespace))
	}

	m.certProvisioner.SetDNSNames(dnsNames)

	m.certProvisioner.RegisterAfterGenerationFunc(func(c *cert.Certificate, needsUpdate bool) error {
		for i, wh := range m.whc.Webhooks {
			wh.ClientConfig.CABundle = c.CACert
			m.whc.Webhooks[i] = wh
		}
		if needsUpdate {
			err = k8sutil.Reconcile(m.log, m.mgr.GetClient(), m.whc, k8sutil.DesiredStatePresent)
			if err != nil {
				return err
			}
		}
		return nil
	})

	defer close(m.trigger)

	return m.certProvisioner.Start(ctx, m.trigger)
}

func (m *ValidatingWebhookCertificateProvisioner) getWHC() error {
	err := m.mgr.GetClient().Get(context.Background(), client.ObjectKey{
		Name: m.name,
	}, m.whc)
	if err != nil {
		return emperror.Wrap(err, "could net get validating webhook")
	}

	return nil
}

func (m *ValidatingWebhookCertificateProvisioner) startInformer() error {
	si, err := m.mgr.GetCache().GetInformerForKind(context.Background(), admissionregistrationv1.SchemeGroupVersion.WithKind("ValidatingWebhookConfiguration"))
	if err != nil {
		return emperror.Wrap(err, "could not get informer")
	}

	si.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, new interface{}) {
			if whc, ok := new.(*admissionregistrationv1.ValidatingWebhookConfiguration); ok && whc.Name == m.whc.Name {
				err = m.mgr.GetClient().Get(context.Background(), client.ObjectKey{
					Name: m.whc.Name,
				}, m.whc)
				if err != nil {
					return
				}
				m.trigger <- struct{}{}
			}
		},
	})

	return nil
}

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

package trustbundle

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"github.com/lestrrat-go/jwx/jwk"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
)

const (
	CertsDataKey                              = "certs"
	RevisionLabel                             = "istio.io/rev"
	TrustDomainLabel                          = "istio.banzaicloud.io/trust-domain"
	TrustBundleCASecretType corev1.SecretType = "istio.banzaicloud.io/trust-bundle-ca"
	WebhookEndpointPath                       = "/spiffe-trust-bundle"
)

type Manager struct {
	mgr manager.Manager
	log logr.Logger

	secrets       map[string]Secret
	controlplanes map[string]*istiov1beta1.Istio
	sequence      int64
}

type TrustBundle struct {
	Sequence    int64     `json:"spiffe_sequence"`
	RefreshHint float64   `json:"spiffe_refresh_hint"`
	Keys        []jwk.Key `json:"keys"`
	TrustDomain string    `json:"trust_domain,omitempty"`
}

func NewManager(mgr manager.Manager, log logr.Logger) *Manager {
	return &Manager{
		mgr: mgr,
		log: log,

		secrets:       make(map[string]Secret),
		controlplanes: make(map[string]*istiov1beta1.Istio, 0),
		sequence:      time.Now().Unix(),
	}
}

func (m *Manager) Start() error {
	err := m.startSecretInformer()
	if err != nil {
		return err
	}

	err = m.startControlplaneInformer()
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	revision := r.URL.Query().Get("revision")
	trustDomain := r.URL.Query().Get("trustDomain")
	if trustDomain == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("'trustDomain' must be specified"))
		return
	}
	if revision == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("'revision' must be specified"))
		return
	}

	tb, err := m.getTrustBundle(trustDomain, revision)
	if err != nil {
		m.log.Error(err, "failed to get trust bundle")
		return
	}

	jsonbuf, err := json.MarshalIndent(tb, "", "  ")
	if err != nil {
		m.log.Error(err, "failed to generate JSON")
		return
	}

	w.Write(jsonbuf)
}

func (m *Manager) startSecretInformer() error {
	i, err := m.mgr.GetCache().GetInformerForKind(context.Background(), corev1.SchemeGroupVersion.WithKind("Secret"))
	if err != nil {
		return err
	}

	i.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if s, ok := obj.(*corev1.Secret); ok && s.Type == TrustBundleCASecretType {
				m.updateSecret(s)
			}
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			if s, ok := newObj.(*corev1.Secret); ok && s.Type == TrustBundleCASecretType {
				m.updateSecret(s)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if s, ok := obj.(*corev1.Secret); ok && s.Type == TrustBundleCASecretType {
				m.deleteSecret(s)
			}
		},
	})

	return nil
}

func (m *Manager) startControlplaneInformer() error {
	i, err := m.mgr.GetCache().GetInformerForKind(context.Background(), istiov1beta1.SchemeGroupVersion.WithKind("Istio"))
	if err != nil {
		return err
	}

	i.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if cp, ok := obj.(*istiov1beta1.Istio); ok {
				m.addControlplane(cp)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if cp, ok := obj.(*istiov1beta1.Istio); ok {
				m.deleteControlplane(cp)
			}
		},
	})

	return nil
}

func (m *Manager) isSecretMatch(s Secret, trustDomain, revision string) bool {
	if s.Labels[TrustDomainLabel] != trustDomain {
		return false
	}

	cpFound := false
	for _, cp := range m.controlplanes {
		if cp.NamespacedRevision() == revision && s.Secret.Namespace == cp.Namespace {
			cpFound = true
		}
	}
	if !cpFound {
		return false
	}

	if s.Labels[RevisionLabel] != "" && s.Labels[RevisionLabel] != revision {
		return false
	}

	return true
}

func (m *Manager) getTrustBundle(trustDomain, revision string) (*TrustBundle, error) {
	keys := make([]jwk.Key, 0)

	for _, s := range m.secrets {
		if !m.isSecretMatch(s, trustDomain, revision) {
			continue
		}

		k, err := s.GetJWKs()
		if err != nil {
			return nil, err
		}

		keys = append(keys, k...)
	}

	tb := &TrustBundle{
		Sequence:    m.sequence,
		RefreshHint: (time.Minute * 5).Seconds(),
		Keys:        keys,
		TrustDomain: trustDomain,
	}

	return tb, nil
}

func (m *Manager) getSecretKey(s *corev1.Secret) string {
	return fmt.Sprintf("%s/%s/%s", s.Namespace, s.Name, s.UID)
}

func (m *Manager) deleteSecret(s *corev1.Secret) {
	key := m.getSecretKey(s)

	delete(m.secrets, key)

	m.log.Info("secret deleted", "key", key)
	m.sequence++
}

func (m *Manager) updateSecret(s *corev1.Secret) error {
	key := m.getSecretKey(s)

	certs, err := ParseCertficates(s.Data[CertsDataKey])
	if err != nil {
		return err
	}

	m.secrets[key] = Secret{
		Secret: s,
		Certs:  certs,
	}

	m.log.Info("secret updated", "key", key)
	m.sequence++

	return nil
}

func (m *Manager) deleteControlplane(cp *istiov1beta1.Istio) {
	key := m.getControlplaneKey(cp)

	delete(m.controlplanes, key)

	m.log.Info("controlplane deleted", "key", key)
}

func (m *Manager) getControlplaneKey(cp *istiov1beta1.Istio) string {
	return fmt.Sprintf("%s/%s/%s", cp.Namespace, cp.Name, cp.UID)
}

func (m *Manager) addControlplane(cp *istiov1beta1.Istio) error {
	key := m.getControlplaneKey(cp)

	m.controlplanes[key] = cp

	m.log.Info("controlplane added", "key", key)

	return nil
}

func ParseCertficates(raw []byte) ([]*x509.Certificate, error) {
	certs := make([]*x509.Certificate, 0)

	for {
		block, rest := pem.Decode(raw)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return certs, emperror.Wrap(err, "could not parse certificate")
			}
			certs = append(certs, cert)
		}
		raw = rest
	}

	return certs, nil
}

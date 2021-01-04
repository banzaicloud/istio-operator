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
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"

	"github.com/goph/emperror"
	"github.com/lestrrat-go/jwx/jwk"
	corev1 "k8s.io/api/core/v1"
)

type Secret struct {
	*corev1.Secret
	Revision string
	Certs    []*x509.Certificate
}

func (s Secret) GetJWKs() ([]jwk.Key, error) {
	keys := make([]jwk.Key, 0)

	for _, cert := range s.Certs {
		key, err := jwk.New(cert.PublicKey.(*rsa.PublicKey))
		if err != nil {
			return nil, emperror.Wrap(err, "failed to create JWK")
		}

		key.Set(jwk.X509CertChainKey, base64.StdEncoding.EncodeToString(cert.Raw))
		key.Set(jwk.KeyUsageKey, "x509-svid")

		keys = append(keys, key)
	}

	return keys, nil
}

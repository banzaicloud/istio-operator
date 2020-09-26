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

package cert

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/go-logr/logr"
)

type AfterCheckFunc func(cert *Certificate, needsUpdate bool) error

type Provisioner struct {
	log             logr.Logger
	dnsNames        []string
	certDir         string
	afterCheckFuncs []AfterCheckFunc
}

func NewCertProvisioner(log logr.Logger, dnsNames []string, certDir string, afterCheckFuncs ...AfterCheckFunc) *Provisioner {
	return &Provisioner{
		log:      log,
		dnsNames: dnsNames,
		certDir:  certDir,

		afterCheckFuncs: afterCheckFuncs,
	}
}

func (p *Provisioner) Init() error {
	_, _, err := p.checkCert()

	return err
}

func (p *Provisioner) SetDNSNames(dnsNames []string) {
	p.dnsNames = dnsNames
}

func (p *Provisioner) RegisterAfterGenerationFunc(f AfterCheckFunc) {
	p.afterCheckFuncs = append(p.afterCheckFuncs, f)
}

func (p *Provisioner) Start(stop <-chan struct{}, trigger <-chan struct{}) error {
	check := func(triggered bool) {
		generated, certificate, err := p.checkCert()
		if err != nil {
			p.log.Error(err, "could not refresh cert")
		}
		for _, f := range p.afterCheckFuncs {
			err = f(certificate, generated || triggered)
			if err != nil {
				p.log.Error(err, "error while running after check func")
			}
		}
	}

	check(true)

	for {
		timer := time.Tick(30 * time.Second)
		select {
		case <-trigger:
			check(true)
		case <-timer:
			check(false)
		case <-stop:
			return nil
		}
	}
}

func (p *Provisioner) checkCert() (bool, *Certificate, error) {
	c, err := ReadCertificate(p.certDir)
	if err != nil {
		return false, nil, err
	}

	valid := true
	if len(p.dnsNames) > 0 {
		for _, name := range p.dnsNames {
			if !c.Verify(name, time.Now().AddDate(0, 6, 0)) {
				valid = false
				break
			}
		}
	} else {
		valid = c.Verify("", time.Now().AddDate(0, 6, 0))
	}

	if valid {
		p.log.V(1).Info("certificate is valid")
		return false, c, nil
	}

	p.log.Info("certificate is invalid, regenerate")
	c, err = p.generateCert(p.dnsNames)
	if err != nil {
		return false, nil, err
	}

	return true, c, nil
}

func (p *Provisioner) generateCert(dnsNames []string) (*Certificate, error) {
	var certificate Certificate
	var caPEM, caPrivateKeyPEM, serverCertPEM, serverPrivKeyPEM *bytes.Buffer
	// CA config
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2020),
		Subject: pkix.Name{
			Organization: []string{"istio-operator"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// CA private key
	caPrivKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	caPrivateKeyPEM = new(bytes.Buffer)
	_ = pem.Encode(caPrivateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})
	certificate.CAKey = caPrivateKeyPEM.Bytes()

	// Self signed CA certificate
	caBytes, err := x509.CreateCertificate(cryptorand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, err
	}

	// PEM encode CA cert
	caPEM = new(bytes.Buffer)
	_ = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	certificate.CACert = caPEM.Bytes()

	commonName := "unspecified"
	if len(dnsNames) > 0 {
		commonName = dnsNames[0]
	}

	// server cert config
	cert := &x509.Certificate{
		DNSNames:     dnsNames,
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"istio-operator"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	// server private key
	serverPrivKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	// sign the server cert
	serverCertBytes, err := x509.CreateCertificate(cryptorand.Reader, cert, ca, &serverPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, err
	}

	// PEM encode the  server cert and key
	serverCertPEM = new(bytes.Buffer)
	_ = pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})
	certificate.Cert = serverCertPEM.Bytes()

	serverPrivKeyPEM = new(bytes.Buffer)
	_ = pem.Encode(serverPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivKey),
	})
	certificate.Key = serverPrivKeyPEM.Bytes()

	err = certificate.Write(p.certDir)
	if err != nil {
		return nil, err
	}

	return &certificate, nil
}

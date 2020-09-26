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
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"path"
	"time"
)

const (
	// CAKeyName is the name of the CA private key
	CAKeyName = "ca.key"
	// CACertName is the name of the CA certificate
	CACertName = "ca.crt"
	// ServerKeyName is the name of the server private key
	ServerKeyName = "tls.key"
	// ServerCertName is the name of the serving certificate
	ServerCertName = "tls.crt"
)

type Certificate struct {
	// PEM encoded private key
	Key []byte
	// PEM encoded serving certificate
	Cert []byte
	// PEM encoded CA private key
	CAKey []byte
	// PEM encoded CA certificate
	CACert []byte
}

func ReadCertificate(dir string) (*Certificate, error) {
	return (&Certificate{}).Read(dir)
}

func (cert *Certificate) Read(dir string) (*Certificate, error) {
	filenames := []string{CAKeyName, CACertName, ServerCertName, ServerKeyName}
	for _, filename := range filenames {
		if !fileExists(path.Join(dir, filename)) {
			continue
		}

		content, err := ioutil.ReadFile(path.Join(dir, filename))
		if err != nil {
			return nil, err
		}

		switch filename {
		case CAKeyName:
			cert.CAKey = content
		case CACertName:
			cert.CACert = content
		case ServerCertName:
			cert.Cert = content
		case ServerKeyName:
			cert.Key = content
		}
	}

	return cert, nil
}

func (cert Certificate) Write(dir string) error {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	for file, content := range map[string][]byte{
		CAKeyName:      cert.CAKey,
		CACertName:     cert.CACert,
		ServerCertName: cert.Cert,
		ServerKeyName:  cert.Key,
	} {
		err = writeFile(dir+"/"+file, content)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cert *Certificate) Verify(dnsName string, checkTime time.Time) bool {
	if cert == nil || cert.Key == nil || cert.Cert == nil {
		return false
	}

	// Verify key and cert are valid pair
	_, err := tls.X509KeyPair(cert.Cert, cert.Key)
	if err != nil {
		return false
	}

	// Verify cert is good for desired DNS name and signed by CA and will be valid for desired period of time.
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(cert.CACert) {
		return false
	}
	block, _ := pem.Decode([]byte(cert.Cert))
	if block == nil {
		return false
	}
	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false
	}
	ops := x509.VerifyOptions{
		DNSName:     dnsName,
		Roots:       pool,
		CurrentTime: checkTime,
	}
	_, err = certificate.Verify(ops)
	return err == nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func writeFile(filepath string, content []byte) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(content)
	if err != nil {
		return err
	}
	return nil
}

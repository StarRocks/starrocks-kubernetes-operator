/*
Copyright 2021-present, StarRocks Inc.

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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

const (
	// DefaultCertDir is the default directory for webhook certificates
	DefaultCertDir = "/tmp/k8s-webhook-server/serving-certs"
	// DefaultServiceName is the default webhook service name
	DefaultServiceName = "webhook-service"
	// DefaultNamespace is the default namespace for webhook service
	DefaultNamespace = "default"
)

// GenerateSelfSignedCerts generates self-signed TLS certificates for webhook server
// The function will create certificates only if they don't already exist.
// Parameters:
// - certDir: directory to store certificates (uses DefaultCertDir if empty)
// - serviceName: webhook service name (uses DefaultServiceName if empty)
// - namespace: webhook service namespace (uses DefaultNamespace if empty)
// - validityDays: certificate validity period in days (defaults to 365 if <= 0)
// Returns the CA certificate in PEM format for use in ValidatingAdmissionWebhook
func GenerateSelfSignedCerts(certDir, serviceName, namespace string, validityDays int) ([]byte, error) {
	if certDir == "" {
		certDir = DefaultCertDir
	}
	if serviceName == "" {
		serviceName = DefaultServiceName
	}
	if namespace == "" {
		namespace = DefaultNamespace
	}
	if validityDays <= 0 {
		validityDays = 365 // Default to 1 year
	}

	// Create certificate directory
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create certificate directory: %v", err)
	}

	keyPath := filepath.Join(certDir, "tls.key")
	certPath := filepath.Join(certDir, "tls.crt")

	// Check if certificates already exist
	if _, err := os.Stat(keyPath); err == nil {
		if _, err := os.Stat(certPath); err == nil {
			// Both files exist, read and return the certificate
			certData, err := os.ReadFile(certPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read existing certificate: %v", err)
			}
			return certData, nil
		}
	}

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: fmt.Sprintf("%s.%s.svc", serviceName, namespace),
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Duration(validityDays) * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames: []string{
			serviceName,
			fmt.Sprintf("%s.%s", serviceName, namespace),
			fmt.Sprintf("%s.%s.svc", serviceName, namespace),
			fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, namespace),
			"localhost",
		},
		IPAddresses: []net.IP{
			net.IPv4(127, 0, 0, 1),
		},
	}

	// Generate certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %v", err)
	}

	// Write private key to file
	keyFile, err := os.Create(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create key file: %v", err)
	}
	defer keyFile.Close()

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	if err := pem.Encode(keyFile, privateKeyPEM); err != nil {
		return nil, fmt.Errorf("failed to write private key: %v", err)
	}

	// Write certificate to file
	certFile, err := os.Create(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate file: %v", err)
	}
	defer certFile.Close()

	certPEM := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	}
	if err := pem.Encode(certFile, certPEM); err != nil {
		return nil, fmt.Errorf("failed to write certificate: %v", err)
	}

	// Return the certificate in PEM format for caBundle
	return pem.EncodeToMemory(certPEM), nil
}

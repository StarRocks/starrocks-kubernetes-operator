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
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGenerateSelfSignedCerts(t *testing.T) {
	tests := []struct {
		name         string
		certDir      string
		serviceName  string
		namespace    string
		validityDays int
		wantErr      bool
	}{
		{
			name:         "valid with defaults",
			certDir:      "",
			serviceName:  "",
			namespace:    "",
			validityDays: 0,
			wantErr:      false,
		},
		{
			name:         "valid with custom values",
			certDir:      "",
			serviceName:  "custom-webhook",
			namespace:    "custom-namespace",
			validityDays: 30,
			wantErr:      false,
		},
		{
			name:         "valid with long validity",
			certDir:      "",
			serviceName:  "test-service",
			namespace:    "test-namespace",
			validityDays: 3650, // 10 years
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			certDir := tempDir
			if tt.certDir != "" {
				certDir = tt.certDir
			}

			caCert, err := GenerateSelfSignedCerts(certDir, tt.serviceName, tt.namespace, tt.validityDays)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSelfSignedCerts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && caCert == nil {
				t.Errorf("GenerateSelfSignedCerts() returned nil CA certificate")
				return
			}

			if !tt.wantErr {
				verifyCertificateGeneration(t, certDir, tt.serviceName, tt.namespace, tt.validityDays)
			}
		})
	}
}

func verifyCertificateGeneration(t *testing.T, certDir, serviceName, namespace string, validityDays int) {
	keyPath := filepath.Join(certDir, "tls.key")
	certPath := filepath.Join(certDir, "tls.crt")

	// Verify files exist
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Errorf("Private key file not created: %s", keyPath)
	}
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		t.Errorf("Certificate file not created: %s", certPath)
	}

	// Parse certificate
	cert := parseCertificateFromFile(t, certPath)
	if cert == nil {
		return
	}

	// Verify certificate properties
	verifyCertificateProperties(t, cert, serviceName, namespace, validityDays)
}

func parseCertificateFromFile(t *testing.T, certPath string) *x509.Certificate {
	certData, err := os.ReadFile(certPath)
	if err != nil {
		t.Errorf("Failed to read certificate file: %v", err)
		return nil
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		t.Error("Failed to decode certificate PEM")
		return nil
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Errorf("Failed to parse certificate: %v", err)
		return nil
	}

	return cert
}

func verifyCertificateProperties(t *testing.T, cert *x509.Certificate, serviceName, namespace string, validityDays int) {
	// Set defaults
	expectedServiceName := serviceName
	if expectedServiceName == "" {
		expectedServiceName = DefaultServiceName
	}
	expectedNamespace := namespace
	if expectedNamespace == "" {
		expectedNamespace = DefaultNamespace
	}

	// Verify CN
	expectedCN := fmt.Sprintf("%s.%s.svc", expectedServiceName, expectedNamespace)
	if cert.Subject.CommonName != expectedCN {
		t.Errorf("Certificate CN = %s, want %s", cert.Subject.CommonName, expectedCN)
	}

	// Verify validity period
	expectedValidityDays := validityDays
	if expectedValidityDays <= 0 {
		expectedValidityDays = 365
	}

	actualValidityDuration := cert.NotAfter.Sub(cert.NotBefore)
	expectedValidityDuration := time.Duration(expectedValidityDays) * 24 * time.Hour
	tolerance := time.Minute

	if actualValidityDuration < expectedValidityDuration-tolerance ||
		actualValidityDuration > expectedValidityDuration+tolerance {
		t.Errorf("Certificate validity duration = %v, want approximately %v",
			actualValidityDuration, expectedValidityDuration)
	}

	// Verify SAN entries
	expectedSANs := []string{
		expectedServiceName,
		fmt.Sprintf("%s.%s", expectedServiceName, expectedNamespace),
		fmt.Sprintf("%s.%s.svc", expectedServiceName, expectedNamespace),
		fmt.Sprintf("%s.%s.svc.cluster.local", expectedServiceName, expectedNamespace),
		"localhost",
	}

	for _, expectedSAN := range expectedSANs {
		found := false
		for _, actualSAN := range cert.DNSNames {
			if actualSAN == expectedSAN {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected SAN %s not found in certificate", expectedSAN)
		}
	}
}

func TestGenerateSelfSignedCerts_SkipsExistingCerts(t *testing.T) {
	tempDir := t.TempDir()

	// First generation
	_, err := GenerateSelfSignedCerts(tempDir, "test-service", "test-namespace", 30)
	if err != nil {
		t.Fatalf("First certificate generation failed: %v", err)
	}

	// Get file modification times
	keyPath := filepath.Join(tempDir, "tls.key")
	certPath := filepath.Join(tempDir, "tls.crt")

	keyInfo1, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("Failed to stat key file: %v", err)
	}

	certInfo1, err := os.Stat(certPath)
	if err != nil {
		t.Fatalf("Failed to stat cert file: %v", err)
	}

	// Wait a bit to ensure different modification time if files were recreated
	time.Sleep(10 * time.Millisecond)

	// Second generation should skip existing certificates
	_, err = GenerateSelfSignedCerts(tempDir, "test-service", "test-namespace", 60)
	if err != nil {
		t.Fatalf("Second certificate generation failed: %v", err)
	}

	// Verify files were not modified
	keyInfo2, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("Failed to stat key file after second generation: %v", err)
	}

	certInfo2, err := os.Stat(certPath)
	if err != nil {
		t.Fatalf("Failed to stat cert file after second generation: %v", err)
	}

	if !keyInfo1.ModTime().Equal(keyInfo2.ModTime()) {
		t.Error("Key file was modified when it should have been skipped")
	}

	if !certInfo1.ModTime().Equal(certInfo2.ModTime()) {
		t.Error("Certificate file was modified when it should have been skipped")
	}
}

func TestGenerateSelfSignedCerts_RegeneratesPartialCerts(t *testing.T) {
	tempDir := t.TempDir()

	keyPath := filepath.Join(tempDir, "tls.key")
	certPath := filepath.Join(tempDir, "tls.crt")

	// Create only key file (partial certificates)
	if err := os.WriteFile(keyPath, []byte("dummy key"), 0600); err != nil {
		t.Fatalf("Failed to create dummy key file: %v", err)
	}

	// Generation should create both files
	_, err := GenerateSelfSignedCerts(tempDir, "test-service", "test-namespace", 30)
	if err != nil {
		t.Fatalf("Certificate generation failed: %v", err)
	}

	// Verify both files exist and have valid content
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Error("Key file doesn't exist after generation")
	}

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		t.Error("Certificate file doesn't exist after generation")
	}

	// Verify certificate is valid (not the dummy content)
	certData, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatalf("Failed to read certificate: %v", err)
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		t.Error("Certificate is not valid PEM")
		return
	}

	_, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Errorf("Certificate is not valid X.509: %v", err)
	}
}

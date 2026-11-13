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

// Package mysqlutil holds the pieces of the operator's MySQL-to-FE connection handling that more
// than one sub controller needs. The CN controller talks to FE to drop compute nodes and
// warehouses; the FE controller talks to FE to find the leader during a leader-aware upgrade.
// Keeping the shared bits here avoids an import cycle between the two packages.
package mysqlutil

import (
	"fmt"
	"strings"
)

// SSL modes for the operator's MySQL connection to FE. The names mirror the MySQL client's
// --ssl-mode option so that the vocabulary in the FE SSL documentation carries over unchanged.
const (
	// SSLModeDisabled never negotiates TLS. An FE configured with ssl_force_secure_transport=true
	// rejects such a connection with error 5205.
	SSLModeDisabled = "DISABLED"

	// SSLModePreferred negotiates TLS when FE advertises the CLIENT_SSL capability and falls back
	// to plaintext when it does not. The server certificate is not verified.
	SSLModePreferred = "PREFERRED"

	// SSLModeRequired always negotiates TLS and fails the connection when FE does not advertise
	// CLIENT_SSL. The server certificate is not verified.
	SSLModeRequired = "REQUIRED"

	// SSLModeVerifyCA and SSLModeVerifyIdentity are recognized but rejected. The keystore FE's
	// documentation tells users to generate is self-signed, so there is no CA to chain to, and its
	// certificate name never matches the in-cluster service DNS name the operator dials.
	SSLModeVerifyCA       = "VERIFY_CA"
	SSLModeVerifyIdentity = "VERIFY_IDENTITY"
)

// NormalizeSSLMode makes a user-supplied mode comparable to the SSLMode* constants: matching is
// case-insensitive, and surrounding whitespace (easy to introduce through Helm values) is ignored.
func NormalizeSSLMode(mode string) string {
	return strings.ToUpper(strings.TrimSpace(mode))
}

// ValidateSSLMode reports whether mode is a value the operator can act on. main calls it right
// after flag.Parse so that a typo stops the operator at startup, instead of surfacing much later
// as a failed FE connection during reconcile.
func ValidateSSLMode(mode string) error {
	switch NormalizeSSLMode(mode) {
	case SSLModeDisabled, SSLModePreferred, SSLModeRequired:
		return nil
	case SSLModeVerifyCA, SSLModeVerifyIdentity:
		return fmt.Errorf("ssl mode %q is not supported: the keystore FE's documentation generates is "+
			"self-signed, so neither its certificate chain nor its certificate name can be verified; "+
			"use %s to require encryption without certificate verification", mode, SSLModeRequired)
	default:
		return fmt.Errorf("invalid ssl mode %q, must be one of %s, %s, %s",
			mode, SSLModeDisabled, SSLModePreferred, SSLModeRequired)
	}
}

// SSLModeDSNSuffix returns the query string to append to the FE DSN for mode, including the leading
// "?" when non-empty. The values map onto go-sql-driver's tls= parameter: "preferred" negotiates TLS
// but falls back to plaintext when the server does not advertise CLIENT_SSL, while "skip-verify"
// insists on TLS. Neither verifies the server certificate.
//
// Unrecognized values yield "" — the plaintext behavior that predates this feature — rather than a
// panic. Production never reaches that branch because main rejects such values via ValidateSSLMode.
func SSLModeDSNSuffix(mode string) string {
	switch NormalizeSSLMode(mode) {
	case SSLModePreferred:
		return "?tls=preferred"
	case SSLModeRequired:
		return "?tls=skip-verify"
	default:
		return ""
	}
}

// DSN builds the go-sql-driver DSN for a root connection to the FE query port.
func DSN(rootPassword, host, port, sslMode string) string {
	return fmt.Sprintf("root:%s@tcp(%s:%s)/%s", rootPassword, host, port, SSLModeDSNSuffix(sslMode))
}

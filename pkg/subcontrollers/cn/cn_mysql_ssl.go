package cn

import "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/mysqlutil"

// SSL modes for the operator's MySQL connection to FE. The definitions live in mysqlutil so that
// the FE controller can share them without importing this package; these aliases keep the names
// callers already use (main, tests) stable.
const (
	SSLModeDisabled       = mysqlutil.SSLModeDisabled
	SSLModePreferred      = mysqlutil.SSLModePreferred
	SSLModeRequired       = mysqlutil.SSLModeRequired
	SSLModeVerifyCA       = mysqlutil.SSLModeVerifyCA
	SSLModeVerifyIdentity = mysqlutil.SSLModeVerifyIdentity
)

// ValidateSSLMode reports whether mode is a value the operator can act on. See mysqlutil.ValidateSSLMode.
func ValidateSSLMode(mode string) error {
	return mysqlutil.ValidateSSLMode(mode)
}

// sslModeDSNSuffix returns the DSN query suffix for mode. See mysqlutil.SSLModeDSNSuffix.
func sslModeDSNSuffix(mode string) string {
	return mysqlutil.SSLModeDSNSuffix(mode)
}

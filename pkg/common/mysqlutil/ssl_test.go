package mysqlutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDSN(t *testing.T) {
	require.Equal(t, "root:secret@tcp(sr-fe-service.ns:9030)/?tls=preferred",
		DSN("secret", "sr-fe-service.ns", "9030", SSLModePreferred))
	require.Equal(t, "root:@tcp(sr-fe-service.ns:9030)/?tls=skip-verify",
		DSN("", "sr-fe-service.ns", "9030", "required"))
	require.Equal(t, "root:@tcp(sr-fe-service.ns:9030)/",
		DSN("", "sr-fe-service.ns", "9030", SSLModeDisabled))
}

func TestValidateSSLMode(t *testing.T) {
	require.NoError(t, ValidateSSLMode(" preferred "))
	require.Error(t, ValidateSSLMode(SSLModeVerifyCA))
	require.Error(t, ValidateSSLMode("nope"))
}

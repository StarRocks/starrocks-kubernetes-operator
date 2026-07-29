package cn

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSLModeDSNSuffix(t *testing.T) {
	tests := []struct {
		name string
		mode string
		want string
	}{
		{name: "disabled adds nothing", mode: SSLModeDisabled, want: ""},
		{name: "preferred negotiates with plaintext fallback", mode: SSLModePreferred, want: "?tls=preferred"},
		{name: "required always negotiates", mode: SSLModeRequired, want: "?tls=skip-verify"},
		{name: "lower case is accepted", mode: "preferred", want: "?tls=preferred"},
		{name: "mixed case is accepted", mode: "ReQuIrEd", want: "?tls=skip-verify"},
		{name: "surrounding spaces are trimmed", mode: "  PREFERRED  ", want: "?tls=preferred"},
		{name: "verify_ca falls back to plaintext", mode: SSLModeVerifyCA, want: ""},
		{name: "verify_identity falls back to plaintext", mode: SSLModeVerifyIdentity, want: ""},
		{name: "unrecognized value falls back to plaintext", mode: "banana", want: ""},
		{name: "empty value falls back to plaintext", mode: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, sslModeDSNSuffix(tt.mode))
		})
	}
}

func TestValidateSSLMode(t *testing.T) {
	for _, mode := range []string{SSLModeDisabled, SSLModePreferred, SSLModeRequired, "preferred", "  REQUIRED "} {
		t.Run("accepts "+mode, func(t *testing.T) {
			assert.NoError(t, ValidateSSLMode(mode))
		})
	}

	// VERIFY_CA / VERIFY_IDENTITY are recognized names, so the error must explain why they are
	// unusable and point at REQUIRED instead of just listing valid values.
	for _, mode := range []string{SSLModeVerifyCA, SSLModeVerifyIdentity} {
		t.Run("rejects "+mode, func(t *testing.T) {
			err := ValidateSSLMode(mode)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "self-signed")
			assert.Contains(t, err.Error(), SSLModeRequired)
		})
	}

	for _, mode := range []string{"", "banana", "true"} {
		t.Run("rejects unrecognized "+mode, func(t *testing.T) {
			err := ValidateSSLMode(mode)
			require.Error(t, err)
			assert.Contains(t, err.Error(), SSLModePreferred)
		})
	}
}

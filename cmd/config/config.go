package config

import "fmt"

var (
	// DNSDomainSuffix is the suffix of the dns domain in k8s
	DNSDomainSuffix string

	// VolumeNameWithHash decides whether adding a hash to the volume name
	VolumeNameWithHash bool

	// WebhookCertValidityDays is the validity period in days for self-signed certificates
	WebhookCertValidityDays int
)

func GetServiceDomainSuffix() string {
	return fmt.Sprintf("svc.%s", DNSDomainSuffix)
}

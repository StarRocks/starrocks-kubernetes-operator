package config

import "fmt"

var (
	// DNSDomainSuffix is the suffix of the dns domain in k8s
	DNSDomainSuffix string

	// VolumeNameWithHash decides whether adding a hash to the volume name
	VolumeNameWithHash bool
)

func GetServiceDomainSuffix() string {
	return fmt.Sprintf("svc.%s", DNSDomainSuffix)
}

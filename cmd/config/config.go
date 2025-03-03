package config

import "fmt"

var DNSDomainSuffix string

func GetServiceDomainSuffix() string {
	return fmt.Sprintf("svc.%s", DNSDomainSuffix)
}

var VolumeNameWithHash bool

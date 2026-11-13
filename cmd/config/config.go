package config

import "fmt"

var (
	// DNSDomainSuffix is the suffix of the dns domain in k8s
	DNSDomainSuffix string

	// VolumeNameWithHash decides whether adding a hash to the volume name
	VolumeNameWithHash bool

	// FeSslMode decides whether the operator's MySQL connection to FE negotiates TLS. Valid values
	// are DISABLED, PREFERRED and REQUIRED — the MySQL client's --ssl-mode vocabulary, matched
	// case-insensitively. It is validated once at startup (see cn.ValidateSSLMode), applies to the
	// whole operator process, and requires an operator restart to change.
	FeSslMode string

	// EnableFeLeaderAwareUpgrade makes the operator drive FE pod replacement itself instead of
	// leaving it to the StatefulSet controller. When true, the FE StatefulSet is forced to the
	// OnDelete update strategy and, on every template change, the operator deletes one FE pod at
	// a time, followers first and the BDBJE leader (as reported by SHOW FRONTENDS) last, so that
	// a single upgrade triggers exactly one leader election. See fe.ReconcileLeaderAwareUpgrade.
	EnableFeLeaderAwareUpgrade bool
)

func GetServiceDomainSuffix() string {
	return fmt.Sprintf("svc.%s", DNSDomainSuffix)
}

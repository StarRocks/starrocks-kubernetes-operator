package v1

// the labels key
const (
	// ComponentLabelKey is Kubernetes recommended label key, it represents the component within the architecture
	ComponentLabelKey string = "app.kubernetes.io/component"

	// OwnerReference represents  owner of the object
	OwnerReference string = "app.celerdata.ownerreference/name"

	// ComponentResourceHash the component hash
	ComponentResourceHash string = "app.celerdata.components/hash"
)

// the labels value. default statefulset name
const (
	DEFAULT_FE       = "fe"
	DEFAULT_BE       = "be"
	DEFAULT_CN       = "cn"
	DEFAULT_FE_PROXY = "fe-proxy"
)

// the env of container
const (
	COMPONENT_NAME  = "COMPONENT_NAME"
	FE_SERVICE_NAME = "FE_SERVICE_NAME"
)

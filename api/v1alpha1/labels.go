package v1alpha1

//the labels key
const (
	// ComponentLabelKey is Kubernetes recommended label key, it represents the component within the architecture
	ComponentLabelKey string = "app.kubernetes.io/component"
	// NameLabelKey is Kubernetes recommended label key, it represents the name of the application
	NameLabelKey string = "app.kubernetes.io/name"

	//OwnerReference list object depended by this object
	OwnerReference string = "app.starrocks.ownerreference/name"

	//ComponentsResourceHash the component hash
	ComponentResourceHash string = "app.starrocks.components/hash"

	//ComponentGeneration record for last update generation for compare with new spec.
	ComponentGeneration string = "app.starrocks.components/generation"
)

//the labels value. default statefulset name
const (
	DEFAULT_FE = "fe"
	DEFAULT_BE = "be"
	DEFAULT_CN = "cn"
)

//config value
const (
	DEFAULT_FE_CONFIG_NAME = "fe-config"
	//TODO: 待指定
	DEFAULT_FE_CONFIG_PATH = ""

	DEFAULT_START_SCRIPT_NAME = "fe-start-script"

	DEFAULT_START_SCRIPT_PATH = ""

	DEFAULT_FE_SERVICE_NAME = "starrocks-fe-service"

	DEFAULT_BE_SERVICE_NAME = "starrocks-be-service"

	DEFAULT_CN_SERVICE_NAME = "starrocks-cn-service"
)

//the env of container
const (
	COMPONENT_NAME = "COMPONENT_NAME"
	SERVICE_NAME   = "SERVICE_NAME"
)

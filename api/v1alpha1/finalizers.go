package v1alpha1

const (
	//FE_STATEFULSET_FINALIZER pre hook wait for fe statefulset deleted.
	FE_STATEFULSET_FINALIZER = "starrocks.com/fe.statefulset/finalizer"

	//BE_STATEFULSET_FINALIZER pre hook wait for be statefulset deleted.
	BE_STATEFULSET_FINALIZER = "starrocks.com/be.statefulset/finalizer"

	//CN_STATEFULSET_FINALIZER pre hook wait for cn statefulset deleted.
	CN_STATEFULSET_FINALIZER = "starrocks.com/cn.statefulset/finalizer"

	FE_SERVICE_FINALIZER = "starrocks.com/fe.service/finalizer"

	BE_SERVICE_FINALIZER = "starrocks.com/be.service/finalizer"

	CN_SERVICE_FINALIZER = "starrocks.com/cn.service/finalizer"
)

var ResourceTypeFinalizers map[string]string

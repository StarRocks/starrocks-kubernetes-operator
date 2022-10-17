package pkg

import ctrl "sigs.k8s.io/controller-runtime"

var (
	//Controllers through the init for add Controller.
	Controllers []Controller
)

type Controller interface {
	Init(mgr ctrl.Manager)
}

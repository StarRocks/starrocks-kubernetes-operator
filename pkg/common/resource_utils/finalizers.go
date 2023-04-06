package resource_utils

type Finalizers []string

func (fs Finalizers) HaveFinalizer(finalizer string) bool {
	for _, v := range fs {
		if finalizer == v {
			return true
		}
	}
	return false
}

func (fs Finalizers) AddFinalizer(finalizer string) Finalizers {
	for _, f := range fs {
		if f == finalizer {
			return fs
		}
	}

	return append(fs, finalizer)
}

func (fs Finalizers) DeleteFinalizer(finalizer string) Finalizers {
	index := -1
	for i, _ := range fs {
		if fs[i] == finalizer {
			index = i
			break
		}
	}

	if index == -1 {
		return fs
	}

	pref := fs[0:index]
	onef := append(pref, fs[index+1:]...)
	return onef.DeleteFinalizer(finalizer)
}

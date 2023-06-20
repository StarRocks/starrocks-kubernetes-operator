package pod

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	HEALTH_API_PATH = "/api/health"
)

// MakeStartupProbe returns a startup probe.
func MakeStartupProbe(port int32, path string) *corev1.Probe {
	return &corev1.Probe{
		FailureThreshold: 60,
		PeriodSeconds:    5,
		ProbeHandler:     makeProbeHandler(port, path),
	}
}

// MakeLivenessProbe returns a liveness.
func MakeLivenessProbe(port int32, path string) *corev1.Probe {
	return &corev1.Probe{
		PeriodSeconds:    5,
		FailureThreshold: 3,
		ProbeHandler:     makeProbeHandler(port, path),
	}
}

// MakeReadinessProbe returns a readiness probe.
func MakeReadinessProbe(port int32, path string) *corev1.Probe {
	return &corev1.Probe{
		PeriodSeconds:    5,
		FailureThreshold: 3,
		ProbeHandler:     makeProbeHandler(port, path),
	}
}

// MakeLifeCycle returns a lifecycle.
func MakeLifeCycle(preStopScriptPath string) *corev1.Lifecycle {
	return &corev1.Lifecycle{
		PreStop: &corev1.LifecycleHandler{
			Exec: &corev1.ExecAction{
				Command: []string{preStopScriptPath},
			},
		},
	}
}

func makeProbeHandler(port int32, path string) corev1.ProbeHandler {
	return corev1.ProbeHandler{
		HTTPGet: &corev1.HTTPGetAction{
			Path: path,
			Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: port,
			},
		},
	}
}

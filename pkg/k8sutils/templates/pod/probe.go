package pod

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// StartupProbe returns a startup probe.
func StartupProbe(startupProbeFailureSeconds *int32, port int32, path string) *corev1.Probe {
	var defaultFailureThreshold int32 = 60
	var defaultPeriodSeconds int32 = 5
	return completeProbe(startupProbeFailureSeconds, defaultFailureThreshold, defaultPeriodSeconds, getProbe(port, path))
}

// LivenessProbe returns a liveness probe.
func LivenessProbe(livenessProbeFailureSeconds *int32, port int32, path string) *corev1.Probe {
	var defaultFailureThreshold int32 = 3
	var defaultPeriodSeconds int32 = 5
	return completeProbe(livenessProbeFailureSeconds, defaultFailureThreshold, defaultPeriodSeconds, getProbe(port, path))
}

// ReadinessProbe returns a readiness probe.
func ReadinessProbe(readinessProbeFailureSeconds *int32, port int32, path string) *corev1.Probe {
	var defaultFailureThreshold int32 = 3
	var defaultPeriodSeconds int32 = 5
	return completeProbe(readinessProbeFailureSeconds, defaultFailureThreshold, defaultPeriodSeconds, getProbe(port, path))
}

func completeProbe(failureSeconds *int32, defaultFailureThreshold int32, defaultPeriodSeconds int32,
	probeHandler corev1.ProbeHandler) *corev1.Probe {
	probe := &corev1.Probe{}
	if failureSeconds != nil && *failureSeconds > 0 {
		probe.FailureThreshold = (*failureSeconds + defaultPeriodSeconds - 1) / defaultPeriodSeconds
	} else {
		probe.FailureThreshold = defaultFailureThreshold
	}
	probe.PeriodSeconds = defaultPeriodSeconds
	probe.ProbeHandler = probeHandler
	return probe
}

func getProbe(port int32, path string) corev1.ProbeHandler {
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

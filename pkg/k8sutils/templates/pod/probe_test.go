package pod

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestMakeLivenessProbe(t *testing.T) {
	type args struct {
		seconds *int32
		port    int32
		path    string
	}
	tests := []struct {
		name string
		args args
		want *corev1.Probe
	}{
		{
			name: "liveness probe with default seconds",
			args: args{
				port: 8080,
				path: "/api/health2",
			},
			want: &corev1.Probe{
				PeriodSeconds:    5,
				FailureThreshold: 3,
				ProbeHandler:     getProbe(8080, "/api/health2"),
			},
		},
		{
			name: "liveness probe with specified seconds",
			args: args{
				seconds: func() *int32 { s := int32(50); return &s }(),
				port:    8080,
				path:    "/api/health2",
			},
			want: &corev1.Probe{
				PeriodSeconds:    5,
				FailureThreshold: 10,
				ProbeHandler:     getProbe(8080, "/api/health2"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LivenessProbe(tt.args.seconds, tt.args.port, tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LivenessProbe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeReadinessProbe(t *testing.T) {
	type args struct {
		seconds *int32
		port    int32
		path    string
	}
	tests := []struct {
		name string
		args args
		want *corev1.Probe
	}{
		{
			name: "readiness probe with default seconds",
			args: args{
				port: 8080,
				path: "/api/health2",
			},
			want: &corev1.Probe{
				PeriodSeconds:    5,
				FailureThreshold: 3,
				ProbeHandler:     getProbe(8080, "/api/health2"),
			},
		},
		{
			name: "readiness probe with specified seconds",
			args: args{
				seconds: func() *int32 { s := int32(50); return &s }(),
				port:    8080,
				path:    "/api/health2",
			},
			want: &corev1.Probe{
				PeriodSeconds:    5,
				FailureThreshold: 10,
				ProbeHandler:     getProbe(8080, "/api/health2"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReadinessProbe(tt.args.seconds, tt.args.port, tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadinessProbe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeStartupProbe(t *testing.T) {
	type args struct {
		port int32
		path string
	}
	tests := []struct {
		name string
		args args
		want *corev1.Probe
	}{
		{
			name: "test",
			args: args{
				port: 8080,
				path: "/api/health2",
			},
			want: &corev1.Probe{
				FailureThreshold: 60,
				PeriodSeconds:    5,
				ProbeHandler:     getProbe(8080, "/api/health2"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StartupProbe(nil, tt.args.port, tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StartupProbe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_makeProbeHandler(t *testing.T) {
	type args struct {
		port int32
		path string
	}
	tests := []struct {
		name string
		args args
		want corev1.ProbeHandler
	}{
		{
			name: "test",
			args: args{
				port: 8080,
				path: "/api/health2",
			},
			want: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/api/health2",
					Port: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 8080,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getProbe(tt.args.port, tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getProbe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_completeProbe(t *testing.T) {
	type args struct {
		originalProbe           *int32
		defaultFailureThreshold int32
		defaultPeriodSeconds    int32
		probeHandler            corev1.ProbeHandler
	}
	tests := []struct {
		name string
		args args
		want *corev1.Probe
	}{
		{
			name: "test complete probe",
			args: args{
				originalProbe:           nil,
				defaultFailureThreshold: 1,
				defaultPeriodSeconds:    1,
				probeHandler:            corev1.ProbeHandler{},
			},
			want: &corev1.Probe{
				ProbeHandler:     corev1.ProbeHandler{},
				FailureThreshold: 1,
				PeriodSeconds:    1,
			},
		},
		{
			name: "test complete probe 2",
			args: args{
				originalProbe:           func() *int32 { v := int32(10); return &v }(),
				defaultFailureThreshold: 1,
				defaultPeriodSeconds:    5,
				probeHandler:            corev1.ProbeHandler{},
			},
			want: &corev1.Probe{
				ProbeHandler:     corev1.ProbeHandler{},
				FailureThreshold: 2,
				PeriodSeconds:    5,
			},
		},
		{
			name: "test complete probe 3",
			args: args{
				originalProbe:           func() *int32 { v := int32(0); return &v }(),
				defaultFailureThreshold: 60,
				defaultPeriodSeconds:    5,
				probeHandler:            corev1.ProbeHandler{},
			},
			want: &corev1.Probe{
				ProbeHandler:     corev1.ProbeHandler{},
				FailureThreshold: 60,
				PeriodSeconds:    5,
			},
		},
		{
			name: "test complete probe 4",
			args: args{
				originalProbe:           func() *int32 { v := int32(1); return &v }(),
				defaultFailureThreshold: 60,
				defaultPeriodSeconds:    5,
				probeHandler:            corev1.ProbeHandler{},
			},
			want: &corev1.Probe{
				ProbeHandler:     corev1.ProbeHandler{},
				FailureThreshold: 1,
				PeriodSeconds:    5,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := completeProbe(tt.args.originalProbe, tt.args.defaultFailureThreshold,
				tt.args.defaultPeriodSeconds, tt.args.probeHandler); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("completeProbe() = %v, want %v", got, tt.want)
			}
		})
	}
}

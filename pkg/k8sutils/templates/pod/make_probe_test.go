package pod

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
	"testing"
)

func TestMakeLivenessProbe(t *testing.T) {
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
				PeriodSeconds:    5,
				FailureThreshold: 3,
				ProbeHandler:     makeProbeHandler(8080, "/api/health2"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MakeLivenessProbe(tt.args.port, tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeLivenessProbe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeReadinessProbe(t *testing.T) {
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
				PeriodSeconds:    5,
				FailureThreshold: 3,
				ProbeHandler:     makeProbeHandler(8080, "/api/health2"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MakeReadinessProbe(tt.args.port, tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeReadinessProbe() = %v, want %v", got, tt.want)
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
				ProbeHandler:     makeProbeHandler(8080, "/api/health2"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MakeStartupProbe(tt.args.port, tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeStartupProbe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeLifeCycle(t *testing.T) {
	type args struct {
		preStopScriptPath string
	}
	tests := []struct {
		name string
		args args
		want *corev1.Lifecycle
	}{
		{
			name: "test",
			args: args{
				preStopScriptPath: "/scripts/pre-stop.sh",
			},
			want: &corev1.Lifecycle{
				PreStop: &corev1.LifecycleHandler{
					Exec: &corev1.ExecAction{
						Command: []string{"/scripts/pre-stop.sh"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := MakeLifeCycle(tt.args.preStopScriptPath)
			if !reflect.DeepEqual(actual, tt.want) {
				t.Errorf("MakeLifeCycle() = %v, want %v", actual, tt.want)
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
			if got := makeProbeHandler(tt.args.port, tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeProbeHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

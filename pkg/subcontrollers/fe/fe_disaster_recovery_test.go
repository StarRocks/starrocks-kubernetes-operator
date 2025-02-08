package fe

import (
	"reflect"
	"testing"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
)

func TestShouldEnterDisasterRecoveryMode(t *testing.T) {
	type args struct {
		feSpec   *v1.StarRocksFeSpec
		feStatus *v1.StarRocksFeStatus
		config   map[string]interface{}
	}
	tests := []struct {
		name     string
		args     args
		want     bool
		wantPort int32
	}{
		{
			name: "cluster data is not shared data mode",
			args: args{
				feSpec: &v1.StarRocksFeSpec{
					StarRocksComponentSpec: v1.StarRocksComponentSpec{
						DisasterRecovery: &v1.DisasterRecovery{
							Enabled:    true,
							Generation: 1,
						},
					},
				},
				feStatus: nil,
				config: map[string]interface{}{
					"run_mode": "shared_nothing",
				},
			},
			want: false,
		},
		{
			name: "disaster recovery field is not present",
			args: args{
				feSpec: &v1.StarRocksFeSpec{
					StarRocksComponentSpec: v1.StarRocksComponentSpec{},
				},
				feStatus: nil,
				config: map[string]interface{}{
					"run_mode": "shared_data",
				},
			},
			want: false,
		},
		{
			name: "disaster recovery is not enabled",
			args: args{
				feSpec: &v1.StarRocksFeSpec{
					StarRocksComponentSpec: v1.StarRocksComponentSpec{
						DisasterRecovery: &v1.DisasterRecovery{
							Enabled: false,
						},
					},
				},
				feStatus: nil,
				config: map[string]interface{}{
					"run_mode": "shared_data",
				},
			},
			want: false,
		},
		{
			name: "observed generation equals to generation",
			args: args{
				feSpec: &v1.StarRocksFeSpec{
					StarRocksComponentSpec: v1.StarRocksComponentSpec{
						DisasterRecovery: &v1.DisasterRecovery{
							Enabled: true,
						},
					},
				},
				feStatus: &v1.StarRocksFeStatus{
					StarRocksComponentStatus: v1.StarRocksComponentStatus{
						DisasterRecovery: &v1.DisasterRecoveryStatus{
							Phase:              v1.DRPhaseDone,
							ObservedGeneration: 0,
						},
					},
				},
				config: map[string]interface{}{
					"run_mode": "shared_data",
				},
			},
			want: false,
		},
		{
			name: "observed generation is larger than generation",
			args: args{
				feSpec: &v1.StarRocksFeSpec{
					StarRocksComponentSpec: v1.StarRocksComponentSpec{
						DisasterRecovery: &v1.DisasterRecovery{
							Enabled:    true,
							Generation: 1,
						},
					},
				},
				feStatus: &v1.StarRocksFeStatus{
					StarRocksComponentStatus: v1.StarRocksComponentStatus{
						DisasterRecovery: &v1.DisasterRecoveryStatus{
							ObservedGeneration: 2,
						},
					},
				},
				config: map[string]interface{}{
					"run_mode": "shared_data",
				},
			},
			want: false,
		},

		// add test cases that it should enter disaster recovery mode

		{
			name: "fe status is nil",
			args: args{
				feSpec: &v1.StarRocksFeSpec{
					StarRocksComponentSpec: v1.StarRocksComponentSpec{
						DisasterRecovery: &v1.DisasterRecovery{
							Enabled: true,
						},
					},
				},
				feStatus: nil,
				config: map[string]interface{}{
					"run_mode": "shared_data",
				},
			},
			want:     true,
			wantPort: 9030,
		},
		{
			name: "generation is larger than observed generation",
			args: args{
				feSpec: &v1.StarRocksFeSpec{
					StarRocksComponentSpec: v1.StarRocksComponentSpec{
						DisasterRecovery: &v1.DisasterRecovery{
							Enabled:    true,
							Generation: 1,
						},
					},
				},
				feStatus: &v1.StarRocksFeStatus{
					StarRocksComponentStatus: v1.StarRocksComponentStatus{
						DisasterRecovery: &v1.DisasterRecoveryStatus{
							ObservedGeneration: 0,
						},
					},
				},
				config: map[string]interface{}{
					"run_mode": "shared_data",
				},
			},
			want:     true,
			wantPort: 9030,
		},
		{
			name: "generation equal to observed generation, and phase is todo",
			args: args{
				feSpec: &v1.StarRocksFeSpec{
					StarRocksComponentSpec: v1.StarRocksComponentSpec{
						DisasterRecovery: &v1.DisasterRecovery{
							Enabled:    true,
							Generation: 1,
						},
					},
				},
				feStatus: &v1.StarRocksFeStatus{
					StarRocksComponentStatus: v1.StarRocksComponentStatus{
						DisasterRecovery: &v1.DisasterRecoveryStatus{
							Phase:              v1.DRPhaseTodo,
							ObservedGeneration: 1,
						},
					},
				},
				config: map[string]interface{}{
					"run_mode": "shared_data",
				},
			},
			want:     true,
			wantPort: 9030,
		},
		{
			name: "generation equal to observed generation, and phase is doing",
			args: args{
				feSpec: &v1.StarRocksFeSpec{
					StarRocksComponentSpec: v1.StarRocksComponentSpec{
						DisasterRecovery: &v1.DisasterRecovery{
							Enabled:    true,
							Generation: 1,
						},
					},
				},
				feStatus: &v1.StarRocksFeStatus{
					StarRocksComponentStatus: v1.StarRocksComponentStatus{
						DisasterRecovery: &v1.DisasterRecoveryStatus{
							Phase:              v1.DRPhaseDoing,
							ObservedGeneration: 1,
						},
					},
				},
				config: map[string]interface{}{
					"run_mode": "shared_data",
				},
			},
			want:     true,
			wantPort: 9030,
		},
		{
			name: "generation equal to observed generation, and phase is doing",
			args: args{
				feSpec: &v1.StarRocksFeSpec{
					StarRocksComponentSpec: v1.StarRocksComponentSpec{
						DisasterRecovery: &v1.DisasterRecovery{
							Enabled:    true,
							Generation: 1,
						},
					},
				},
				feStatus: &v1.StarRocksFeStatus{
					StarRocksComponentStatus: v1.StarRocksComponentStatus{
						DisasterRecovery: &v1.DisasterRecoveryStatus{
							Phase:              v1.DRPhaseDoing,
							ObservedGeneration: 1,
						},
					},
				},
				config: map[string]interface{}{
					"run_mode":   "shared_data",
					"query_port": "9090",
				},
			},
			want:     true,
			wantPort: 9090,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := ShouldEnterDisasterRecoveryMode(tt.args.feSpec, tt.args.feStatus, tt.args.config); got != tt.want {
				t.Errorf("ShouldEnterDisasterRecoveryMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRewriteStatefulSetForDisasterRecovery(t *testing.T) {
	type args struct {
		sts                *appv1.StatefulSet
		observedGeneration int64
	}
	tests := []struct {
		name string
		args args
		want *appv1.StatefulSet
	}{
		{
			name: "test rewrite statefulset for disaster recovery",
			args: args{
				sts: &appv1.StatefulSet{
					Spec: appv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										LivenessProbe:  &corev1.Probe{},
										ReadinessProbe: &corev1.Probe{},
										StartupProbe:   &corev1.Probe{},
										Env: []corev1.EnvVar{
											{
												Name:  "key",
												Value: "value",
											},
										},
									},
								},
							},
						},
					},
				},
				observedGeneration: 1,
			},
			want: &appv1.StatefulSet{
				Spec: appv1.StatefulSetSpec{
					Replicas: func(i int32) *int32 { return &i }(1),
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									LivenessProbe:  nil,
									ReadinessProbe: PortReadyProbe(9030),
									StartupProbe:   nil,
									Env: []corev1.EnvVar{
										{
											Name:  "key",
											Value: "value",
										},
										{
											Name:  "RESTORE_CLUSTER_GENERATION",
											Value: "1",
										},
										{
											Name:  "RESTORE_CLUSTER_SNAPSHOT",
											Value: "true",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rewriteStatefulSetForDisasterRecovery(tt.args.sts, tt.args.observedGeneration, 9030); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rewriteStatefulSetForDisasterRecovery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hasClusterSnapshotConf(t *testing.T) {
	type args struct {
		configMaps []v1.ConfigMapReference
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "has cluster snapshot conf",
			args: args{
				configMaps: []v1.ConfigMapReference{
					{
						SubPath: "cluster_snapshot.yaml",
					},
				},
			},
			want: true,
		},
		{
			name: "has cluster snapshot conf",
			args: args{
				configMaps: []v1.ConfigMapReference{
					{
						MountPath: "fe/conf",
					},
				},
			},
			want: true,
		},
		{
			name: "no cluster snapshot conf",
			args: args{
				configMaps: []v1.ConfigMapReference{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasClusterSnapshotConf(tt.args.configMaps); got != tt.want {
				t.Errorf("hasClusterSnapshotConf() = %v, want %v", got, tt.want)
			}
		})
	}
}

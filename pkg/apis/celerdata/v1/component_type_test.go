package v1

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestValidUpdateStrategy(t *testing.T) {
	type args struct {
		updateStrategy *appsv1.StatefulSetUpdateStrategy
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Validating update strategy with valid max unavailable",
			args: args{
				updateStrategy: &appsv1.StatefulSetUpdateStrategy{
					RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
						MaxUnavailable: func() *intstr.IntOrString {
							i := intstr.FromInt(1)
							return &i
						}(),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Validating update strategy with max unavailable 0",
			args: args{
				updateStrategy: &appsv1.StatefulSetUpdateStrategy{
					RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
						MaxUnavailable: func() *intstr.IntOrString {
							i := intstr.FromInt(0)
							return &i
						}(),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Validating update strategy with max unavailable 0%",
			args: args{
				updateStrategy: &appsv1.StatefulSetUpdateStrategy{
					RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
						MaxUnavailable: func() *intstr.IntOrString {
							i := intstr.FromString("0%")
							return &i
						}(),
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidUpdateStrategy(tt.args.updateStrategy); (err != nil) != tt.wantErr {
				t.Errorf("ValidUpdateStrategy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

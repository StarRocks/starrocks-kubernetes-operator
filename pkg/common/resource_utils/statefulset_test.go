package resource_utils

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestStatefulSetDeepEqual(t *testing.T) {
	type args struct {
		new             *appsv1.StatefulSet
		old             *appsv1.StatefulSet
		excludeReplicas bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "equal",
			args: args{
				new: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
				},
				old: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
				},
			},
			want: true,
		},
		{
			name: "not equal because of finalizer",
			args: args{
				new: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Finalizers: []string{"test"},
					},
				},
				old: &appsv1.StatefulSet{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StatefulSetDeepEqual(tt.args.new, tt.args.old, tt.args.excludeReplicas); got != tt.want {
				t.Errorf("StatefulSetDeepEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

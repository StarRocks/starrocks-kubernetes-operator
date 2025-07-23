package resource_utils

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/hash"
)

func TestStatefulSetDeepEqual(t *testing.T) {
	type args struct {
		expect *appsv1.StatefulSet
		actual *appsv1.StatefulSet
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "equal by both calculating the hash",
			args: args{
				expect: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
				},
				actual: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
				},
			},
			want: true,
		},
		{
			name: "equal by getting hash value from annotations",
			args: args{
				expect: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
				},
				actual: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							v1.ComponentResourceHash: "300056134",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "not equal because of finalizer",
			args: args{
				expect: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Finalizers: []string{"test"},
					},
				},
				actual: &appsv1.StatefulSet{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beginHash := hash.HashObject(tt.args.expect)
			if hash, got := StatefulSetDeepEqual(tt.args.expect, tt.args.actual); got != tt.want {
				t.Errorf("StatefulSetDeepEqual() = %v, want %v, expected hash: %v", got, tt.want, hash)
			}
			endHash := hash.HashObject(tt.args.expect)
			if beginHash != endHash {
				t.Errorf("StatefulSetDeepEqual() changed the expected Statefulset object, expected: %v, got: %v", beginHash, endHash)
			}
		})
	}
}

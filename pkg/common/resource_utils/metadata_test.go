package resource_utils

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMergeMetadata(t *testing.T) {
	type args struct {
		new *metav1.ObjectMeta
		old metav1.ObjectMeta
	}
	tests := []struct {
		name string
		args args
		want *metav1.ObjectMeta
	}{
		{
			name: "test",
			args: args{
				new: &metav1.ObjectMeta{
					Labels: map[string]string{
						"new":     "new",
						"exiting": "new",
					},
					Annotations: Annotations{
						"new":     "new",
						"exiting": "new",
					},
					Finalizers: []string{
						"new", "exiting",
					},
					OwnerReferences: []metav1.OwnerReference{{Name: "new"}, {Name: "exiting"}},
				},
				old: metav1.ObjectMeta{
					Labels: map[string]string{
						"old":     "old",
						"exiting": "old",
					},
					Annotations: Annotations{
						"old":     "old",
						"exiting": "old",
					},
					Finalizers:      []string{"old", "exiting"},
					OwnerReferences: []metav1.OwnerReference{{Name: "old"}, {Name: "exiting"}},
				},
			},
			want: &metav1.ObjectMeta{
				Labels: map[string]string{
					"new":     "new",
					"exiting": "new",
					"old":     "old",
				},
				Annotations: Annotations{
					"new":     "new",
					"exiting": "new",
					"old":     "old",
				},
				Finalizers: []string{
					"new", "exiting", "old",
				},
				OwnerReferences: []metav1.OwnerReference{{Name: "new"}, {Name: "exiting"}, {Name: "old"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MergeMetadata(tt.args.new, tt.args.old)
			if !reflect.DeepEqual(tt.args.new, tt.want) {
				t.Errorf("MergeMetadata() = %v, want %v", tt.args.new, tt.want)
			}
		})
	}
}

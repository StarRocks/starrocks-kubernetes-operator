package hash

import (
	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"testing"
)

func TestHashObject(t *testing.T) {
	type args struct {
		object interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test hash object",
			args: args{
				object: v1.MountInfo{
					Name:      "test",
					MountPath: "/my/path",
				},
			},
			want: "3195412022",
		},
		{
			name: "test hash object 2",
			args: args{
				object: v1.MountInfo{
					Name:      "s1",
					MountPath: "/pkg/mounts/volumes1",
				},
			},
			want: "4145994286",
		},
		{
			name: "test hash object 3",
			args: args{
				object: v1.MountInfo{
					Name:      "s2",
					MountPath: "/pkg/mounts/volumes2",
				},
			},
			want: "1955561708",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HashObject(tt.args.object); got != tt.want {
				t.Errorf("HashObject() = %v, want %v", got, tt.want)
			}
		})
	}
}

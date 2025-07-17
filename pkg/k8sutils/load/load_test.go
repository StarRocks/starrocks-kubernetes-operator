package load

import (
	"reflect"
	"testing"

	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
)

func TestSelector(t *testing.T) {
	type args struct {
		clusterName string
		spec        v1.SpecInterface
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "test selector",
			args: args{
				clusterName: "kube-starrocks",
				spec:        (*v1.StarRocksFeProxySpec)(nil),
			},
			want: map[string]string{
				"app.kubernetes.io/component":       "fe-proxy",
				"app.starrocks.ownerreference/name": "kube-starrocks-fe-proxy",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Selector(tt.args.clusterName, tt.args.spec); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Selector() = %v, want %v", got, tt.want)
			}
		})
	}
}

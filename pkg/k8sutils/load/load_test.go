package load

import (
	"reflect"
	"testing"

	v1 "github.com/CelerData/celerdata-kubernetes-operator-internal/pkg/apis/celerdata/v1"
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
				clusterName: "kube-celerdata",
				spec:        (*v1.CelerDataFeProxySpec)(nil),
			},
			want: map[string]string{
				"app.kubernetes.io/component":       "fe-proxy",
				"app.celerdata.ownerreference/name": "kube-celerdata-fe-proxy",
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

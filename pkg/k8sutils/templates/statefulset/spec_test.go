package statefulset

import (
	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

func TestMakePVCList(t *testing.T) {
	type args struct {
		volumes []v1.StorageVolume
	}
	tests := []struct {
		name string
		args args
		want []corev1.PersistentVolumeClaim
	}{
		{
			name: "test MakePVCList",
			args: args{
				volumes: []v1.StorageVolume{
					{
						Name:             "test",
						StorageClassName: func() *string { name := "test"; return &name }(),
						StorageSize:      "1Gi",
					},
				},
			},
			want: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "test"},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						StorageClassName: &[]string{"test"}[0],
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MakePVCList(tt.args.volumes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakePVCList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeSelector(t *testing.T) {
	type args struct {
		clusterName string
		spec        interface{}
	}
	tests := []struct {
		name string
		args args
		want resource_utils.Labels
	}{
		{
			name: "test MakeSelector",
			args: args{
				clusterName: "test",
				spec:        &v1.StarRocksFeSpec{},
			},
			want: resource_utils.Labels{
				v1.OwnerReference:    "test-fe",
				v1.ComponentLabelKey: "fe",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MakeSelector(tt.args.clusterName, tt.args.spec); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeSelector() = %v, want %v", got, tt.want)
			}
		})
	}
}

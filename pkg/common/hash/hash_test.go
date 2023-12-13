package hash_test

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/hash"
)

func TestHashObject(t *testing.T) {
	type args struct {
		object interface{}
	}
	type Object struct {
		name string
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
			want: "1417214019",
		},
		{
			name: "test hash object 2",
			args: args{
				object: v1.MountInfo{
					Name:      "s1",
					MountPath: "/pkg/mounts/volumes1",
				},
			},
			want: "1614698443",
		},
		{
			name: "test hash object 3",
			args: args{
				object: v1.MountInfo{
					Name:      "s2",
					MountPath: "/pkg/mounts/volumes2",
				},
			},
			want: "122981513",
		},
		{
			name: "test hash object 4",
			args: args{
				object: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "test-warehouse-cn-service",
						Namespace:   "default",
						Annotations: map[string]string{},
						OwnerReferences: []metav1.OwnerReference{
							{
								Name:               "test",
								Controller:         func() *bool { b := true; return &b }(),
								BlockOwnerDeletion: func() *bool { b := true; return &b }(),
							},
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Name:       "thrift",
								Protocol:   "TCP",
								Port:       9060,
								TargetPort: intstr.FromInt(9060),
							},
							{
								Name:       "webserver",
								Protocol:   "TCP",
								Port:       8040,
								TargetPort: intstr.FromInt(8040),
							},
							{
								Name:       "heartbeat",
								Protocol:   "TCP",
								Port:       9050,
								TargetPort: intstr.FromInt(9050),
							},
							{
								Name:       "brpc",
								Protocol:   "TCP",
								Port:       8060,
								TargetPort: intstr.FromInt(8060),
							},
						},
						Type:           corev1.ServiceTypeLoadBalancer,
						LoadBalancerIP: "127.0.0.1",
					},
					Status: corev1.ServiceStatus{},
				},
			},
			want: "749390023",
		},
		{
			name: "test hash object 5",
			args: args{
				object: Object{name: "jack"},
			},
			// the last hash input is 'hash_test.Object{name:"jack"}', which include the type name
			want: "3676726822",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hash.HashObject(tt.args.object); got != tt.want {
				t.Errorf("HashObject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTwoObject(t *testing.T) {
	type T1 struct {
		name string
		age  int
	}
	o1 := T1{
		name: "test",
		age:  10,
	}

	type T2 struct {
		name string
		age  int
	}
	o2 := T2{
		name: "test",
		age:  10,
	}

	o3 := struct {
		name string
		age  int
	}{
		name: "test",
		age:  10,
	}
	o4 := T1{
		name: "test",
	}

	if hash.HashObject(o1) != hash.HashObject(o2) {
		fmt.Println("two object has different types")
	}

	if hash.HashObject(o1) != hash.HashObject(o3) {
		fmt.Println("two object has different types")
	}

	if hash.HashObject(o1) == hash.HashObject(o4) {
		fmt.Println("two object has the same types and values")
	}
}

func TestTwoObject2(t *testing.T) {
	a := int(10)
	b := int(10)
	type T1 struct {
		name string
		age  *int
	}
	o1 := T1{
		name: "test",
		age:  &a,
	}

	o2 := T1{
		name: "test",
		age:  &b,
	}

	if hash.HashObject(o1) == hash.HashObject(o2) {
		fmt.Println("two object has the same types and values")
	}
}

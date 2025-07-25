package controllers

import (
	"context"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	controllerruntime "sigs.k8s.io/controller-runtime"

	v1 "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/fake"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/cn"
)

func TestMain(_ *testing.M) {
	v1.Register()
}

func newStarRocksWarehouseController(objects ...runtime.Object) *StarRocksWarehouseReconciler {
	client := fake.NewFakeClient(v1.Scheme, objects...)
	warehouseController := &StarRocksWarehouseReconciler{
		recorder: record.NewFakeRecorder(10),
		Client:   client,
		subControllers: []subcontrollers.WarehouseSubController{
			cn.New(client, fake.GetEventRecorderFor(nil)),
		},
	}
	return warehouseController
}

func TestStarRocksWarehouseReconciler_Reconcile(t *testing.T) {
	type fields struct {
		reconciler *StarRocksWarehouseReconciler
	}
	type args struct {
		ctx context.Context
		req controllerruntime.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    controllerruntime.Result
		wantErr bool
	}{
		{
			name: "test warehouse reconcile without cr",
			fields: fields{
				reconciler: newStarRocksWarehouseController(),
			},
			args: args{
				ctx: context.TODO(),
				req: controllerruntime.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test",
						Namespace: "test",
					},
				},
			},
			want:    controllerruntime.Result{},
			wantErr: false,
		},
		{
			name: "test warehouse reconcile without specified cluster",
			fields: fields{
				reconciler: newStarRocksWarehouseController(
					&v1.StarRocksWarehouse{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test",
						},
						Spec: v1.StarRocksWarehouseSpec{
							Template: &v1.WarehouseComponentSpec{},
						},
					}),
			},
			args: args{
				ctx: context.TODO(),
				req: controllerruntime.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test",
						Namespace: "test",
					},
				},
			},
			want:    controllerruntime.Result{},
			wantErr: false,
		},
		{
			name: "test warehouse reconcile without cluster",
			fields: fields{
				reconciler: newStarRocksWarehouseController(
					&v1.StarRocksWarehouse{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test",
						},
						Spec: v1.StarRocksWarehouseSpec{
							StarRocksCluster: "cluster",
						},
					}),
			},
			args: args{
				ctx: context.TODO(),
				req: controllerruntime.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test",
						Namespace: "test",
					},
				},
			},
			want:    controllerruntime.Result{},
			wantErr: false,
		},
		{
			name: "test warehouse reconcile with not ready cluster",
			fields: fields{
				reconciler: newStarRocksWarehouseController(
					&v1.StarRocksWarehouse{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test",
						},
						Spec: v1.StarRocksWarehouseSpec{
							StarRocksCluster: "test",
							Template:         &v1.WarehouseComponentSpec{},
						},
					},
					&v1.StarRocksCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test",
						},
						Spec: v1.StarRocksClusterSpec{
							StarRocksFeSpec: &v1.StarRocksFeSpec{
								StarRocksComponentSpec: v1.StarRocksComponentSpec{
									StarRocksLoadSpec: v1.StarRocksLoadSpec{
										ConfigMapInfo: v1.ConfigMapInfo{
											ConfigMapName: "fe-configmap",
											ResolveKey:    "fe.conf",
										},
									},
								},
							},
						},
					},
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "fe-configmap",
							Namespace: "test",
						},
						Data: map[string]string{
							"fe.conf": "run_mode = shared_data",
						},
					},
					&corev1.Endpoints{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-fe-service",
							Namespace: "test",
						},
						Subsets: []corev1.EndpointSubset{{
							// no ready address
							Addresses: []corev1.EndpointAddress{},
						}},
					},
				),
			},
			args: args{
				ctx: context.TODO(),
				req: controllerruntime.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test",
						Namespace: "test",
					},
				},
			},
			want:    controllerruntime.Result{},
			wantErr: false,
		},
		{
			name: "test warehouse reconcile",
			fields: fields{
				reconciler: newStarRocksWarehouseController(
					&v1.StarRocksWarehouse{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test",
						},
						Spec: v1.StarRocksWarehouseSpec{
							StarRocksCluster: "test",
							Template:         &v1.WarehouseComponentSpec{},
						},
					},
					&v1.StarRocksCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "test",
						},
						Spec: v1.StarRocksClusterSpec{
							StarRocksFeSpec: &v1.StarRocksFeSpec{
								StarRocksComponentSpec: v1.StarRocksComponentSpec{
									StarRocksLoadSpec: v1.StarRocksLoadSpec{
										ConfigMapInfo: v1.ConfigMapInfo{
											ConfigMapName: "fe-configmap",
											ResolveKey:    "fe.conf",
										},
									},
								},
							},
						},
					},
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "fe-configmap",
							Namespace: "test",
						},
						Data: map[string]string{
							"fe.conf": "run_mode = shared_data",
						},
					},
					&corev1.Endpoints{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-fe-service",
							Namespace: "test",
						},
						Subsets: []corev1.EndpointSubset{{
							Addresses: []corev1.EndpointAddress{{IP: "127.0.0.1"}},
						}},
					},
				),
			},
			args: args{
				ctx: context.TODO(),
				req: controllerruntime.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test",
						Namespace: "test",
					},
				},
			},
			want:    controllerruntime.Result{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fields.reconciler.Reconcile(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Reconcile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Reconcile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

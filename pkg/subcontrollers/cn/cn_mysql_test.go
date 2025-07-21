package cn

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/fake"
)

func TestNewSQLExecutor(t *testing.T) {
	type args struct {
		ctx       context.Context
		k8sClient client.Client
		namespace string
		name      string
	}
	tests := []struct {
		name    string
		args    args
		want    *SQLExecutor
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "test NewSQLExecutor",
			args: args{
				ctx: context.Background(),
				k8sClient: fake.NewFakeClient(
					func() *runtime.Scheme {
						schema := runtime.NewScheme()
						_ = clientgoscheme.AddToScheme(schema)
						return schema
					}(),
					&appsv1.StatefulSet{
						TypeMeta: metav1.TypeMeta{
							Kind:       "StatefulSet",
							APIVersion: appsv1.SchemeGroupVersion.String(),
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "my-sts",
							Namespace: "default",
							Annotations: map[string]string{
								"test": "test",
							},
						},
						Spec: appsv1.StatefulSetSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Env: []corev1.EnvVar{
												{
													Name:  "MYSQL_PWD",
													Value: "123456",
												},
												{
													Name:  "FE_SERVICE_NAME",
													Value: "fe",
												},
												{
													Name:  "FE_QUERY_PORT",
													Value: "9030",
												},
											},
										},
									},
								},
							},
						},
					},
				),
				namespace: "default",
				name:      "my-sts",
			},
			want: &SQLExecutor{
				RootPassword:       "123456",
				FeServiceName:      "fe",
				FeServiceNamespace: "default",
				FeServicePort:      "9030",
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSQLExecutor(tt.args.ctx, tt.args.k8sClient, tt.args.namespace, tt.args.name)
			if !tt.wantErr(t, err, fmt.Sprintf("NewSQLExecutor(%v, %v, %v, %v)", tt.args.ctx, tt.args.k8sClient, tt.args.namespace, tt.args.name)) {
				return
			}
			assert.Equalf(t, tt.want, got, "NewSQLExecutor(%v, %v, %v, %v)", tt.args.ctx, tt.args.k8sClient, tt.args.namespace, tt.args.name)
		})
	}
}

func TestSQLExecutor_Execute(t *testing.T) {
	type fields struct {
		RootPassword  string
		FeServiceName string
		FeServicePort string
	}
	type args struct {
		ctx        context.Context
		statements string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "test ExecuteContext",
			fields: fields{
				RootPassword:  "root",
				FeServiceName: "localhost",
				FeServicePort: "3306",
			},
			args: args{
				ctx:        context.Background(),
				statements: "drop warehouse test",
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &SQLExecutor{
				RootPassword:  tt.fields.RootPassword,
				FeServiceName: tt.fields.FeServiceName,
				FeServicePort: tt.fields.FeServicePort,
			}

			// create mock db
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			// set expected behavior on mock db
			mock.ExpectExec(tt.args.statements).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err = executor.ExecuteContext(tt.args.ctx, db, tt.args.statements)
			tt.wantErr(t, err, fmt.Sprintf("ExecuteContext(%v, %v)", tt.args.ctx, tt.args.statements))
		})
	}
}

func TestSQLExecutor_QueryShowComputeNodes(t *testing.T) {
	type fields struct {
		RootPassword       string
		FeServiceName      string
		FeServiceNamespace string
		FeServicePort      string
	}
	type args struct {
		ctx context.Context
		db  *sql.DB
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "test ExecuteContext",
			fields: fields{
				RootPassword:  "",
				FeServiceName: "localhost",
				FeServicePort: "9030",
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &SQLExecutor{
				RootPassword:       tt.fields.RootPassword,
				FeServiceName:      tt.fields.FeServiceName,
				FeServiceNamespace: tt.fields.FeServiceNamespace,
				FeServicePort:      tt.fields.FeServicePort,
			}
			// create mock db
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			// set expected behavior on mock db
			mock.ExpectQuery(ShowComputeNodesStatement).WillReturnRows(
				sqlmock.NewRows([]string{"ComputeNodeId", "IP", "WarehouseName"}).AddRow([]byte("id"), []byte("fqdn"), []byte("wh1")),
			)

			result, err := executor.QueryShowComputeNodes(tt.args.ctx, db)
			tt.wantErr(t, err, fmt.Sprintf("ShowComputeNodes(%v, %v)", tt.args.ctx, tt.args.db))
			assert.Equal(t, 1, len(result.ComputeNodesByWarehouse))
			assert.Equal(t, "id", result.ComputeNodesByWarehouse["wh1"][0].ComputeNodeId)
			assert.Equal(t, "fqdn", result.ComputeNodesByWarehouse["wh1"][0].FQDN)
			assert.Equal(t, "wh1", result.ComputeNodesByWarehouse["wh1"][0].WarehouseName)
		})
	}
}

func TestSQLExecutor_ExecuteDropComputeNode(t *testing.T) {
	type fields struct {
		RootPassword       string
		FeServiceName      string
		FeServiceNamespace string
		FeServicePort      string
	}
	type args struct {
		ctx context.Context
		db  *sql.DB
		cn  ComputeNode
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "test ExecuteDropComputeNode",
			fields: fields{
				RootPassword:       "root",
				FeServiceName:      "fe-search",
				FeServiceNamespace: "default",
				FeServicePort:      "9030",
			},
			args: args{
				ctx: context.Background(),
				db: func() *sql.DB {
					db, mock, err := sqlmock.New()
					require.NoError(t, err)
					statement := fmt.Sprintf("ALTER SYSTEM DROP COMPUTE NODE \"%v:%v\" FROM WAREHOUSE %v", "fqdn", "9010", "wh1")
					mock.ExpectExec(statement).WillReturnResult(sqlmock.NewResult(1, 1))
					return db
				}(),
				cn: ComputeNode{
					ComputeNodeId: "id",
					FQDN:          "fqdn",
					HeartbeatPort: "9010",
					WarehouseName: "wh1",
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &SQLExecutor{
				RootPassword:       tt.fields.RootPassword,
				FeServiceName:      tt.fields.FeServiceName,
				FeServiceNamespace: tt.fields.FeServiceNamespace,
				FeServicePort:      tt.fields.FeServicePort,
			}
			tt.wantErr(t, executor.ExecuteDropComputeNode(tt.args.ctx, tt.args.db, tt.args.cn), fmt.Sprintf("ExecuteDropComputeNode(%v, %v, %v)", tt.args.ctx, tt.args.db, tt.args.cn))
		})
	}
}

func TestSQLExecutor_ExecuteDropWarehouse(t *testing.T) {
	type fields struct {
		RootPassword       string
		FeServiceName      string
		FeServiceNamespace string
		FeServicePort      string
	}
	type args struct {
		ctx           context.Context
		db            *sql.DB
		warehouseName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "test ExecuteDropComputeNode",
			fields: fields{
				RootPassword:       "root",
				FeServiceName:      "fe-search",
				FeServiceNamespace: "default",
				FeServicePort:      "9030",
			},
			args: args{
				ctx: context.Background(),
				db: func() *sql.DB {
					db, mock, err := sqlmock.New()
					require.NoError(t, err)
					mock.ExpectExec("DROP WAREHOUSE wh1").
						WillReturnResult(sqlmock.NewResult(1, 1))
					return db
				}(),
				warehouseName: "wh1",
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &SQLExecutor{
				RootPassword:       tt.fields.RootPassword,
				FeServiceName:      tt.fields.FeServiceName,
				FeServiceNamespace: tt.fields.FeServiceNamespace,
				FeServicePort:      tt.fields.FeServicePort,
			}
			tt.wantErr(t, executor.ExecuteDropWarehouse(tt.args.ctx, tt.args.db, tt.args.warehouseName), fmt.Sprintf("ExecuteDropWarehouse(%v, %v, %v)", tt.args.ctx, tt.args.db, tt.args.warehouseName))
		})
	}
}

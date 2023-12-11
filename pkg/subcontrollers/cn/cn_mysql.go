package cn

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-logr/logr"
	_ "github.com/go-sql-driver/mysql" // import mysql driver
	appv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/load"
)

// SQLExecutor is used to execute sql statements.
// Component CN needs to connect to mysql and execute sql statements. E.g.: When StarRocksWarehouse is deleted, the
// related 'DROP WAREHOUSE <name>' statement needs to be executed.
type SQLExecutor struct {
	RootPassword  string
	FeServiceName string
	FeServicePort string
}

// NewSQLExecutor creates a SQLExecutor instance. It will get the root password, fe service name, and fe service port
// from the environment variables of the component CN.
func NewSQLExecutor(ctx context.Context, k8sClient client.Client, namespace, aliasName string) (*SQLExecutor, error) {
	rootPassword := ""
	feServiceName := ""
	feServicePort := ""
	logger := logr.FromContextOrDiscard(ctx)

	var est appv1.StatefulSet
	if err := k8sClient.Get(ctx,
		types.NamespacedName{
			Namespace: namespace,
			Name:      load.Name(aliasName, (*srapi.StarRocksCnSpec)(nil)),
		},
		&est); err != nil {
		return nil, err
	}

	var err error
	for _, envVar := range est.Spec.Template.Spec.Containers[0].Env {
		if envVar.Name == "MYSQL_PWD" {
			rootPassword, err = k8sutils.GetEnvVarValue(ctx, k8sClient, namespace, envVar)
			if err != nil {
				logger.Error(err, "failed to get MYSQL_PWD from env vars, use the default password: empty string")
			}
		} else if envVar.Name == "FE_SERVICE_NAME" {
			feServiceName, err = k8sutils.GetEnvVarValue(ctx, k8sClient, namespace, envVar)
			if err != nil {
				logger.Error(err, "failed to get FE_SERVICE_NAME from env vars")
				return nil, err
			}
		} else if envVar.Name == "FE_QUERY_PORT" {
			feServicePort, err = k8sutils.GetEnvVarValue(ctx, k8sClient, namespace, envVar)
			if err != nil {
				logger.Error(err, "failed to get FE_QUERY_PORT from env vars")
				return nil, err
			}
		}
	}

	return &SQLExecutor{
		RootPassword:  rootPassword,
		FeServiceName: feServiceName,
		FeServicePort: feServicePort,
	}, nil
}

// Execute sql statements. Every time a SQL statement needs to be executed, a new sql.DB instance will be created.
// This is because SQL statements are executed infrequently.
func (executor *SQLExecutor) Execute(ctx context.Context, db *sql.DB, statements string) error {
	var err error
	if db == nil {
		db, err = sql.Open("mysql", fmt.Sprintf("root:%s@tcp(%s:%s)/",
			executor.RootPassword, executor.FeServiceName, executor.FeServicePort))
		if err != nil {
			return err
		}
		defer db.Close()
	}

	_, err = db.ExecContext(ctx, statements)
	if err != nil {
		return err
	}

	return nil
}

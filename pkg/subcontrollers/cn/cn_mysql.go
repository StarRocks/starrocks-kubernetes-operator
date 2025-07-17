package cn

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	"github.com/go-logr/logr"
	_ "github.com/go-sql-driver/mysql" // import mysql driver
	appv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/load"
)

const (
	ShowComputeNodesStatement = "SHOW COMPUTE NODES"
)

// SQLExecutor is used to execute sql statements.
// Component CN needs to connect to mysql and execute sql statements. E.g.: When StarRocksWarehouse is deleted, the
// related 'DROP WAREHOUSE <name>' statement needs to be executed.
type SQLExecutor struct {
	RootPassword       string
	FeServiceName      string
	FeServiceNamespace string
	FeServicePort      string
}

// NewSQLExecutor creates a SQLExecutor instance. It will get the root password, fe service name, and fe service port
// from the environment variables of the component CN.
func NewSQLExecutor(ctx context.Context, k8sClient client.Client, namespace, aliasName string) (*SQLExecutor, error) {
	rootPassword := ""
	feServiceName := ""
	feServicePort := ""
	logger := logr.FromContextOrDiscard(ctx)

	var sts appv1.StatefulSet
	if err := k8sClient.Get(ctx,
		types.NamespacedName{
			Namespace: namespace,
			Name:      load.Name(aliasName, (*srapi.StarRocksCnSpec)(nil)),
		},
		&sts); err != nil {
		return nil, err
	}

	var err error
	for _, envVar := range sts.Spec.Template.Spec.Containers[0].Env {
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
		RootPassword:       rootPassword,
		FeServiceName:      feServiceName,
		FeServiceNamespace: namespace,
		FeServicePort:      feServicePort,
	}, nil
}

// ExecuteContext sql statements. Every time a SQL statement needs to be executed, a new sql.DB instance will be created.
// This is because SQL statements are executed infrequently.
func (executor *SQLExecutor) ExecuteContext(ctx context.Context, db *sql.DB, statement string) error {
	var err error
	if db == nil {
		db, err = sql.Open("mysql", fmt.Sprintf("root:%s@tcp(%s.%s:%s)/",
			executor.RootPassword, executor.FeServiceName, executor.FeServiceNamespace, executor.FeServicePort))
		if err != nil {
			return err
		}
		defer db.Close()
	}

	_, err = db.ExecContext(ctx, statement)
	if err != nil {
		return err
	}

	return nil
}

func (executor *SQLExecutor) QueryContext(ctx context.Context, db *sql.DB, statements string) (*sql.Rows, error) {
	var err error
	if db == nil {
		db, err = sql.Open("mysql", fmt.Sprintf("root:%s@tcp(%s.%s:%s)/",
			executor.RootPassword, executor.FeServiceName, executor.FeServiceNamespace, executor.FeServicePort))
		if err != nil {
			return nil, err
		}
		defer db.Close()
	}

	rows, err := db.QueryContext(ctx, statements)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

type ShowComputeNodesResult struct {
	ComputeNodes []ComputeNode
}

type ComputeNode struct {
	ComputeNodeId string
	FQDN          string
	HeartbeatPort string
}

func (executor *SQLExecutor) QueryShowComputeNodes(ctx context.Context, db *sql.DB) (*ShowComputeNodesResult, error) {
	rows, err := executor.QueryContext(ctx, db, ShowComputeNodesStatement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// iterate over the rows
	result := ShowComputeNodesResult{}
	for rows.Next() {
		var columns []string
		columns, err = rows.Columns()
		if err != nil {
			return nil, err
		}

		// Create a slice of `interface{}` to hold the values dynamically
		// Note: all data types of fields are sql.RawBytes([]byte)
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the row into the slice
		err = rows.Scan(valuePtrs...)
		if err != nil {
			return nil, err
		}

		// Map the values to specific fields based on column names
		computeNode := ComputeNode{}
		for i, col := range columns {
			switch col {
			case "ComputeNodeId":
				computeNode.ComputeNodeId = string(values[i].([]byte))
			case "IP":
				computeNode.FQDN = string(values[i].([]byte))
			case "HeartbeatPort":
				computeNode.HeartbeatPort = string(values[i].([]byte))
			}
		}
		result.ComputeNodes = append(result.ComputeNodes, computeNode)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// The FQDN format is like kube-starrocks-cn-2.kube-starrocks-cn-search.default.svc.cluster.local
	// Sorting the compute nodes by FQDN can help us to remove the last several compute nodes if scale-in operation happens.
	sort.Slice(result.ComputeNodes, func(i, j int) bool {
		return result.ComputeNodes[i].FQDN < result.ComputeNodes[j].FQDN
	})

	return &result, nil
}

func (executor *SQLExecutor) ExecuteDropComputeNode(ctx context.Context, db *sql.DB, node ComputeNode) error {
	dropStatement := fmt.Sprintf("ALTER SYSTEM DROP COMPUTE NODE \"%v:%v\"", node.FQDN, node.HeartbeatPort)
	return executor.ExecuteContext(ctx, db, dropStatement)
}

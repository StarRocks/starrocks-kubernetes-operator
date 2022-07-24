package fe

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/agiledragon/gomonkey/v2"
	"testing"
)

func TestGetNodes(t *testing.T) {
	patches := gomonkey.ApplyFunc(getDb, func(_, _, _ string) (*sql.DB, error) {
		db, mock, err := sqlmock.New()
		if err != nil {
			return nil, err
		}
		/**
		| ComputeNodeId | Cluster         | IP             | HeartbeatPort | BePort | HttpPort | BrpcPort | LastStartTime       | LastHeartbeat       | Alive | SystemDecommissioned | ClusterDecommissioned | ErrMsg                                             | Version          |
		+---------------+-----------------+----------------+---------------+--------+----------+----------+---------------------+---------------------+-------+----------------------+-----------------------+----------------------------------------------------+------------------+
		| 1587000       | default_cluster | 127.0.0.1      | 9050          | 9060   | 8040     | 8060     | 2022-07-24 14:04:12 | 2022-07-24 17:02:45 | true  | false                | false
		**/

		mock.ExpectQuery("show compute nodes").WillReturnRows(sqlmock.NewRows([]string{"ComputeNodeId",
			"Cluster", "IP", "HeartbeatPort", "BePort", "HttpPort", "BrpcPort", "LastStartTime", "LastHeartbeat",
			"Alive", "SystemDecommissioned", "ClusterDecommissioned", "ErrMsg", "Version"}).
			AddRow("1587000", "default_cluster", "127.0.0.1", "9050", "9060", "8040", "8060", "2022-07-24 14:04:12", "2022-07-24 17:02:45",
				"true", "false", "false", "", "2.2.2-76-36ea96e"))
		return db, nil
	})
	defer patches.Reset()

	nodes, err := GetNodes("feAddr", "usr", "pwd")
	t.Log(nodes)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 {
		t.Fatal("nodes len != 1")
	}
	if nodes[0].ComputeNodeId != "1587000" {
		t.Fatal("nodes[0].ComputeNodeId != 1587000")
	}
	if nodes[0].Ip != "127.0.0.1" {
		t.Fatal("nodes[0].Ip != 127.0.0.1)")
	}
}

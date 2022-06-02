/*
Copyright 2022 StarRocks.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package fe

import (
	"database/sql"
	"fmt"
	"k8s.io/klog/v2"
	"math/rand"
	"strings"
	"time"

	// mysql driver init
	_ "github.com/go-sql-driver/mysql"
)

type Sql string

var (
	SqlGetNodes Sql = "show compute nodes"
	SqlAddNode  Sql = "alter system add compute node"
	SqlDropNode Sql = "alter system drop compute node"
)

// CN table:
// ComputeNodeId | Cluster | IP |
// HeartbeatPort | BePort | HttpPort | BrpcPort | LastStartTime | LastHeartbeat |
// Alive | SystemDecommissioned | ClusterDecommissioned | ErrMsg
type NodeInfo struct {
	ComputeNodeId         string
	Cluster               string
	Ip                    string
	HeartbeatPort         int32
	BePort                int32
	HttpPort              int32
	BrpcPort              int32
	LastStartTime         sql.NullString
	LastHeartbeat         sql.NullString
	Alive                 bool
	SystemDecommissioned  bool
	ClusterDecommissioned bool
	ErrMsg                string
	Version               string
}

// add cn node to fe
func AddNode(feAddr, usr, pwd, cnAddr string) error {
	db, err := getDb(feAddr, usr, pwd)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(fmt.Sprintf("%s '%s'", SqlAddNode, cnAddr))
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			klog.Warningf("%v already exists", cnAddr)
			return nil
		}
		return err
	}
	return nil
}

// drop cn node on fe
func DropNode(feAddr, usr, pwd, cnAddr string) error {
	db, err := getDb(feAddr, usr, pwd)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(fmt.Sprintf("%s '%s'", SqlDropNode, cnAddr))
	if err != nil {
		if strings.Contains(err.Error(), "not exists") {
			klog.Warningf("%v not exists", cnAddr)
			return nil
		}
		return err
	}
	return nil
}

// get all cn nodes on fe
func GetNodes(feAddr, usr, pwd string) ([]*NodeInfo, error) {
	db, err := getDb(feAddr, usr, pwd)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query(string(SqlGetNodes))
	if err != nil {
		return nil, err
	}
	var nodes []*NodeInfo
	for rows.Next() {
		node := &NodeInfo{}
		err := rows.Scan(
			&node.ComputeNodeId,
			&node.Cluster,
			&node.Ip,
			&node.HeartbeatPort,
			&node.BePort,
			&node.HttpPort,
			&node.BrpcPort,
			&node.LastStartTime,
			&node.LastHeartbeat,
			&node.Alive,
			&node.SystemDecommissioned,
			&node.ClusterDecommissioned,
			&node.ErrMsg,
			&node.Version,
		)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func getDb(url, usr, pwd string) (*sql.DB, error) {
	source := fmt.Sprintf("%s:%s@tcp(%s)/", usr, pwd, url)
	db, err := sql.Open("mysql", source)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(30 * time.Second)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	return db, nil
}

// return a random fe addr
func PickFe(feAddrs []string) string {
	rand.Seed(time.Now().UnixNano())
	randInt := rand.Intn(2048)
	fePick := feAddrs[randInt%len(feAddrs)]
	return fePick
}

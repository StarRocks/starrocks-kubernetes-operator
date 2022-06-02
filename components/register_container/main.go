/*
Copyright 2022, StarRocks Limited

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
package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/StarRocks/starrocks-kubernetes-operator/common"
	"github.com/StarRocks/starrocks-kubernetes-operator/internal/fe"
	"github.com/avast/retry-go/v4"
	"k8s.io/klog/v2"
)

/*
register container is for register cn node to fe, container run as a sidecar
*/

func getIp() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}

	for _, address := range addrs {
		// skip loopbackAddress
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	panic("can not found ip")
}

// cn alive check
func alive(ip string) (bool, error) {
	_, err := http.Get(fmt.Sprintf("http://%s:8040/api/health", ip))
	if err != nil {
		return false, err
	}
	return true, nil
}

func main() {
	port := os.Getenv(common.EnvKeyCnPort)
	feAddrsStr := strings.TrimSpace(os.Getenv(common.EnvKeyFeAddrs))
	usr := os.Getenv(common.EnvKeyFeUsr)
	pwd := os.Getenv(common.EnvKeyFePwd)

	hostIp := getIp()
	cnAddr := fmt.Sprintf("%s:%s", hostIp, port)
	feAddrs := strings.Split(feAddrsStr, common.FeAddrsSeparator)

	ch := make(chan struct{}) // block the process
	for i := 0; i < 2; i++ {
		_ = retry.Do(func() error {
			_, err := alive(hostIp)
			if err != nil {
				klog.Error(err)
				return err
			}
			fePick := fe.PickFe(feAddrs)
			klog.Info(fePick)
			err = fe.AddNode(fePick, usr, pwd, cnAddr)
			if err != nil {
				klog.Error(err)
				return err
			}
			return nil
		}, retry.DelayType(retry.BackOffDelay), retry.Attempts(0), retry.MaxDelay(20*time.Second))
		time.Sleep(1 * time.Minute)
	}
	<-ch
}

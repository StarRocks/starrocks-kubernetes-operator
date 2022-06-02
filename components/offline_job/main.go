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
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/StarRocks/starrocks-kubernetes-operator/common"
	"github.com/StarRocks/starrocks-kubernetes-operator/internal/fe"
	"github.com/avast/retry-go/v4"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

/*
offline job is for deregister useless cn node on fe
*/

var (
	feAddrsStr = strings.TrimSpace(os.Getenv(common.EnvKeyFeAddrs))
	usr        = os.Getenv(common.EnvKeyFeUsr)
	pwd        = os.Getenv(common.EnvKeyFePwd)
	cnNs       = os.Getenv(common.EnvKeyCnNs)
	cnName     = os.Getenv(common.EnvKeyCnName)
	cnPort     = os.Getenv(common.EnvKeyCnPort)
)

func main() {
	feAddrs := strings.Split(feAddrsStr, common.FeAddrsSeparator)

	cfg, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	dynamidCli, err := dynamic.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}
	kubeCli, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}

	err = retry.Do(func() error {
		fePick := fe.PickFe(feAddrs)
		err := sync(dynamidCli, kubeCli, fePick)
		if err != nil {
			klog.Error(err)
			return err
		}
		return nil
	}, retry.DelayType(retry.BackOffDelay), retry.Attempts(5))
	if err != nil {
		panic(err)
	}
}

// drop cn node on fe which not survive on k8s
func sync(dynamicCli dynamic.Interface, kubeCli kubernetes.Interface, feAddr string) error {
	nodes, err := fe.GetNodes(feAddr, usr, pwd)
	if err != nil {
		return err
	}
	registedIpsSet := sets.NewString()
	for _, node := range nodes {
		// skip node which not belong to this cr, a temporary handling before labelling cn
		if node.Alive {
			continue
		}
		registedIpsSet = registedIpsSet.Insert(node.Ip)
	}
	deploy, err := kubeCli.AppsV1().Deployments(cnNs).Get(context.TODO(), cnName, v1.GetOptions{})
	if err != nil {
		return err
	}
	labelMap, err := v1.LabelSelectorAsMap(deploy.Spec.Selector)
	if err != nil {
		return err
	}
	pods, err := kubeCli.CoreV1().Pods(cnNs).List(context.TODO(), v1.ListOptions{LabelSelector: labels.SelectorFromSet(labelMap).String()})
	if err != nil {
		return err
	}

	podIpsSet := sets.NewString()
	for _, pod := range pods.Items {
		podIpsSet.Insert(pod.Status.PodIP)
	}
	klog.Infof("registed: %v", registedIpsSet)
	klog.Infof("podIps: %v", podIpsSet)
	for _, registed := range registedIpsSet.List() {
		if podIpsSet.Has(registed) {
			continue
		}
		addr := fmt.Sprintf("%s:%s", registed, cnPort)
		klog.Infof("drop node %v", addr)
		if err = fe.DropNode(feAddr, usr, pwd, addr); err != nil {
			return err
		}
	}
	return nil
}

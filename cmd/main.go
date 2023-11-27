/*
Copyright 2021-present, StarRocks Inc.

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
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	"k8s.io/klog/v2"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/StarRocks/starrocks-kubernetes-operator/pkg"
	//+kubebuilder:scaffold:imports
)

var (
	VERSION    string
	GOVERSION  string
	COMMIT     string
	BUILD_DATE string
)

var (
	_setupLog             = ctrl.Log.WithName("setup")
	_metricsAddr          string
	_enableLeaderElection bool
	_probeAddr            string
	_printVar             bool
	_namespace            string
)

// Print version information to a given out writer.
func Print(out io.Writer) {
	if _printVar {
		fmt.Fprint(out, "version="+VERSION+"\ngoversion="+GOVERSION+"\ncommit="+COMMIT+"\nbuild_date="+BUILD_DATE+"\n")
	}
}

func init() {
	// the implied flag: kubeconfig.
	// KUBECONFIG env will be used if you have config.
	flag.StringVar(&_metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&_probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&_enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&_printVar, "version", false, "Prints current version.")
	flag.StringVar(&_namespace, "namespace", "", "if specified, "+
		"restricts the manager's cache to watch objects in the desired namespace. Defaults to all namespaces.")

	// set klog flag
	klog.InitFlags(nil)
	// to use klog.V for debugging, we have to set the flag.
	_ = flag.Set("v", "2")
}

func main() {
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	if _printVar {
		Print(os.Stdout)
		return
	}

	duration := 2 * time.Minute
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 pkg.Scheme,
		MetricsBindAddress:     _metricsAddr,
		Port:                   9443,
		SyncPeriod:             &duration,
		HealthProbeBindAddress: _probeAddr,
		LeaderElection:         _enableLeaderElection,
		LeaderElectionID:       "c6c79638.starrocks.com",
		Namespace:              _namespace,
	})
	if err != nil {
		_setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// setup all reconciles
	if err := pkg.SetupClusterReconciler(mgr); err != nil {
		_setupLog.Error(err, "unable to set up cluster reconciler")
		os.Exit(1)
	}

	if err := pkg.SetupWarehouseReconciler(mgr); err != nil {
		_setupLog.Error(err, "unable to set up warehouse reconciler")
		os.Exit(1)
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		_setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		_setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	if err := k8sutils.GetKubernetesVersion(); err != nil {
		_setupLog.Error(err, "unable to get kubernetes version, continue to start manager")
	}

	_setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		_setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

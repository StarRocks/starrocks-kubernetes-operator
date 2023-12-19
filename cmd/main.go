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
	"context"
	"flag"
	"os"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/controllers"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
)

var (
	_metricsAddr          string
	_enableLeaderElection bool
	_probeAddr            string
	_namespace            string
)

func main() {
	flag.StringVar(&_metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&_probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&_enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&_namespace, "namespace", "", "if specified, "+
		"restricts the manager's cache to watch objects in the desired namespace. Defaults to all namespaces.")

	// Set up logger.
	opts := zap.Options{}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	logger := ctrl.Log.WithName("main")

	// Register CRD to SchemeBuilder
	srapi.Register()

	duration := 2 * time.Minute
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 srapi.Scheme,
		MetricsBindAddress:     _metricsAddr,
		Port:                   9443,
		SyncPeriod:             &duration,
		HealthProbeBindAddress: _probeAddr,
		LeaderElection:         _enableLeaderElection,
		LeaderElectionID:       "c6c79638.starrocks.com",
		Namespace:              _namespace,
	})
	if err != nil {
		logger.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// setup all reconciles
	if err := controllers.SetupClusterReconciler(mgr); err != nil {
		logger.Error(err, "unable to set up cluster reconciler")
		os.Exit(1)
	}

	if err := controllers.SetupWarehouseReconciler(context.Background(), mgr); err != nil {
		logger.Error(err, "unable to set up warehouse reconciler")
		os.Exit(1)
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		logger.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		logger.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	if err := k8sutils.GetKubernetesVersion(); err != nil {
		logger.Error(err, "unable to get kubernetes version, continue to start manager")
	}

	logger.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		logger.Error(err, "problem running manager")
		os.Exit(1)
	}
}

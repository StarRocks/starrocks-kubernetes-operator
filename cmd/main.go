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
	"fmt"
	"os"
	"time"

	"github.com/go-logr/logr"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/StarRocks/starrocks-kubernetes-operator/cmd/config"
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/controllers"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	srwebhook "github.com/StarRocks/starrocks-kubernetes-operator/pkg/webhook"
)

var (
	_metricsAddr          string
	_enableLeaderElection bool
	_probeAddr            string
	_namespace            string
	_enableWebhooks       bool
)

// caBundleUpdater updates the ValidatingAdmissionWebhook with CA bundle after manager starts
type caBundleUpdater struct {
	client client.Client
	caCert []byte
	logger logr.Logger
}

// Start implements the Runnable interface
func (c *caBundleUpdater) Start(ctx context.Context) error {
	c.logger.Info("updating webhook CA bundle")

	webhookName := "starrockscluster-validating-webhook"

	// Get the existing ValidatingAdmissionWebhook
	webhook := &admissionregistrationv1.ValidatingWebhookConfiguration{}
	if err := c.client.Get(ctx, types.NamespacedName{Name: webhookName}, webhook); err != nil {
		return fmt.Errorf("failed to get ValidatingAdmissionWebhook: %v", err)
	}

	// Update caBundle for all webhooks
	for i := range webhook.Webhooks {
		webhook.Webhooks[i].ClientConfig.CABundle = c.caCert
	}

	// Update the webhook configuration
	if err := c.client.Update(ctx, webhook); err != nil {
		return fmt.Errorf("failed to update ValidatingAdmissionWebhook: %v", err)
	}

	c.logger.Info("webhook CA bundle updated successfully")
	return nil
}

func main() {
	flag.StringVar(&_metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&_probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&_enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&_namespace, "namespace", "", "if specified, "+
		"restricts the manager's cache to watch objects in the desired namespace. Defaults to all namespaces.")
	flag.StringVar(&config.DNSDomainSuffix, "dns-domain-suffix", "cluster.local", "The suffix of the dns domain in k8s")
	flag.BoolVar(&config.VolumeNameWithHash, "volume-name-with-hash", true, "Add a hash to the volume name")
	flag.BoolVar(&_enableWebhooks, "enable-webhooks", false, "Enable admission webhooks. "+
		"Requires TLS certificates to be configured. Disabled by default for local development.")
	flag.IntVar(&config.WebhookCertValidityDays, "webhook-cert-validity-days", 365,
		"Validity period in days for self-signed webhook certificates")

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

	if err := controllers.SetupWarehouseReconciler(mgr, _namespace); err != nil {
		logger.Error(err, "unable to set up warehouse reconciler")
		os.Exit(1)
	}

	// setup webhooks if enabled
	if _enableWebhooks {
		// Get webhook configuration from environment variables
		webhookNamespace := os.Getenv("WEBHOOK_NAMESPACE")
		webhookServiceName := os.Getenv("WEBHOOK_SERVICE_NAME")
		webhookCertDir := os.Getenv("WEBHOOK_CERT_DIR")

		// Generate self-signed certificates if not provided
		caCert, err := srwebhook.GenerateSelfSignedCerts(webhookCertDir, webhookServiceName, webhookNamespace, config.WebhookCertValidityDays)
		if err != nil {
			logger.Error(err, "unable to generate self-signed certificates")
			os.Exit(1)
		}
		logger.Info("self-signed certificates generated for webhook server", "validityDays",
			config.WebhookCertValidityDays, "namespace", webhookNamespace, "serviceName", webhookServiceName, "certDir", webhookCertDir)

		// Add CA bundle updater as a runnable that starts after the manager cache is ready
		if err := mgr.Add(&caBundleUpdater{
			client: mgr.GetClient(),
			caCert: caCert,
			logger: logger.WithName("ca-bundle-updater"),
		}); err != nil {
			logger.Error(err, "unable to add CA bundle updater")
			os.Exit(1)
		}

		if err := setupWebhooks(mgr); err != nil {
			logger.Error(err, "unable to set up webhooks")
			os.Exit(1)
		}
		logger.Info("webhooks enabled")
	} else {
		logger.Info("webhooks disabled - use --enable-webhooks to enable")
	}

	// +kubebuilder:scaffold:builder

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

// setupWebhooks sets up webhooks for the manager
func setupWebhooks(mgr ctrl.Manager) error {
	// Set up the webhook server
	hookServer := mgr.GetWebhookServer()
	if hookServer == nil {
		return fmt.Errorf("failed to get webhook server from manager")
	}

	// Register the StarRocksCluster validating webhook
	validator := &srwebhook.StarRocksClusterValidator{
		Client: mgr.GetClient(),
	}

	hookServer.Register("/validate-starrocks-com-v1-starrockscluster",
		&webhook.Admission{Handler: validator})

	return nil
}

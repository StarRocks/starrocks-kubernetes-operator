package feproxy

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cmdconfig "github.com/StarRocks/starrocks-kubernetes-operator/cmd/config"
	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
	rutils "github.com/StarRocks/starrocks-kubernetes-operator/pkg/common/resource_utils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/load"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/k8sutils/templates/service"
	"github.com/StarRocks/starrocks-kubernetes-operator/pkg/subcontrollers/fe"
)

func (controller *FeProxyController) SyncConfigMap(ctx context.Context, src *srapi.StarRocksCluster) error {
	feProxySpec := src.Spec.StarRocksFeProxySpec

	feSpec := src.Spec.StarRocksFeSpec
	config, err := fe.GetFEConfig(ctx, controller.k8sClient, feSpec, src.Namespace)
	if err != nil {
		return err
	}
	httpPort := rutils.GetPort(config, rutils.HTTP_PORT)

	feSearchServiceName := service.SearchServiceName(src.Name, feSpec)
	feExternalServiceName := service.ExternalServiceName(src.Name, feSpec)
	proxyPass := fmt.Sprintf("http://%s.%s.%s:%d", feExternalServiceName, src.GetNamespace(), cmdconfig.GetServiceDomainSuffix(), httpPort)

	resolver := feProxySpec.Resolver
	if resolver == "" {
		resolver = fmt.Sprintf("%s.%s", "kube-dns.kube-system", cmdconfig.GetServiceDomainSuffix())
	}

	or := metav1.NewControllerRef(src, src.GroupVersionKind())
	configmap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            load.Name(src.Name, feProxySpec),
			Namespace:       src.Namespace,
			Labels:          load.Labels(src.Name, feProxySpec),
			OwnerReferences: []metav1.OwnerReference{*or},
		},
		Data: map[string]string{
			"nginx.conf": fmt.Sprintf(`
pid   /tmp/nginx.pid;
worker_processes 4;
include /usr/share/nginx/modules/*.conf;
events {
  worker_connections 256;
}

http {
  sendfile            on;
  tcp_nopush          on;
  tcp_nodelay         on;
  keepalive_timeout   65;
  types_hash_max_size 2048;
  client_max_body_size 0;
  ignore_invalid_headers off;
  underscores_in_headers on;
  proxy_read_timeout 600s;
  proxy_http_version 1.1;

  client_body_temp_path /tmp/client_temp;
  proxy_temp_path       /tmp/proxy_temp_path;
  fastcgi_temp_path     /tmp/fastcgi_temp;
  uwsgi_temp_path       /tmp/uwsgi_temp;
  scgi_temp_path        /tmp/scgi_temp;

  default_type        application/octet-stream;

  server {
    listen 8080;
    resolver %[1]s valid=10s;
    proxy_intercept_errors on;
    recursive_error_pages on;

    location /nginx/health {
      access_log off;
      return 200;
    }

    location / {
      # see https://serverfault.com/questions/240476/how-to-force-nginx-to-resolve-dns-of-a-dynamic-hostname-everytime-when-doing-p/593003#593003 for why we use set
      set $fe_service "%[2]s";
      proxy_pass $fe_service;
      proxy_set_header Expect $http_expect;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      error_page 307 = @handle_redirect;
    }

    location /api/transaction/load {
      set $fe_service "%[2]s";
      proxy_pass $fe_service;
      proxy_pass_request_body off;
      proxy_set_header Expect $http_expect;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      error_page 307 = @handle_redirect;
    }

    location ~ ^/api/.*/.*/_stream_load$ {
      set $fe_service "%[2]s";
      proxy_pass $fe_service;
      proxy_pass_request_body off;
      proxy_set_header Expect $http_expect;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      error_page 307 = @handle_redirect;
    }

    location @handle_redirect {
      if ($upstream_http_location ~ "%[3]s") {
        rewrite ^ /_redirect_to_fe last;
      }
      if ($upstream_http_location !~ "%[3]s") {
        rewrite ^ /_redirect_to_others last;
      }
    }

    location /_redirect_to_fe {
      set $redirect_uri '$upstream_http_location';
      proxy_pass $redirect_uri;
      proxy_set_header Expect $http_expect;
      proxy_pass_request_body off;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      error_page 307 = @handle_redirect;
    }

    location /_redirect_to_others {
      set $redirect_uri '$upstream_http_location';
      proxy_pass $redirect_uri;
      proxy_set_header Expect $http_expect;
      proxy_pass_request_body on;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      error_page 307 = @handle_redirect;
    }
  }
}
`, resolver, proxyPass, feSearchServiceName),
		},
	}

	return k8sutils.ApplyConfigMap(ctx, controller.k8sClient, &configmap)
}

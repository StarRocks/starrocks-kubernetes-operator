#!/bin/bash

# Generate self-signed certificates for webhook server local development
# This script creates TLS certificates in the temporary directory expected by the webhook server

set -e

CERT_DIR="/tmp/k8s-webhook-server/serving-certs"
SERVICE_NAME="webhook-service"
NAMESPACE="starrocks-operator-system"

echo "Generating self-signed certificates for webhook server..."

# Create certificate directory
mkdir -p "${CERT_DIR}"

# Generate private key
openssl genrsa -out "${CERT_DIR}/tls.key" 2048

# Generate certificate signing request
cat > "${CERT_DIR}/csr.conf" <<EOF
[req]
default_bits = 2048
prompt = no
distinguished_name = dn
req_extensions = v3_req

[dn]
CN = ${SERVICE_NAME}.${NAMESPACE}.svc

[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = ${SERVICE_NAME}
DNS.2 = ${SERVICE_NAME}.${NAMESPACE}
DNS.3 = ${SERVICE_NAME}.${NAMESPACE}.svc
DNS.4 = ${SERVICE_NAME}.${NAMESPACE}.svc.cluster.local
DNS.5 = localhost
IP.1 = 127.0.0.1
EOF

# Generate certificate signing request
openssl req -new -key "${CERT_DIR}/tls.key" -out "${CERT_DIR}/tls.csr" -config "${CERT_DIR}/csr.conf"

# Generate self-signed certificate
openssl x509 -req -in "${CERT_DIR}/tls.csr" -signkey "${CERT_DIR}/tls.key" -out "${CERT_DIR}/tls.crt" -days 365 -extensions v3_req -extfile "${CERT_DIR}/csr.conf"

# Clean up
rm "${CERT_DIR}/tls.csr" "${CERT_DIR}/csr.conf"

echo "Certificates generated successfully in ${CERT_DIR}"
echo "You can now run the operator with --enable-webhooks flag"
echo ""
echo "Files created:"
echo "  ${CERT_DIR}/tls.key (private key)"
echo "  ${CERT_DIR}/tls.crt (certificate)"
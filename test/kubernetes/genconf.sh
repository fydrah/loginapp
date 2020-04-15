#!/bin/bash

# /!\ For testing purpose only

CURR_DIR=$(dirname $0)

NODE_IP=$1

mkdir -p ${CURR_DIR}/generated/ssl

for cert in dex loginapp
do

echo "Generating CSR for ${cert}"

cat << EOF > ${CURR_DIR}/generated/ssl/req-${cert}.cnf
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name

[req_distinguished_name]

[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment, dataEncipherment, keyAgreement, keyCertSign, cRLSign
extendedKeyUsage = clientAuth, serverAuth, emailProtection, codeSigning
subjectAltName = @alt_names

[alt_names]
DNS.1 = ${cert}.${NODE_IP}.nip.io
EOF

[ -e ${CURR_DIR}/generated/ssl/key-${cert}.pem ] || openssl genrsa -out ${CURR_DIR}/generated/ssl/key-${cert}.pem 2048 >/dev/null
[ -e ${CURR_DIR}/generated/ssl/csr-${cert}.pem ] || openssl req -new -key ${CURR_DIR}/generated/ssl/key-${cert}.pem \
    -out ${CURR_DIR}/generated/ssl/csr-${cert}.pem -subj "/CN=kubernetes" -config ${CURR_DIR}/generated/ssl/req-${cert}.cnf >/dev/null

kubectl get csr ${cert} >/dev/null || cat <<EOF | kubectl create -f -
apiVersion: certificates.k8s.io/v1beta1
kind: CertificateSigningRequest
metadata:
  name: ${cert}
spec:
  request: $(base64 -i ${CURR_DIR}/generated/ssl/csr-${cert}.pem -w0)
  usages:
  - digital signature
  - key encipherment
  - client auth
  - server auth
EOF
echo "OK"

echo "Approving CSR for ${cert}"
sleep 0.5
[ "$(kubectl get csr ${cert} -o jsonpath='{.status.conditions[0].type}')" == "Approved" ] || kubectl certificate approve ${cert}
echo "OK"

echo "Waiting for certificate for ${cert}"
while [ -z "$(kubectl get csr ${cert} -o jsonpath='{.status.certificate}')" ]
do
  sleep 0.5
  echo -n "."
done
echo "OK"

echo "Creating TLS secret for ${cert}"
cat <<EOF > ${CURR_DIR}/generated/${cert}-certs.yaml
---
apiVersion: v1
kind: Secret
metadata:
  name: ${cert}-certs
  namespace: kube-system
type: kubernetes.io/tls
data:
  tls.crt: $(kubectl get csr ${cert} -o jsonpath='{.status.certificate}')
  tls.key: $(base64 -i ${CURR_DIR}/generated/ssl/key-${cert}.pem -w0)
EOF
done

echo "Generating dex and loginapp configurations"

### Dex
cat <<EOF > ${CURR_DIR}/generated/dex-config.yaml
---
kind: ConfigMap
apiVersion: v1
metadata:
  name: dex
  namespace: kube-system
data:
  config.yaml: |
    issuer: https://dex.${NODE_IP}.nip.io:32000
    storage:
      type: kubernetes
      config:
        inCluster: true
    web:
      https: 0.0.0.0:5556
      tlsCert: /etc/dex/tls/tls.crt
      tlsKey: /etc/dex/tls/tls.key
    oauth2:
      skipApprovalScreen: true

    staticClients:
    - id: loginapp
      redirectURIs:
      - 'https://loginapp.${NODE_IP}.nip.io:32001/callback'
      name: 'Example App'
      secret: ZXhhbXBsZS1hcHAtc2VjcmV0

    enablePasswordDB: true
    staticPasswords:
    - email: "admin@example.com"
      # bcrypt hash of the string "password"
      hash: "\$2a\$10\$2b2cU8CPhOTaGrs1HRQuAueS7JTT5ZHsHSzYiFPm1leZck7Mc8T4W"
      username: "admin"
      userID: "08a8684b-db88-4b73-90a9-3cd1661f5466"
EOF

### Loginapp
cat <<EOF > ${CURR_DIR}/generated/loginapp-config.yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: loginapp-config
  namespace: kube-system
data:
  config.yaml: |
    name: "Kubernetes Auth"
    listen: "0.0.0.0:8443"
    oidc:
      client:
        id: "loginapp"
        secret: ZXhhbXBsZS1hcHAtc2VjcmV0
        redirectURL: "https://loginapp.${NODE_IP}.nip.io:32001/callback"
      issuer:
        rootCA: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
        url: "https://dex.${NODE_IP}.nip.io:32000"
      extraScopes:
        - groups
    tls:
      enabled: true
      cert: /ssl/tls.crt
      key: /ssl/tls.key
    log:
      level: Debug
      format: json
    clusters:
      - name: myfakecluster
        server: https://myfakecluster.org
        certificate-authority: |
          -----BEGIN CERTIFICATE-----
          MIIC/zCCAeegAwIBAgIULkYvGJPRl50tMoVE4BNM0laRQncwDQYJKoZIhvcNAQEL
          BQAwDzENMAsGA1UEAwwEbXljYTAeFw0xOTAyMTgyMjA5NTJaFw0xOTAyMjgyMjA5
          NTJaMA8xDTALBgNVBAMMBG15Y2EwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK
          -----END CERTIFICATE-----
        insecure-skip-tls-verify: false
EOF

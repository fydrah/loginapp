#!/bin/bash

# /!\ For testing purpose only

CURR_DIR=$(dirname $0)
NODE_IP=$(docker inspect loginapp-control-plane -f '{{ .NetworkSettings.Networks.kind.IPAddress }}')

mkdir -p ${CURR_DIR}/generated/ssl ${CURR_DIR}/kubernetes/generated ${CURR_DIR}/helm/generated

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
DNS.2 = ${cert}.127.0.0.1.nip.io
EOF

[ -e ${CURR_DIR}/generated/ssl/key-${cert}.pem ] || openssl genrsa -out ${CURR_DIR}/generated/ssl/key-${cert}.pem 2048 >/dev/null
[ -e ${CURR_DIR}/generated/ssl/csr-${cert}.pem ] || openssl req -new -key ${CURR_DIR}/generated/ssl/key-${cert}.pem \
    -out ${CURR_DIR}/generated/ssl/csr-${cert}.pem -subj "/CN=system:node:loginapp;/O=system:nodes" -config ${CURR_DIR}/generated/ssl/req-${cert}.cnf >/dev/null

kubectl get csr ${cert} >/dev/null || cat <<EOF | kubectl create -f -
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: ${cert}
spec:
  signerName: kubernetes.io/kubelet-serving
  request: $(base64 -i ${CURR_DIR}/generated/ssl/csr-${cert}.pem -w0)
  usages:
  - key encipherment
  - digital signature
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
kubectl get csr ${cert} -o jsonpath='{.status.certificate}' | base64 -d > ${CURR_DIR}/generated/ssl/${cert}.crt
cat <<EOF > ${CURR_DIR}/kubernetes/generated/${cert}-certs.yaml
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
cat <<EOF > ${CURR_DIR}/helm/generated/dex-overrides.yaml
fullnameOverride: dex
https:
  enabled: true
service:
  type: NodePort
  ports:
    https:
      nodePort: 32000
volumes:
  - name: tls
    secret:
      secretName: dex-certs
volumeMounts:
  - name: tls
    mountPath: /etc/dex/tls
config:
  issuer: https://dex.${NODE_IP}.nip.io:32000
  storage:
    type: kubernetes
    config:
      inCluster: true
  web:
    https: 0.0.0.0:5554
    tlsCert: /etc/dex/tls/tls.crt
    tlsKey: /etc/dex/tls/tls.key
  oauth2:
    skipApprovalScreen: true

  staticClients:
  - id: loginapp
    redirectURIs:
    - 'https://loginapp.${NODE_IP}.nip.io:32001/callback'
    name: 'Loginapp Kube'
    secret: ZXhhbXBsZS1hcHAtc2VjcmV0
  - id: loginapp-local
    redirectURIs:
    - 'https://loginapp.127.0.0.1.nip.io:8443/callback'
    name: 'Loginapp local'
    secret: ZXhhbXBsZS1hcHAtc2VjcmV1

  enablePasswordDB: true
  staticPasswords:
  - email: "admin@example.com"
    # bcrypt hash of the string "password"
    hash: "$2a$10$2b2cU8CPhOTaGrs1HRQuAueS7JTT5ZHsHSzYiFPm1leZck7Mc8T4W"
    username: "admin"
    userID: "08a8684b-db88-4b73-90a9-3cd1661f5466"
EOF

### Loginapp
cat <<EOF > ${CURR_DIR}/kubernetes/generated/loginapp-config.yaml
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
        redirectURL: "https://loginapp.${NODE_IP}.nip.io:32001/callback"
      issuer:
        rootCA: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
        url: "https://dex.${NODE_IP}.nip.io:32000"
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
          MIIDZzCCAk+gAwIBAgIRAM/Oqk7538VUzHwMZ+x7eN0wDQYJKoZIhvcNAQELBQAw
          FTETMBEGA1UEAxMKa3ViZXJuZXRlczAeFw0yMTExMDUxMDUyMDdaFw0yMjExMDUx
          MDUyMDdaMDcxFTATBgNVBAoTDHN5c3RlbTpub2RlczEeMBwGA1UEAwwVc3lzdGVt
          Om5vZGU6bG9naW5hcHA7MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA
          yc4mV0leq+fJ5KsF6yZTKtWRVzwcZGeYZv3gf1R9zSdexshOtBORRaNHJmuAu0st
          aHunVPlsSInBGBf4pO4OTLilu965/OQOvFvIFLzl3LNMoHOvUMofJ8lwBlcczil+
          UYf+xvDNYOYN3Q8WIBs5W2yHgDdFdfb9Z/X6ezL31T0/1+VYXR5cSawIKP9zEY/z
          miUWqrqRVZN0qvFYp5phMUbcIlNknEo/p0nUXfnz87P/7862BJuVaJVZmAMcg3IS
          q9Xl27P4NVnIvI5WPrlNrDNv3AupEJFS0c/Gms3nr6BZWNg5tM2mRhAr1Gfm37me
          fFqM0gxJn2sHjSF5y+M0jQIDAQABo4GPMIGMMA4GA1UdDwEB/wQEAwIFoDATBgNV
          HSUEDDAKBggrBgEFBQcDATAMBgNVHRMBAf8EAjAAMB8GA1UdIwQYMBaAFOw9vdeK
          hC8ALh7H+2QYpXPUWUU6MDYGA1UdEQQvMC2CFWRleC4xNzIuMTcuMC4yLm5pcC5p
          b4IUZGV4LjEyNy4wLjAuMS5uaXAuaW8wDQYJKoZIhvcNAQELBQADggEBAHgdgQ07
          YXERpAw8wRxxxX6Ozsod/vKww6E5Q4KmU8FxTqkx7hxQpp3neC6b5+yrsbGw2h0R
          8U9Kog+oNLtoN9K9AqVhbP/t6Ny8XCuEx/WXq01jBfwcrP4BCFD7oOfK5V5Ah5ey
          11KzZ10tsUTlNXmqVFUr93tCy5Yf4iK8k0SJpxWPcoltzWCG1H5l/j5frlK8AbmD
          Yn3zbspD+UG7wEbZWM4SFXEDC4DpEOMtRaEDFZPXa5zcSnnCUxsFIbmEPun6Wk1t
          0qICrfqPjH98dOqfu5+IBb6EoCrTkp9t9ic9hrkdS5CYBvwZHJf+a5/rmte+xyix
          pvbI7bJ+Sw8Sx2E=
          -----END CERTIFICATE-----
        insecure-skip-tls-verify: false
---
apiVersion: v1
kind: Secret
metadata:
  name: loginapp-secret-env
  namespace: kube-system
type: Opaque
data:
  # original: ZXhhbXBsZS1hcHAtc2VjcmV0
  LOGINAPP_OIDC_CLIENT_SECRET: WlhoaGJYQnNaUzFoY0hBdGMyVmpjbVYw
EOF

cat <<EOF > ${CURR_DIR}/generated/loginapp-config-manual.yaml
---
name: "Kubernetes Auth"
listen: "0.0.0.0:8443"
oidc:
  client:
    id: "loginapp-local"
    secret: ZXhhbXBsZS1hcHAtc2VjcmV1
    redirectURL: "https://loginapp.127.0.0.1.nip.io:8443/callback"
  issuer:
    rootCA: "${CURR_DIR}/generated/ssl/ca.crt"
    url: "https://dex.${NODE_IP}.nip.io:32000"
tls:
  enabled: true
  cert: ${CURR_DIR}/generated/ssl/loginapp.crt
  key: ${CURR_DIR}/generated/ssl/key-loginapp.pem
log:
  level: Debug
  format: json
clusters:
  - name: myfakecluster
    server: https://myfakecluster.org
    certificate-authority: |
      -----BEGIN CERTIFICATE-----
      MIIDZzCCAk+gAwIBAgIRAM/Oqk7538VUzHwMZ+x7eN0wDQYJKoZIhvcNAQELBQAw
      FTETMBEGA1UEAxMKa3ViZXJuZXRlczAeFw0yMTExMDUxMDUyMDdaFw0yMjExMDUx
      MDUyMDdaMDcxFTATBgNVBAoTDHN5c3RlbTpub2RlczEeMBwGA1UEAwwVc3lzdGVt
      Om5vZGU6bG9naW5hcHA7MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA
      yc4mV0leq+fJ5KsF6yZTKtWRVzwcZGeYZv3gf1R9zSdexshOtBORRaNHJmuAu0st
      aHunVPlsSInBGBf4pO4OTLilu965/OQOvFvIFLzl3LNMoHOvUMofJ8lwBlcczil+
      UYf+xvDNYOYN3Q8WIBs5W2yHgDdFdfb9Z/X6ezL31T0/1+VYXR5cSawIKP9zEY/z
      miUWqrqRVZN0qvFYp5phMUbcIlNknEo/p0nUXfnz87P/7862BJuVaJVZmAMcg3IS
      q9Xl27P4NVnIvI5WPrlNrDNv3AupEJFS0c/Gms3nr6BZWNg5tM2mRhAr1Gfm37me
      fFqM0gxJn2sHjSF5y+M0jQIDAQABo4GPMIGMMA4GA1UdDwEB/wQEAwIFoDATBgNV
      HSUEDDAKBggrBgEFBQcDATAMBgNVHRMBAf8EAjAAMB8GA1UdIwQYMBaAFOw9vdeK
      hC8ALh7H+2QYpXPUWUU6MDYGA1UdEQQvMC2CFWRleC4xNzIuMTcuMC4yLm5pcC5p
      b4IUZGV4LjEyNy4wLjAuMS5uaXAuaW8wDQYJKoZIhvcNAQELBQADggEBAHgdgQ07
      YXERpAw8wRxxxX6Ozsod/vKww6E5Q4KmU8FxTqkx7hxQpp3neC6b5+yrsbGw2h0R
      8U9Kog+oNLtoN9K9AqVhbP/t6Ny8XCuEx/WXq01jBfwcrP4BCFD7oOfK5V5Ah5ey
      11KzZ10tsUTlNXmqVFUr93tCy5Yf4iK8k0SJpxWPcoltzWCG1H5l/j5frlK8AbmD
      Yn3zbspD+UG7wEbZWM4SFXEDC4DpEOMtRaEDFZPXa5zcSnnCUxsFIbmEPun6Wk1t
      0qICrfqPjH98dOqfu5+IBb6EoCrTkp9t9ic9hrkdS5CYBvwZHJf+a5/rmte+xyix
      pvbI7bJ+Sw8Sx2E=
      -----END CERTIFICATE-----
    insecure-skip-tls-verify: false
EOF

cat <<EOF > ${CURR_DIR}/helm/generated/overrides.yaml
replicas: 2
service:
  type: NodePort
  nodePort: 32001
args:
  - "-v"
config:
  secret: opdzojferfijcreoo
  clientID: loginapp
  clientSecret: ZXhhbXBsZS1hcHAtc2VjcmV0
  clientRedirectURL: "https://loginapp.${NODE_IP}.nip.io:32001/callback"
  issuerURL: "https://dex.${NODE_IP}.nip.io:32000"
  # Don't check issuer certificates (self-signed)
  issuerInsecureSkipVerify: true
  refreshToken: true
  tls:
    # This will generate a self-signed certificate
    enabled: true
    altnames:
      - "loginapp.172.17.0.2.nip.io"
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
configOverwrites:
  web:
    # "name": "admin", "email": admin@example.com
    mainUsernameClaim: "name"
dex:
  enabled: true
  fullnameOverride: dex
  replicas: 2
  https:
    enabled: true
  service:
    type: NodePort
    ports:
      https:
        nodePort: 32000
  volumes:
    - name: tls
      secret:
        secretName: dex-certs
  volumeMounts:
    - name: tls
      mountPath: /etc/dex/tls
  config:
    issuer: https://dex.${NODE_IP}.nip.io:32000
    storage:
      type: kubernetes
      config:
        inCluster: true
    web:
      https: 0.0.0.0:5554
      tlsCert: /etc/dex/tls/tls.crt
      tlsKey: /etc/dex/tls/tls.key
    oauth2:
      skipApprovalScreen: true
    staticClients:
    - id: loginapp
      redirectURIs:
      - 'https://loginapp.${NODE_IP}.nip.io:32001/callback'
      name: 'Loginapp Kube'
      secret: ZXhhbXBsZS1hcHAtc2VjcmV0
    - id: loginapp-local
      redirectURIs:
      - 'https://loginapp.127.0.0.1.nip.io:8443/callback'
      name: 'Loginapp local'
      secret: ZXhhbXBsZS1hcHAtc2VjcmV1

    enablePasswordDB: true
    staticPasswords:
    - email: "admin@example.com"
      # bcrypt hash of the string "password"
      hash: "a0b2cU8CPhOTaGrs1HRQuAueS7JTT5ZHsHSzYiFPm1leZck7Mc8T4W"
      username: "admin"
      userID: "08a8684b-db88-4b73-90a9-3cd1661f5466"
EOF

### Get Kubernetes certificate authority
echo "Get Kubernetes certificate authority (${CURR_DIR}/generated/ssl/ca.crt)"
kubectl config view --minify --flatten  -o jsonpath='{.clusters[0].cluster.certificate-authority-data}' | base64 -d > ${CURR_DIR}/generated/ssl/ca.crt
kubectl -n kube-system create configmap root-ca --from-file=ca.crt=${CURR_DIR}/generated/ssl/ca.crt 2>/dev/null

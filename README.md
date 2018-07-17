# Loginapp

[![Docker Repository on Quay](https://quay.io/repository/fydrah/loginapp/status "Docker Repository on Quay")](https://quay.io/repository/fydrah/loginapp)

**Simple application for Kubernetes CLI configuration with OIDC**

Original source code from [coreos/dex repository](https://github.com/coreos/dex/tree/master/cmd/example-app)

## Usage

```shell
NAME:
    loginapp - Simple application for Kubernetes CLI configuration with OIDC

AUTHOR:
    fydrah <flav.hardy@gmail.com>

USAGE:
    loginapp [global options] command [command options]

COMMANDS:
    serve    Run loginapp application
    help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
    --help, -h     show help
    --version, -v  print the version
```

## Configuration

```yaml
# AppName
name: "Kubernetes Auth"
# Bind IP and port (format: "IP:PORT")
listen: "0.0.0.0:5555"
# OIDC relative configuration
oidc:
  # Client configuration
  client:
    # Application ID
    id: "loginapp"
    # Application Secret
    secret: ZXhhbXBsZS1hcHAtc2VjcmV0
    # Application Redirect URL
    redirect_url: "https://127.0.0.1:5555/callback"
  # Issuer configuration
  issuer:
    # Location of issuer root CA certificate
    root_ca: "example/ssl/ca.pem"
    # Issuer URL
    url: "https://dex.example.com:5556"
  # Extra scopes
  extra_scopes:
    - groups
  # Enable offline scope
  offline_as_scope: true
  # Request token on behalf of other clients
  cross_clients: []
# Tls support
tls:
  # Enable tls termination
  enabled: true
  # Certificate location
  cert: example/ssl/cert.pem
  # Key location
  key: example/ssl/key.pem
# Logging configuration
log:
  # Loglevel: debug|info|warning|error|fatal|panic
  level: debug
  # Log format: json|text
  format: json
```

## Kubernetes

This application is built to run on a Kubernetes cluster. You will find usage examples here:
* Helm: [Helm chart](https://github.com/ObjectifLibre/k8s-ldap/tree/master/charts/k8s-ldap) is available on ObjectifLibre/k8s-ldap repository.

* Kubernetes: A full example is available on [ObjectifLibre/k8s-ldap repository](https://github.com/ObjectifLibre/k8s-ldap)

## Dev

* Setup Dex

```
  # Configure github oauth secrets if needed.
  # You must create an app in your github account before.
  cat <<EOF > dev.env
GITHUB_CLIENT_ID=yourclientid
GITHUB_CLIENT_SECRET=yoursecretid
EOF
  # Configure hosts entry
  echo "127.0.0.1 dex.example.com" | sudo tee -a /etc/hosts
  docker-compose up -d
```

  * User: admin@example.com
  * Password: password

* Manage dependencies

Loginapp uses [golang dep](https://golang.github.io/dep/docs/installation.html).

```
  # Update dependencies
  dep ensure
```

* Compile, configure and run

Configuration files are located in [example directory](./example/)

```
  make
  bin/loginapp serve example/config-loginapp.yaml
```

## Contibutions

Contributions (and issues) are welcomed.

I started this project to learn golang, so you will surely find some weird stuff. Please let me know if some code could be improved.

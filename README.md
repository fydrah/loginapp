# Loginapp

[![Docker Repository on Quay](https://quay.io/repository/fydrah/loginapp/status "Docker Repository on Quay")](https://quay.io/repository/fydrah/loginapp) [![codebeat badge](https://codebeat.co/badges/bb90084d-9b89-4af7-9a2c-150b7d4802da)](https://codebeat.co/projects/github-com-fydrah-loginapp-master) [![Codacy Badge](https://api.codacy.com/project/badge/Grade/0689fc84adb844cab356a625625ef54c)](https://www.codacy.com/app/fydrah/loginapp?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=fydrah/loginapp&amp;utm_campaign=Badge_Grade) [![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Ffydrah%2Floginapp.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Ffydrah%2Floginapp?ref=badge_shield)



**Web application for Kubernetes CLI configuration with OIDC**

Original source code from [coreos/dex repository](https://github.com/coreos/dex/tree/master/cmd/example-app)

## Usage

```shell
NAME:
    loginapp - Web application for Kubernetes CLI configuration with OIDC

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
# default: mandatory
name: "Kubernetes Auth"
# Bind IP and port (format: "IP:PORT")
# default: mandatory
listen: "0.0.0.0:5555"
# OIDC configuration
oidc:
  # Client configuration
  client:
    # Application ID
    # default: mandatory
    id: "loginapp"
    # Application Secret
    # default: mandatory
    secret: ZXhhbXBsZS1hcHAtc2VjcmV0
    # Application Redirect URL
    # default: mandatory
    redirect_url: "https://127.0.0.1:5555/callback"
  # Issuer configuration
  issuer:
    # Location of issuer root CA certificate
    # default: mandatory
    root_ca: "example/ssl/ca.pem"
    # Issuer URL
    # default: mandatory
    url: "https://dex.example.com:5556"
  # Extra scopes
  # default: []
  extra_scopes:
    - groups
  # Enable offline scope
  # default: false
  offline_as_scope: true
  # Request token on behalf of other clients
  # default: []
  cross_clients: []
# Tls support
tls:
  # Enable tls termination
  # default: false
  enabled: true
  # Certificate location
  # default: mandatory if tls.enabled is true
  cert: example/ssl/cert.pem
  # Key location
  # default: mandatory if tls.enabled is true
  key: example/ssl/key.pem
# Logging configuration
log:
  # Loglevel: debug|info|warning|error|fatal|panic
  # default: info
  level: debug
  # Log format: json|text
  # default: json
  format: json
# Configure the web behavior
web_output:
  # ClientID to output (useful for cross_client)
  # default: value of 'oidc.client.id'
  main_client_id: loginapp
  # Claims to use for kubeconfig username.
  # default: name
  main_username_claim: email
  # Assets directory
  # default: ${pwd}/assets
  assets_dir: /assets
  # Skip main page of login app
  # default: false
  skip_main_page: false
# Prometheus exporter configuration
prometheus:
  # Port to use. Metrics are available at
  # http://IP:PORT/metrics
  # default: 9090
  port: 9090
```

Two main examples are available:
* [Full configuration example](./example/config-loginapp-full.yaml) (each config option is set)
* [Minimal configuration example](./example/config-loginapp-minimal.yaml) (only mandatory options)

## Kubernetes

This application is built to run on a Kubernetes cluster. You will find usage examples here:
* Helm: [Helm chart](https://github.com/ObjectifLibre/k8s-ldap/tree/master/charts/k8s-ldap) is available on ObjectifLibre/k8s-ldap repository.

* Kubernetes: A full example is available on [ObjectifLibre/k8s-ldap repository](https://github.com/ObjectifLibre/k8s-ldap)

## Dev

###### Setup Dex

* (Optional) Configure GitHub OAuth App

```shell
  # Configure github oauth secrets if needed.
  # You must create an app in your github account before.
  cat <<EOF > dev.env
GITHUB_CLIENT_ID=yourclientid
GITHUB_CLIENT_SECRET=yoursecretid
EOF
```

* Configure host entry

```shell
  echo "127.0.0.1 dex.example.com" | sudo tee -a /etc/hosts
  docker-compose up -d
```

* User: admin@example.com
* Password: password

###### Manage dependencies

Loginapp uses [golang dep](https://golang.github.io/dep/docs/installation.html).

```shell
  # Update dependencies
  dep ensure
```

###### Compile, configure and run

Configuration files are located in [example directory](./example/)

```shell
  make
  bin/loginapp serve example/config-loginapp-full.yaml
```

You can also build a temporary Docker image for loginapp, and
run it with docker-compose (uncomment lines and replace image name):

```shell
  make docker-tmp
```

###### Run checks

Some checks can be launched before commits:
* errorcheck: check for unchecked errors
* gocyclo: cyclomatic complexities of functions
* gosimple: simplify code

```shell
  make checks
```

Run also gofmt before any new commit:

```shell
  make gofmt
```

## Contributions

Contributions (and issues) are welcomed.

I started this project to learn golang, so you will surely find some weird stuff. Please let me know if some code could be improved.

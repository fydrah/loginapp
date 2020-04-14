# Loginapp

[![Docker Repository on Quay](https://quay.io/repository/fydrah/loginapp/status "Docker Repository on Quay")](https://quay.io/repository/fydrah/loginapp) [![codebeat badge](https://codebeat.co/badges/bb90084d-9b89-4af7-9a2c-150b7d4802da)](https://codebeat.co/projects/github-com-fydrah-loginapp-master) [![Codacy Badge](https://api.codacy.com/project/badge/Grade/0689fc84adb844cab356a625625ef54c)](https://www.codacy.com/app/fydrah/loginapp?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=fydrah/loginapp&amp;utm_campaign=Badge_Grade) [![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Ffydrah%2Floginapp.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Ffydrah%2Floginapp?ref=badge_shield)

![Loginapp Demo](docs/img/demo.gif)


**Web application for Kubernetes CLI configuration with OIDC**

The code base of this repository use some source code from the original
[dexidp/dex repository](https://github.com/dexidp/dex/tree/master/cmd/example-app).

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
  # Extra auth code options
  # Some extra auth code options are required for ADFS compatibility (ex: resource).
  # See: https://docs.microsoft.com/fr-fr/windows-server/identity/ad-fs/overview/ad-fs-scenarios-for-developers
  # default: {}
  extra_auth_code_opts:
    resource: XXXXX
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
# Prometheus exporter configuration
prometheus:
  # Port to use. Metrics are available at
  # http://IP:PORT/metrics
  # default: 9090
  port: 9090
# Clusters list for CLI configuration
clusters:
  - name: mycluster
    server: https://mycluster.org
    certificate-authority: |
      -----BEGIN CERTIFICATE-----
      MIIC/zCCAeegAwIBAgIULkYvGJPRl50tMoVE4BNM0laRQncwDQYJKoZIhvcNAQEL
      BQAwDzENMAsGA1UEAwwEbXljYTAeFw0xOTAyMTgyMjA5NTJaFw0xOTAyMjgyMjA5
      NTJaMA8xDTALBgNVBAMMBG15Y2EwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK
      -----END CERTIFICATE-----
    insecure-skip-tls-verify: false
```

Two main examples are available:
* [Full configuration example](./example/config-loginapp-full.yaml) (each config option is set)
* [Minimal configuration example](./example/config-loginapp-minimal.yaml) (only mandatory options)

## Kubernetes

You have many ways to run loginapp:

* In a container:

  * On top of a Kubernetes cluster
  * As a standalone container

* Just with the binary:

  * As a daemon (systemd service)
  * Run binary for [development purpose](##Dev)

We advice to run loginapp on top of a kubernetes cluster, as a [Deployment](./deployments/kubernetes/as_deployment/) or a [DaemonSet](./deployments/kubernetes/as_daemonset/)

## Dev

###### Manage dependencies

Loginapp uses go modules to manage dependencies.

```shell
  # Retrieve dependencies (vendor)
  go mod vendor
```

###### Compile, configure and run

Configuration files are located in [example directory](./example/)

```shell
  $ make
```

Run also gofmt before any new commit:

```shell
  make gofmt
```

###### Dev env

Loginapp uses [kind](https://github.com/kubernetes-sigs/kind) and [skaffold](https://github.com/GoogleContainerTools/skaffold) for development environment.

Setup steps:

1. Launch a kind cluster:

    ```
    $ test/kubernetes/kindup.sh
    [...]
    Now you can run:

    test/kubernetes/genconf.sh 172.17.0.2
    ```

2. Generate Dex & Loginapp certificates and configuration for the dev env:

    ```
    # "172.17.0.2" is the IP of the kind control plane container
    # If you lost it, run:
    # "docker inspect loginapp-control-plane -f '{{ .NetworkSettings.Networks.bridge.IPAddress }}"
    $ test/kubernetes/genconf.sh 172.17.0.2
    [...]
    Creating TLS secret for loginapp
    Generating dex and loginapp configurations
    ```

3. Launch skaffold:

    ```
    $ skaffold run
    ```

4. To access loginapp UI, go to https://loginapp.${NODE_IP}.nip.io:32001, where **NODE_IP** is the IP of the kind control plane container. Default user/password configured by Dex is:

    * User: admin@example.com
    * Password: password

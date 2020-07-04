# loginapp

[Loginapp](https://github.com/fydrah/loginapp) - OIDC authentication helper for Kubernetes

## TL;DR;

```console
$ helm repo add fhardy-stable https://registry.fhardy.fr/chartrepo/stable
$ helm repo update
$ helm install loginapp fhardy-stable/loginapp -n auth
```

## Introduction

This chart deploys  on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites

- Identity Provider configuration (like Dex ou Keycloak)
- `config.secret` must be changed
- `config.clientID`, `config.clientSecret` and `config.clientRedirectURL` must be changed to conform with the application credentials configured on your identity provider
- `config.issuerURL` is the full URL before `/.well-known/openid-configuration` path (ex: for `https://dex.example.org/dex/.well-known/openid-configuration` it is `https://dex.example.org/dex`
- `config.issuerRootCA.configMap` must be created before deploying loginapp. The configMap must contain the plain issuer root CA in the key `config.issuerRootCA.key`
- If you setup an Ingress with SSL offloading, you must disable TLS for loginapp with `config.tls.enabled: false`

## Installing the Chart

To install the chart with the release name `loginapp`:

```console
$ helm install loginapp fhardy-stable/loginapp -n auth
```

The command deploys  on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the `loginapp`:

```console
$ helm delete loginapp -n auth
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the `loginapp` chart and their default values.

|            Parameter            |                                                                                                                         Description                                                                                                                         |               Default               |
|---------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------|
| replicas                        |                                                                                                                                                                                                                                                             | `1`                                 |
| image                           |                                                                                                                                                                                                                                                             | `quay.io/fydrah/loginapp:v3.1.0`    |
| imagePullPolicy                 |                                                                                                                                                                                                                                                             | `IfNotPresent`                      |
| imagePullSecrets                |                                                                                                                                                                                                                                                             | `[]`                                |
| nameOverride                    |                                                                                                                                                                                                                                                             | `""`                                |
| fullnameOverride                |                                                                                                                                                                                                                                                             | `""`                                |
| serviceAccount.create           | Specifies whether a service account should be created                                                                                                                                                                                                       | `true`                              |
| serviceAccount.annotations      | Annotations to add to the service account                                                                                                                                                                                                                   | `{}`                                |
| serviceAccount.name             | The name of the service account to use. If not set and create is true, a name is generated using the fullname template                                                                                                                                      | ``                                  |
| podSecurityContext              |                                                                                                                                                                                                                                                             | `{}`                                |
| securityContext                 |                                                                                                                                                                                                                                                             | `{}`                                |
| service.type                    |                                                                                                                                                                                                                                                             | `ClusterIP`                         |
| service.port                    |                                                                                                                                                                                                                                                             | `5555`                              |
| service.nodePort                |                                                                                                                                                                                                                                                             | ``                                  |
| service.loadBalancerIP          |                                                                                                                                                                                                                                                             | ``                                  |
| ingress.enabled                 |                                                                                                                                                                                                                                                             | `false`                             |
| ingress.annotations             |                                                                                                                                                                                                                                                             | `{}`                                |
| ingress.tls                     |                                                                                                                                                                                                                                                             | `[]`                                |
| resources                       |                                                                                                                                                                                                                                                             | `{}`                                |
| nodeSelector                    |                                                                                                                                                                                                                                                             | `{}`                                |
| tolerations                     |                                                                                                                                                                                                                                                             | `[]`                                |
| affinity                        |                                                                                                                                                                                                                                                             | `{}`                                |
| env                             | Additionnal env vars <br> Example: <br> `LOGINAPP_XXXXXX: "value"`                                                                                                                                                                                          | `{}`                                |
| args                            | Additional args <br> Example: <br> `- "-v" # This is for debug logs`                                                                                                                                                                                        | `[]`                                |
| config.name                     | Application name, defaults to Release name                                                                                                                                                                                                                  | ``                                  |
| config.secret                   | Application secret if empty, generate a random string please setup a real secret otherwise helm will generate a new secret at each deployment                                                                                                               | ``                                  |
| config.clientID                 | OIDC Client ID                                                                                                                                                                                                                                              | `"loginapp"`                        |
| config.clientSecret             | OIDC Client secret                                                                                                                                                                                                                                          | ``                                  |
| config.clientRedirectURL        | OIDC Client redirect URL This must end with /callback if empty, defaults to: 1. '{{ .Values.ingress.hosts[0].host }}/callback' if 'ingress.enabled: true' and 'ingress.hosts[0]' exists 2. '{{ .Release.Name }}.{{ .Release.Namespace }}.svc:5555/callback' | ``                                  |
| config.issuerRootCA             | Issuer root CA configMap ConfigMap containing the root CA and key to use inside the configMap. This configMap must exist                                                                                                                                    | `{"configMap":null,"key":"ca.crt"}` |
| config.issuerInsecureSkipVerify | Skip issuer certificate validation This is usefull for testing purpose, but not recommended in production                                                                                                                                                   | `false`                             |
| config.issuerURL                | Issuer url                                                                                                                                                                                                                                                  | `"https://dex.example.org:32000"`   |
| config.refreshToken             | Include refresh token in request                                                                                                                                                                                                                            | `false`                             |
| config.tls.enabled              | Enable TLS for deployment                                                                                                                                                                                                                                   | `true`                              |
| config.tls.secretName           | Secret name where certificates are stored if empty and 'tls.enabled: true', generate self signed certificates if not empty, use the kubernetes secret 'secretName' (type: kubernetes.io/tls)                                                                | ``                                  |
| config.tls.altnames             | Self singed certificat DNS names <br> Example: <br> `- loginapp.172.17.0.2.nip.io`                                                                                                                                                                          | `[]`                                |
| config.tls.altIPs               | Self signed certificat IPs                                                                                                                                                                                                                                  | `[]`                                |
| config.clusters                 | List of kubernetes clusters to add on web frontend                                                                                                                                                                                                          | `[]`                                |
| configOverwrites                | Configuration overrides, this is a free configuration merged with the previous generated configuration 'config'. Use this to add or overwrites values. <br> Example: <br> `oidc:` <br> `extraScopes:` <br> `- groups`                                       | `{}`                                |
| dex.enabled                     |                                                                                                                                                                                                                                                             | `false`                             |


Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example:

```console
$ helm install loginapp fhardy-stable/loginapp -n auth --set replicas=1
```

Alternatively, a YAML file that specifies the values for the parameters can be provided while
installing the chart. For example:

```console
$ helm install loginapp fhardy-stable/loginapp -n auth --values values.yaml
```

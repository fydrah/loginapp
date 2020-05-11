# Prepare your deployment

## Identity provider (IdP)

Several identity providers exist:

- [Dex](https://github.com/dexidp/dex/)
- [Keycloak](https://www.keycloak.org/)
- [Microsoft ADFS](https://docs.microsoft.com/en-us/windows-server/identity/ad-fs/development/ad-fs-openid-connect-oauth-concepts)
- ...


The first step is to choose the IdP Loginapp will work with.

If you do not already have an IdP, the Loginapp helm charts suggest the deployment of Dex through its [official chart](https://github.com/helm/charts/tree/master/stable/dex).


### (optional) Dex chart usage with Loginapp chart

There is a dedicated section for Dex in [the chart values](../helm/loginapp/values.yaml):

```yaml
dex:
  enabled: false
```

Use it to configure Dex deployment.


The configuration requirements for Loginapp are:

- A client ID and Secret:

    ```yaml
    dex:
      [...]
      config:
        staticClients:
        - id: loginapp
          redirectURIs:
          - 'https://loginapp.172.17.0.2.nip.io:32001/callback'
          name: 'Loginapp Kube'
          secret: ZXhhbXBsZS1hcHAtc2VjcmV0
      [...]
    ```

- Issuer redirect URL in `HTTPS`:

    ```yaml
    dex:
      [...]
      config:
        # Must start with https
        issuer: https://dex.172.17.0.2.nip.io:32000
    ```

   Note: the issuer needs to be in `HTTPS`, but it is up to you to configure Dex deployment in `HTTPS`, or to use an Ingress for SSL offloading in front of Dex deployment.


### Required configuration from IdP

If you do not setup Dex yourself, or if you use an other IdP, your must retrieve the following informations:

- **Client ID**: used later for `config.clientID`
- **Client secret**: used later for `config.clientSecret`
- **Issuer URL**: used later for `config.issuerURL`. This is the full URL just before the path `/.well-known/openid-configuration`
- **Issuer root CA**: used later for `config.issuerRootCA`. This is the CA who signed the issuer certificate

## Kubernetes

To enable OIDC authentication on your Kubernetes cluster you must to configure the API Server with these options:
- `--oidc-issuer-url`: same value as `config.issuerURL`
- `--oidc-client-id`:
    * Same value as `config.clientID`, or...
    * Use cross-client authentication (ex: client ID `kubernetes` for the cluster, and configure the client ID `loginapp` to issue token on the behalf of `kubernetes`. See https://github.com/dexidp/dex/blob/master/Documentation/custom-scopes-claims-clients.md#cross-client-trust-and-authorized-party)
- `--oidc-username-claim`: same value as `configOverwrites.web.mainUsernameClaim`, Loginapp use `email` as default, so fill `email` for this option of change `configOverwrites.web.mainUsernameClaim` to the claim you want.


More information available at https://kubernetes.io/docs/reference/access-authn-authz/authentication/#configuring-the-api-server

## Loginapp

There are many ways to configure Loginapp on top of Kubernetes:

|   | deployment |        service        | ingress | Certificate(s)       |
|---|:----------:|:---------------------:|:-------:|----------------------|
| 1 | HTTP       | NodePort/LoadBalancer | N/A     | N/A                  |
| 2 | HTTPS      | NodePort/LoadBalancer | N/A     | Self-signed          |
| 3 | HTTPS      | NodePort/LoadBalancer | N/A     | Custom               |
| 4 | HTTP       | ClusterIP             | HTTP    | N/A                  |
| 5 | HTTP       | ClusterIP             | HTTPS   | Custom               |
| 6 | HTTPS      | ClusterIP             | HTTPS   | Self-signed / Custom |

### Deployment 1

```yaml
service:
  type: NodePort # Or LoadBalancer
  nodePort: 32001 # If type NodePort

ingress: 
  enabled: false

config:
  tls:
    enabled: false
  clientRedirectURL: http://loginapp.example.org:32001/callback
```


### Deployment 2

```yaml
service:
  type: NodePort # Or LoadBalancer
  nodePort: 32001 # If type NodePort

ingress: 
  enabled: false

config:
  tls:
    enabled: true
    altNames:
    - loginapp.example.org:32001
  clientRedirectURL: https://loginapp.example.org:32001/callback
```

### Deployment 3

```yaml
service:
  type: NodePort # Or LoadBalancer
  nodePort: 32001 # If type NodePort

ingress: 
  enabled: false

config:
  tls:
    enabled: true
    secretName: loginapp-tls # This secret (kubernetes.io/tls) must exist
  clientRedirectURL: https://loginapp.example.org:32001/callback
```

### Deployment 4

```yaml
service:
  type: ClusterIP

ingress: 
  enabled: true
  hosts:
  - host: loginapp.example.org
    paths: [/]

config:
  tls:
    enabled: false
  clientRedirectURL: http://loginapp.example.org/callback
```

### Deployment 5

```yaml
service:
  type: ClusterIP

ingress: 
  enabled: true
  hosts:
  - host: loginapp.example.org
    paths: [/]
  tls:
  - secretName: loginapp-tls # This secret (kubernetes.io/tls) must exist (or use Letsencrypt)
    
config:
  tls:
    enabled: false
  clientRedirectURL: https://loginapp.example.org/callback
```

### Deployment 6

```yaml
service:
  type: ClusterIP

ingress: 
  enabled: true
  annotations:
    nginx.ingress.kubernetes.io/backend-protocol: "HTTPS" # If you use nginx ingress
  hosts:
  - host: loginapp.example.org
    paths: [/]
  tls:
  - secretName: loginapp-tls # This secret (kubernetes.io/tls) must exist (or use Letsencrypt)
    
config:
  tls:
    enabled: true
  clientRedirectURL: https://loginapp.example.org/callback
```

### Required configurations

Here are the required configuration for Loginapp:

- `config.secret`: generate a secure secret for the application (this is **NOT** the OIDC client Secret)
- `config.clientID`: see [required configuration from IdP](#Required-configuration-from-IdP) 
- `config.clientSecret`: see [required configuration from IdP](#Required-configuration-from-IdP)
- `config.clientRedirectURL`: this value depends on the deployment type you use. Must end with `/callback`
- `config.issuerURL`: see [required configuration from IdP](#Required-configuration-from-IdP)
- `config.issuerRootCA` (or `config.issuerInsecureSkipVerify` for testing)

### Configuration overwrite 

Use the configuration key `configOverwrites` to overwrite generated configuration (see: [Configuration](../README.md#Configuration)):

```yaml
configOverwrite:
  oidc:
    issuer:
      insecureSkipVerify: true
    crossClients:
      - kubernetes
  web:
    mainClientID: kubernetes 
```

### Certificates

Self-signed certificates are valid for `365` days. By default, it includes the Kubernetes service DNS entries (`SVCNAME` & `SVCNAME.NAMESPACE.svc`)


### Refresh token

Loginapp can ask the issuer to include a refresh token to the response. This meens your user will be able to ask for a new token with it (depends on the refresh token TTL), without requesting Loginapp.

But, this meens your users **will have access to the client ID and client secret** used by Loginapp.

This is a potential security issue (see: https://github.com/kubernetes/kubernetes/issues/37822).


Configure refresh token request:

```yaml
config:
  refreshToken: true
```

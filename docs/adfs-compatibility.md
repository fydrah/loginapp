# ADFS compatibility

## Resource URL

ADFS accepts a resource URL as a part of authentication request flow, it is used to identify your Web API.

To add resource URL on the request, use the OIDC extra auth code options config :

```yaml
config:
  oidc:
    extra:
      authCodeOpts:
        resource: xxxxxx
```


If you require the resource URL to be included in Kubeconfig (ex: for refresh tokens), update the Kubeconfig configuration part:

```yaml
config:
  web:
    kubeconfig:
      extraOpts:
        resource: xxxxxx
```

This will automatically add the extra options to the generated Kubeconfig and kubectl command:

```yaml
- name: admin@example.com
  user:
    auth-provider:
      config:
        resource: xxxxx         # added here
        client-id: loginapp
        [...]
```


For more informations:
* ADFS: https://docs.microsoft.com/en-us/windows-server/identity/ad-fs/overview/ad-fs-openid-connect-oauth-flows-scenarios
* Issue: https://github.com/fydrah/loginapp/issues/16
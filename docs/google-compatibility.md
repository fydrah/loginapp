# Google compatibility

## Refresh token

Google openid-connect API supports refresh token, but not if requested as a scope.

You must configure an extra option in request. To do so, add this to Loginapp configuration file:

```yaml
config:
  oidc:
    extra:
      authCodeOpts:
        access_type: offline
        prompt: consent
```


These options will be added to the authentication request.


For more informations:
* Google OpenID documentation: https://developers.google.com/identity/protocols/oauth2/openid-connect
* Issue: https://github.com/fydrah/loginapp/issues/32
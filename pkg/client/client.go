package client

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	oidc "github.com/coreos/go-oidc"
	"github.com/fydrah/loginapp/pkg/config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

var c *Client

// Client is an OpenID client, it handles all OIDC/OAuth2 interactions
// between the provider and the creator of this Client
type Client struct {
	Config     *config.OIDC
	Provider   *oidc.Provider
	Verifier   *oidc.IDTokenVerifier
	Scopes     []string
	HTTPClient *http.Client
}

func New(cfg *config.OIDC) *Client {
	c = new(Client)
	c.Config = cfg
	c.PrepareScopes()
	return c
}

// OAuth2Config return the OAuth2Config for the client
func (c *Client) OAuth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     c.Config.Client.ID,
		ClientSecret: c.Config.Client.Secret,
		RedirectURL:  c.Config.Client.RedirectURL,
		Endpoint:     c.Provider.Endpoint(),
		Scopes:       c.Scopes,
	}
}

// AuthCodeURL generate an authorisation code URL based on the application
// name. The function uses also extra auth code options configured for
// this client
func (c *Client) AuthCodeURL(r *http.Request, secret string) string {
	var (
		extraAuthCodeOptions []oauth2.AuthCodeOption
		authCodeURL          string
	)
	for p, v := range c.Config.Extra.AuthCodeOpts {
		extraAuthCodeOptions = append(extraAuthCodeOptions, oauth2.SetAuthURLParam(p, v))
	}
	// We should comply to https://tools.ietf.org/html/rfc6749#section-10.12 for
	// CSRF protection. This is a temporary fix until we find a better way
	// Currently don't know if it can be achieved without session affinity
	// or external storage (memcached, redis..)
	authCodeURL = c.OAuth2Config().AuthCodeURL(GenerateState(r, secret), extraAuthCodeOptions...)
	log.Debugf("auth code url: %s", authCodeURL)
	log.Debugf("request token with the following scopes: %v", c.Scopes)
	return authCodeURL
}

// GenerateState performs a base64 encoded hash of the client user agent
// This is a poor way to permform CSRF protection, but better than a hardcoded
// value.
func GenerateState(r *http.Request, secret string) string {
	csum := sha256.Sum256([]byte(r.UserAgent() + secret))
	return base64.StdEncoding.EncodeToString(csum[:])
}

// VerifyState performs verification for CSRF
func VerifyState(r *http.Request, s string, secret string) bool {
	state := GenerateState(r, secret)
	return s == state
}

// AuthCodeToToken converts an authorization code into a IDToken
func (c *Client) AuthCodeToIDToken(ctx context.Context, authCode string) (*oauth2.Token, string, *oidc.IDToken, error) {
	clientCtx := c.Context()
	token, err := c.OAuth2Config().Exchange(clientCtx, authCode)
	if err != nil {
		log.Errorf("token exchange failed with context %v and authCode %v", clientCtx, authCode)
		return nil, "", nil, err
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, "", nil, fmt.Errorf("no id_token in token response")
	}
	idToken, vErr := c.Verifier.Verify(ctx, rawIDToken)
	if vErr != nil {
		return nil, "", nil, err
	}
	return token, rawIDToken, idToken, nil
}

// ExtractClaims returns claims for a given IDToken
func ExtractClaims(t *oidc.IDToken) (map[string]interface{}, error) {
	var (
		claims     json.RawMessage
		jsonClaims map[string]interface{}
	)
	// get claims from id token
	if err := t.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims from idToken: %v", err)
	}
	// format json claims output as a proper json object
	buff := new(bytes.Buffer)
	if err := json.Indent(buff, []byte(claims), "", "  "); err != nil {
		return nil, fmt.Errorf("failed to format claims output: %v", err)
	}
	// unmarshal json data into claim interface
	if err := json.Unmarshal(claims, &jsonClaims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %v", err)
	}
	return jsonClaims, nil
}

// Context returns Client context
func (c *Client) Context() context.Context {
	return oidc.ClientContext(context.Background(), c.HTTPClient)
}

// PrepareScopes setup scopes slice based on the client configuration
func (c *Client) PrepareScopes() {
	c.Scopes = append(c.Scopes, c.Config.Scopes...)

	if len(c.Config.Extra.Scopes) > 0 {
		log.Warning("[DEPRECATED] Using 'oidc.extra.scopes' is deprecated, please use 'oidc.scopes' instead to override default oidc scopes")
	}
	c.Scopes = append(c.Scopes, c.Config.Extra.Scopes...)
	// Prepare cross client auth
	// see https://github.com/coreos/dex/blob/master/Documentation/custom-scopes-claims-clients.md
	for _, crossClient := range c.Config.CrossClients {
		c.Scopes = append(c.Scopes, "audience:server:client_id:"+crossClient)
	}

	if c.Config.OfflineAsScope {
		c.Scopes = append(c.Scopes, "offline_access")
	}
}

// ProviderSetup setup Client's provider
func (c *Client) ProviderSetup() error {
	if err := backoff.Retry(func() error {
		var bErr error
		if c.Provider, bErr = oidc.NewProvider(c.Context(), c.Config.Issuer.URL); bErr != nil {
			log.Errorf("failed to query provider %q: %v", c.Config.Issuer.URL, bErr)
			return bErr
		}
		return nil
	}, backoff.NewExponentialBackOff()); err != nil {
		return err
	}
	var ss struct {
		// What scopes does a provider support?
		//
		// See: https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderMetadata
		ScopesSupported []string `json:"scopes_supported"`
	}
	if err := c.Provider.Claims(&ss); err != nil {
		return fmt.Errorf("failed to parse provider scopes_supported: %v", err)
	}

	// Ugly. Should be moved to an other place, and should comply
	// go-oidc doc: go doc go-oidc.ScopeOfflineAccess
	if !c.Config.OfflineAsScope {
		if len(ss.ScopesSupported) > 0 {
			// See if scopes_supported has the "offline_access" scope.
			c.Config.OfflineAsScope = func() bool {
				for _, scope := range ss.ScopesSupported {
					if scope == oidc.ScopeOfflineAccess {
						return true
					}
				}
				return false
			}()
		}
	}
	return nil
}

// VerifierSetup setup Client's verifier
func (c *Client) VerifierSetup() {
	c.Verifier = c.Provider.Verifier(&oidc.Config{
		ClientID: c.Config.Client.ID,
	})
}

// TLSSetup setup tls transport for the client
func (c *Client) TLSSetup() error {
	tlsConfig := tls.Config{RootCAs: x509.NewCertPool(), InsecureSkipVerify: c.Config.Issuer.InsecureSkipVerify}
	if !c.Config.Issuer.InsecureSkipVerify {
		rootCABytes, err := ioutil.ReadFile(c.Config.Issuer.RootCA)
		if err != nil {
			return fmt.Errorf("failed to read root-ca: %v", err)
		}
		if !tlsConfig.RootCAs.AppendCertsFromPEM(rootCABytes) {
			return fmt.Errorf("no certs found in root CA file %q", c.Config.Issuer.RootCA)
		}
	}
	c.HTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tlsConfig,
			Proxy:           http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	return nil
}

// Setup setups Client
func (c *Client) Setup() error {
	if err := c.TLSSetup(); err != nil {
		return err
	}
	if err := c.ProviderSetup(); err != nil {
		return err
	}
	c.VerifierSetup()
	return nil
}

// Healthz reports if the client is ready to perform
// requests to the issuer
func (c *Client) Healthz() bool {
	if c.Provider == nil {
		log.Debug("provider not ready")
		return false
	}
	wellKnown := strings.TrimSuffix(c.Config.Issuer.URL, "/") + "/.well-known/openid-configuration"
	if _, err := c.HTTPClient.Head(wellKnown); err != nil {
		return false
	}
	return true
}

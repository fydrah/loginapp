package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coreos/go-oidc"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

const appState = "Dex login request"

type app struct {
	ClientID        string `yaml:"client_id"`
	ClientSecret    string `yaml:"client_secret"`
	RedirectURI     string `yaml:"redirect_url"`
	InitExtraScopes string `yaml:"extra_scopes"`
	DisableChoices  bool   `yaml:"disable_choices"`
	AppName         string `yaml:"app_name"`
	IssuerURL       string `yaml:"issuer_url"`
	RootCAs         string `yaml:"issuer_root_ca"`
	Listen          string `yaml:"listen"`
	TlsCert         string `yaml:"tls_cert"`
	TlsKey          string `yaml:"tls_key"`
	Debug           bool   `yaml:"debug"`

	verifier *oidc.IDTokenVerifier
	provider *oidc.Provider

	// Does the provider use "offline_access" scope to request a refresh token
	// or does it use "access_type=offline" (e.g. Google)?
	offlineAsScope bool

	client *http.Client
}

// return an HTTP client which trusts the provided root CAs.
func httpClientForRootCAs(rootCAs string) (*http.Client, error) {
	tlsConfig := tls.Config{RootCAs: x509.NewCertPool()}
	rootCABytes, err := ioutil.ReadFile(rootCAs)
	if err != nil {
		return nil, fmt.Errorf("failed to read root-ca: %v", err)
	}
	if !tlsConfig.RootCAs.AppendCertsFromPEM(rootCABytes) {
		return nil, fmt.Errorf("no certs found in root CA file %q", rootCAs)
	}
	return &http.Client{
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
	}, nil
}

type debugTransport struct {
	t http.RoundTripper
}

func (d debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return nil, err
	}
	log.Printf("%s", reqDump)

	resp, err := d.t.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}
	log.Printf("%s", respDump)
	return resp, nil
}

func cmd() *cobra.Command {
	var (
		a          app
	)
	c := cobra.Command{
		Use:   "login-app",
		Short: "Kubernetes login OIDC app",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("Missing configuration file path")
			}
			configData, err := ioutil.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("failed to read config file %s: %v", args[0], err)
			}
			if err := yaml.Unmarshal(configData, &a); err != nil {
				return fmt.Errorf("error parse config file %s: %v", args[0], err)
			}
			u, err := url.Parse(a.RedirectURI)
			if err != nil {
				return fmt.Errorf("parse redirect-uri: %v", err)
			}
			listenURL, err := url.Parse(a.Listen)
			if err != nil {
				return fmt.Errorf("parse listen address: %v", err)
			}

			if a.RootCAs != "" {
				client, err := httpClientForRootCAs(a.RootCAs)
				if err != nil {
					return err
				}
				a.client = client
			}

			if a.Debug {
				if a.client == nil {
					a.client = &http.Client{
						Transport: debugTransport{http.DefaultTransport},
					}
				} else {
					a.client.Transport = debugTransport{a.client.Transport}
				}
			}

			if a.client == nil {
				a.client = http.DefaultClient
			}
			// TODO(ericchiang): Retry with backoff
			ctx := oidc.ClientContext(context.Background(), a.client)
			provider, err := oidc.NewProvider(ctx, a.IssuerURL)
			if err != nil {
				return fmt.Errorf("Failed to query provider %q: %v", a.IssuerURL, err)
			}

			var s struct {
				// What scopes does a provider support?
				//
				// See: https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderMetadata
				ScopesSupported []string `json:"scopes_supported"`
			}
			if err := provider.Claims(&s); err != nil {
				return fmt.Errorf("Failed to parse provider scopes_supported: %v", err)
			}

			if len(s.ScopesSupported) == 0 {
				// scopes_supported is a "RECOMMENDED" discovery claim, not a required
				// one. If missing, assume that the provider follows the spec and has
				// an "offline_access" scope.
				a.offlineAsScope = true
			} else {
				// See if scopes_supported has the "offline_access" scope.
				a.offlineAsScope = func() bool {
					for _, scope := range s.ScopesSupported {
						if scope == oidc.ScopeOfflineAccess {
							return true
						}
					}
					return false
				}()
			}

			a.provider = provider
			a.verifier = provider.Verifier(&oidc.Config{ClientID: a.ClientID})

			http.HandleFunc("/", a.handleIndex)
			http.HandleFunc("/login", a.handleLogin)
			http.HandleFunc(u.Path, a.handleCallback)

			switch listenURL.Scheme {
			case "http":
				log.Printf("listening on %s", a.Listen)
				return http.ListenAndServe(listenURL.Host, nil)
			case "https":
				log.Printf("listening on %s", a.Listen)
				return http.ListenAndServeTLS(listenURL.Host, a.TlsCert, a.TlsKey, nil)
			default:
				return fmt.Errorf("listen address %q is not using http or https", a.Listen)
			}
		},
	}
	return &c
}

func main() {
	if err := cmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}
}

func (a *app) handleIndex(w http.ResponseWriter, r *http.Request) {
	renderIndex(w, a.InitExtraScopes, a.DisableChoices, a.AppName)
}

func (a *app) oauth2Config(scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     a.ClientID,
		ClientSecret: a.ClientSecret,
		Endpoint:     a.provider.Endpoint(),
		Scopes:       scopes,
		RedirectURL:  a.RedirectURI,
	}
}

//func (a *app) handleLogin(w http.ResponseWriter, r *http.Request) {
//	var scopes []string
//	authCodeURL := ""
//	scopes = append(scopes, "openid", "profile", "email", "groups")
//	authCodeURL = a.oauth2Config(scopes).AuthCodeURL(appState, oauth2.AccessTypeOffline)
//	http.Redirect(w, r, authCodeURL, http.StatusSeeOther)
//}

func (a *app) handleLogin(w http.ResponseWriter, r *http.Request) {
	var scopes []string
	if a.InitExtraScopes != "" {
		for _, scope := range strings.Split(a.InitExtraScopes, ",") {
			scopes = append(scopes, scope)
		}
	}
	if extraScopes := r.FormValue("extra_scopes"); extraScopes != "" {
		for _, scope := range strings.Split(extraScopes, ",") {
			scopes = append(scopes, scope)
		}
	}
	var clients []string
	if crossClients := r.FormValue("cross_client"); crossClients != "" {
		clients = strings.Split(crossClients, ",")
	}
	for _, client := range clients {
		scopes = append(scopes, "audience:server:client_id:"+client)
	}

	authCodeURL := ""
	scopes = append(scopes, "openid", "profile", "email", "groups")
	if r.FormValue("offline_access") != "yes" {
		authCodeURL = a.oauth2Config(scopes).AuthCodeURL(appState)
	} else if a.offlineAsScope {
		scopes = append(scopes, "offline_access")
		authCodeURL = a.oauth2Config(scopes).AuthCodeURL(appState)
	} else {
		authCodeURL = a.oauth2Config(scopes).AuthCodeURL(appState, oauth2.AccessTypeOffline)
	}
	log.Printf("%s",a.oauth2Config)
	http.Redirect(w, r, authCodeURL, http.StatusSeeOther)
}

func (a *app) handleCallback(w http.ResponseWriter, r *http.Request) {
	var (
		err   error
		token *oauth2.Token
	)

	ctx := oidc.ClientContext(r.Context(), a.client)
	oauth2Config := a.oauth2Config(nil)
	switch r.Method {
	case "GET":
		// Authorization redirect callback from OAuth2 auth flow.
		if errMsg := r.FormValue("error"); errMsg != "" {
			http.Error(w, errMsg+": "+r.FormValue("error_description"), http.StatusBadRequest)
			return
		}
		code := r.FormValue("code")
		if code == "" {
			http.Error(w, fmt.Sprintf("no code in request: %q", r.Form), http.StatusBadRequest)
			return
		}
		if state := r.FormValue("state"); state != appState {
			http.Error(w, fmt.Sprintf("expected state %q got %q", appState, state), http.StatusBadRequest)
			return
		}
		token, err = oauth2Config.Exchange(ctx, code)
	case "POST":
		// Form request from frontend to refresh a token.
		refresh := r.FormValue("refresh_token")
		if refresh == "" {
			http.Error(w, fmt.Sprintf("no refresh_token in request: %q", r.Form), http.StatusBadRequest)
			return
		}
		t := &oauth2.Token{
			RefreshToken: refresh,
			Expiry:       time.Now().Add(-time.Hour),
		}
		token, err = oauth2Config.TokenSource(ctx, t).Token()
	default:
		http.Error(w, fmt.Sprintf("method not implemented: %s", r.Method), http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get token: %v", err), http.StatusInternalServerError)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "no id_token in token response", http.StatusInternalServerError)
		return
	}

	idToken, err := a.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to verify ID token: %v", err), http.StatusInternalServerError)
		return
	}
	var claims json.RawMessage
	idToken.Claims(&claims)

	buff := new(bytes.Buffer)
	json.Indent(buff, []byte(claims), "", "  ")

	renderToken(w, a.RedirectURI, rawIDToken, token.RefreshToken, buff.Bytes(), a.ClientSecret)
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/go-oidc"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

/**
 * Type def
 */
type Server struct {
	client		*http.Client
	config		AppConfig
	provider	*oidc.Provider
	router		*httprouter.Router
	verifier	*oidc.IDTokenVerifier
	context		context.Context
}

/**
 * Render
 */
func (s *Server) renderIndex(w http.ResponseWriter) {
	renderTemplate(w, indexTmpl, s.config)
}

/**
 * OpenID
 */
func (s *Server) OAuth2Config(scopes []string, endpoint string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     s.config.ClientID,
		ClientSecret: s.config.ClientSecret,
		RedirectURL:  s.config.RedirectURL+"/"+endpoint,
		Endpoint:     s.provider.Endpoint(),
		Scopes:       scopes,
	}
}

func (s *Server) PrepareCallbackUrl(endpoint string) string {
	// Prepare scopes
	var (
		scopes []string
		authCodeURL string
	)
	if s.config.ExtraScopes != "" {
		for _, scope := range strings.Split(s.config.ExtraScopes, ",") {
			scopes = append(scopes, scope)
		}
	}
	// Prepare cross client auth
	// see https://github.com/coreos/dex/blob/master/Documentation/custom-scopes-claims-clients.md
	var clients []string
	if s.config.CrossClients != "" {
		clients = strings.Split(s.config.CrossClients, ",")
	}
	for _, client := range clients {
		scopes = append(scopes, "audience:server:client_id:"+client)
	}

	scopes = append(scopes, "openid", "profile", "email", "groups")
	if s.config.OfflineAsScope {
		scopes = append(scopes, "offline_access")
		authCodeURL = s.OAuth2Config(scopes, endpoint).AuthCodeURL(s.config.AppName)
	} else if !s.config.OfflineAsScope {
		authCodeURL = s.OAuth2Config(scopes, endpoint).AuthCodeURL(s.config.AppName)
	} else {
		authCodeURL = s.OAuth2Config(scopes, endpoint).AuthCodeURL(s.config.AppName, oauth2.AccessTypeOffline)
	}
	return authCodeURL
}

/**
 * Handlers
 */
func (s *Server) HandleGetIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	s.renderIndex(w)
}

func (s *Server) HandlePostLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	switch r.FormValue("access") {
	case "Dashboard":
		http.Redirect(w, r, s.PrepareCallbackUrl("dashboard"), http.StatusSeeOther)
		// Set Cookie
		// Then redirect to Dashboard Proxy
		// Set Header with Token: Bearer XXXXXXXX
	case "CLI":
		http.Redirect(w, r, s.PrepareCallbackUrl("cli"), http.StatusSeeOther)
	}
}

func (s *Server) HandlePostCLI(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}

func (s *Server) HandlePostDashboard(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}

func (s *Server) HandleGetCallbackDashboard(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
}

func (s *Server) HandleGetCallbackCLI(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var (
		err   error
		token *oauth2.Token
	)
	oauth2Config := s.OAuth2Config(nil, "cli")
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
	if state := r.FormValue("state"); state != s.config.AppName {
		http.Error(w, fmt.Sprintf("expected state %q got %q", s.config.AppName, state), http.StatusBadRequest)
		return
	}
	token, err = oauth2Config.Exchange(s.context, code)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get token: %v", err), http.StatusInternalServerError)
		return
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "no id_token in token response", http.StatusInternalServerError)
		return
	}
	idToken, err := s.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to verify ID token: %v", err), http.StatusInternalServerError)
		return
	}
	var claims json.RawMessage
	idToken.Claims(&claims)

	buff := new(bytes.Buffer)
	json.Indent(buff, []byte(claims), "", "  ")
	//renderToken(w, a.RedirectURI, rawIDToken, token.RefreshToken, buff.Bytes(), a.ClientSecret)
}

/**
 * Run
 */
func (s *Server) Run() error {
	// router setup
	s.router = httprouter.New()
	s.Routes()
	// client setup
	if s.client == nil {
		client, err := httpClientForRootCAs(s.config.IssuerRootCA)
		if err != nil {
			return err
		}
		s.client = client
	}
	// OIDC setup
	// TODO(ericchiang): Retry with backoff
	s.context = oidc.ClientContext(context.Background(), s.client)
	provider, err := oidc.NewProvider(s.context, s.config.IssuerURL)
	if err != nil {
		return fmt.Errorf("Failed to query provider %q: %v", s.config.IssuerURL, err)
	}
	var ss struct {
		// What scopes does a provider support?
		//
		// See: https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderMetadata
		ScopesSupported []string `json:"scopes_supported"`
	}
	if err := provider.Claims(&ss); err != nil {
		return fmt.Errorf("Failed to parse provider scopes_supported: %v", err)
	}

	if len(ss.ScopesSupported) == 0 {
		// scopes_supported is a "RECOMMENDED" discovery claim, not a required
		// one. If missing, assume that the provider follows the spec and has
		// an "offline_access" scope.
		s.config.OfflineAsScope = true
	} else {
		// See if scopes_supported has the "offline_access" scope.
		s.config.OfflineAsScope = func() bool {
			for _, scope := range ss.ScopesSupported {
				if scope == oidc.ScopeOfflineAccess {
					return true
				}
			}
			return false
		}()
	}

	s.provider = provider
	s.verifier = provider.Verifier(&oidc.Config{ClientID: s.config.ClientID})
	// Proxy setup

	// Run
	if s.config.TlsEnabled {
		fmt.Printf("listening on https://%s\n", s.config.Listen)
		if err := fmt.Errorf("%v", http.ListenAndServeTLS(s.config.Listen, s.config.TlsCert, s.config.TlsKey, s.router)); err != nil {
			return err
		}
	} else {
		fmt.Printf("listening on http://%s\n", s.config.Listen)
		if err := fmt.Errorf("%v", http.ListenAndServe(s.config.Listen, s.router)); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) RunProxy() error {
	du, err := url.Parse(s.config.DashboardUrl)
	if err != nil {
		return err
	}
	rp := httputil.NewSingleHostReverseProxy(du)
	if err := fmt.Errorf("%v", http.ListenAndServe(":8080", rp)); err != nil {
		return err
	}
	return nil
}

func main() {
	s := &Server{}
	if err := s.config.Init(os.Args); err != nil {
		log.Fatal(err)
	}
	if s.config.DashboardProxyEnabled {
		go func(){
			log.Fatal(s.RunProxy())
		}()
	}
	log.Fatal(s.Run())
}

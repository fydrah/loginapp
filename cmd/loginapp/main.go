/*
Copyright 2018 fydrah

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// Some code comes from @ericchiang (Dex - CoreOS)
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

type KubeUserInfo struct {
	IDToken      string
        RefreshToken string
        RedirectURL  string
        Claims       interface{}
        ClientSecret string
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
	renderTemplate(w, indexTmpl, s.config)
}

func (s *Server) HandlePostLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	switch r.FormValue("access") {
	case "CLI":
		http.Redirect(w, r, s.PrepareCallbackUrl("cli"), http.StatusSeeOther)
	}
}

func (s *Server) HandlePostCLI(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}


func (s *Server) HandleGetCallbackCLI(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	k, err := s.ProcessCallback(w, r, "cli")
	if err != nil {
		fmt.Errorf("error handling cli callback: %v", err)
		return
	}
	renderTemplate(w, tokenTmpl, k)
}

func (s *Server) ProcessCallback(w http.ResponseWriter, r *http.Request, c string) (KubeUserInfo, error) {
	var (
		err		error
		token		*oauth2.Token
		json_claims	map[string]interface{}
	)
	oauth2Config := s.OAuth2Config(nil, c)

	// Authorization redirect callback from OAuth2 auth flow.
	if errMsg := r.FormValue("error"); errMsg != "" {
		http.Error(w, errMsg+": "+r.FormValue("error_description"), http.StatusBadRequest)
		return KubeUserInfo{}, fmt.Errorf("%v: %q", errMsg, r.FormValue("error_description"))
	}
	code := r.FormValue("code")
	if code == "" {
		http.Error(w, fmt.Sprintf("no code in request: %q", r.Form), http.StatusBadRequest)
		return KubeUserInfo{}, fmt.Errorf("no code in request: %q", r.Form)
	}
	if state := r.FormValue("state"); state != s.config.AppName {
		http.Error(w, fmt.Sprintf("expected state %q got %q", s.config.AppName, state), http.StatusBadRequest)
		return KubeUserInfo{}, fmt.Errorf("expected state %q got %q", s.config.AppName, state)
	}
	token, err = oauth2Config.Exchange(s.context, code)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get token: %v", err), http.StatusInternalServerError)
		return KubeUserInfo{}, fmt.Errorf("failed to get token: %v", err)
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "no id_token in token response", http.StatusInternalServerError)
		return KubeUserInfo{}, fmt.Errorf("no id_token in token response")
	}
	idToken, err := s.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to verify ID token: %v", err), http.StatusInternalServerError)
		return KubeUserInfo{}, fmt.Errorf("Failed to verify ID token: %v", err)
	}
	var claims json.RawMessage
	if err := idToken.Claims(&claims); err != nil {
		return KubeUserInfo{}, fmt.Errorf("Failed to unmarshal claims from idToken: %v", err)
	}
	buff := new(bytes.Buffer)
	json.Indent(buff, []byte(claims), "", "  ")
	if err := json.Unmarshal(claims, &json_claims); err != nil {
		panic(err)
	}
	return KubeUserInfo {
		IDToken:	rawIDToken,
		RefreshToken:	token.RefreshToken,
		RedirectURL:	oauth2Config.RedirectURL,
		Claims:		json_claims,
		ClientSecret:	s.config.ClientSecret,
	}, nil
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

func main() {
	s := &Server{}
	if err := s.config.Init(os.Args); err != nil {
		log.Fatal(err)
	}
	log.Fatal(s.Run())
}

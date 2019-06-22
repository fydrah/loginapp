// Copyright 2018 fydrah
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Some code comes from @ericchiang (Dex - CoreOS)

package loginapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/coreos/go-oidc"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"html/template"
	"net/http"
)

// Server is the description
// of loginapp web server
type Server struct {
	client     *http.Client
	config     AppConfig
	provider   *oidc.Provider
	router     *httprouter.Router
	verifier   *oidc.IDTokenVerifier
	context    context.Context
	promrouter *httprouter.Router
}

// OAuth2Config generate oauth config
// based on scopes and yaml configuration file
func (s *Server) OAuth2Config(scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     s.config.OIDC.Client.ID,
		ClientSecret: s.config.OIDC.Client.Secret,
		RedirectURL:  s.config.OIDC.Client.RedirectURL,
		Endpoint:     s.provider.Endpoint(),
		Scopes:       scopes,
	}
}

// PrepareCallbackURL build and return
// an authCodeURL based on scopes provided
func (s *Server) PrepareCallbackURL() string {
	// Prepare scopes
	var (
		scopes               []string
		extraAuthCodeOptions []oauth2.AuthCodeOption
		authCodeURL          string
	)

	scopes = append(scopes, s.config.OIDC.ExtraScopes...)
	// Prepare cross client auth
	// see https://github.com/coreos/dex/blob/master/Documentation/custom-scopes-claims-clients.md
	for _, crossClient := range s.config.OIDC.CrossClients {
		scopes = append(scopes, "audience:server:client_id:"+crossClient)
	}

	scopes = append(scopes, "openid", "profile", "email", "groups")
	if *s.config.OIDC.OfflineAsScope {
		scopes = append(scopes, "offline_access")
	}
	for p, v := range s.config.OIDC.ExtraAuthCodeOpts {
		extraAuthCodeOptions = append(extraAuthCodeOptions, oauth2.SetAuthURLParam(p, v))
	}
	authCodeURL = s.OAuth2Config(scopes).AuthCodeURL(s.config.Name, extraAuthCodeOptions...)
	log.Debugf("auth code url: %s", authCodeURL)
	log.Debugf("request token with the following scopes: %v", scopes)
	return authCodeURL
}

// ProcessCallback check callback
// from our IdP after a successful user
// login.
func (s *Server) ProcessCallback(w http.ResponseWriter, r *http.Request) (KubeUserInfo, error) {
	var (
		err        error
		token      *oauth2.Token
		jsonClaims map[string]interface{}
	)
	oauth2Config := s.OAuth2Config(nil)

	// Authorization redirect callback from OAuth2 auth flow.
	if errMsg := r.FormValue("error"); errMsg != "" {
		msg := fmt.Sprintf("%v: %v", errMsg, r.FormValue("error_description"))
		http.Error(w, msg, http.StatusBadRequest)
		return KubeUserInfo{}, fmt.Errorf(msg)
	}
	code := r.FormValue("code")
	if code == "" {
		msg := fmt.Sprintf("no code in request: %q", r.Form)
		http.Error(w, msg, http.StatusBadRequest)
		return KubeUserInfo{}, fmt.Errorf(msg)
	}
	if state := r.FormValue("state"); state != s.config.Name {
		msg := fmt.Sprintf("expected state %q got %q", s.config.Name, state)
		http.Error(w, msg, http.StatusBadRequest)
		return KubeUserInfo{}, fmt.Errorf(msg)
	}
	token, err = oauth2Config.Exchange(s.context, code)
	if err != nil {
		msg := fmt.Sprintf("failed to get token: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return KubeUserInfo{}, fmt.Errorf(msg)
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		msg := "no id_token in token response"
		http.Error(w, msg, http.StatusInternalServerError)
		return KubeUserInfo{}, fmt.Errorf(msg)
	}
	idToken, err := s.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		msg := fmt.Sprintf("failed to verify ID token: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return KubeUserInfo{}, fmt.Errorf(msg)
	}
	var claims json.RawMessage
	if err := idToken.Claims(&claims); err != nil {
		return KubeUserInfo{}, fmt.Errorf("failed to unmarshal claims from idToken: %v", err)
	}
	buff := new(bytes.Buffer)
	if err := json.Indent(buff, []byte(claims), "", "  "); err != nil {
		return KubeUserInfo{}, fmt.Errorf("failed to format claims output: %v", err)
	}
	if err := json.Unmarshal(claims, &jsonClaims); err != nil {
		panic(err)
	}
	var usernameClaim interface{}
	if usernameClaim = jsonClaims[s.config.WebOutput.MainUsernameClaim]; usernameClaim == nil {
		msg := fmt.Sprintf("failed to find a claim matching the main_username_claim '%v'", s.config.WebOutput.MainUsernameClaim)
		http.Error(w, msg, http.StatusInternalServerError)
		return KubeUserInfo{}, fmt.Errorf(msg)
	}
	log.Debugf("token issued with claims: %v", jsonClaims)
	return KubeUserInfo{
		IDToken:       rawIDToken,
		RefreshToken:  token.RefreshToken,
		RedirectURL:   oauth2Config.RedirectURL,
		Claims:        jsonClaims,
		UsernameClaim: usernameClaim.(string),
		AppConfig:     s.config,
	}, nil
}

// RenderTemplate renders
// go-template formatted html page
func (s *Server) RenderTemplate(w http.ResponseWriter, tmpl *template.Template, data interface{}) {
	err := tmpl.Execute(w, data)
	if err == nil {
		return
	}
	log.Debugf("data: %v", data)
	switch err := err.(type) {
	case *template.Error:
		log.Errorf("error rendering template %s: %s", tmpl.Name(), err)

		http.Error(w, "internal server error", http.StatusInternalServerError)
	default:
		// An error with the underlying write, such as the connection being
		// dropped. Ignore for now.
	}
}

// Run launch app
func (s *Server) Run() error {
	var (
		provider   *oidc.Provider
		backoffErr error
	)
	// router setup
	s.router = httprouter.New()
	s.Routes()
	// client setup
	if s.client == nil {
		client, err := httpClientForRootCAs(s.config.OIDC.Issuer.RootCA)
		if err != nil {
			return err
		}
		s.client = client
	}
	// OIDC setup
	// Retry with backoff
	s.context = oidc.ClientContext(context.Background(), s.client)
	setupProvider := func() error {
		if provider, backoffErr = oidc.NewProvider(s.context, s.config.OIDC.Issuer.URL); backoffErr != nil {
			log.Errorf("failed to query provider %q: %v", s.config.OIDC.Issuer.URL, backoffErr)
			return backoffErr
		}
		return nil
	}
	if err := backoff.Retry(setupProvider, backoff.NewExponentialBackOff()); err != nil {
		return err
	}
	var ss struct {
		// What scopes does a provider support?
		//
		// See: https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderMetadata
		ScopesSupported []string `json:"scopes_supported"`
	}
	if err := provider.Claims(&ss); err != nil {
		return fmt.Errorf("failed to parse provider scopes_supported: %v", err)
	}

	if s.config.OIDC.OfflineAsScope == nil {
		if len(ss.ScopesSupported) > 0 {
			// See if scopes_supported has the "offline_access" scope.
			s.config.OIDC.OfflineAsScope = func() *bool {
				b := new(bool)
				for _, scope := range ss.ScopesSupported {
					if scope == oidc.ScopeOfflineAccess {
						*b = true
						return b
					}
				}
				*b = false
				return b
			}()
		}
	}

	s.provider = provider
	s.verifier = provider.Verifier(&oidc.Config{ClientID: s.config.OIDC.Client.ID})

	// Start prometheus metric exporter
	log.Infof("export metric on http://0.0.0.0:%v", s.config.Prometheus.Port)
	go PrometheusMetrics(s.config.Prometheus.Port)

	// Run
	if s.config.TLS.Enabled {
		log.Infof("listening on https://%s", s.config.Listen)
		if err := fmt.Errorf("%v", http.ListenAndServeTLS(s.config.Listen, s.config.TLS.Cert, s.config.TLS.Key, LoggingHandler(s.router))); err != nil {
			return err
		}
	} else {
		log.Infof("listening on http://%s", s.config.Listen)
		if err := fmt.Errorf("%v", http.ListenAndServe(s.config.Listen, LoggingHandler(s.router))); err != nil {
			return err
		}
	}
	return nil
}

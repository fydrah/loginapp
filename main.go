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

// Loginapp is an OIDC authentication web interface.
// It is mainly designed to render the token issued by an IdP (like Dex) in
// a kubernetes kubeconfig format.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/coreos/go-oidc"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	logger = logrus.New()
)

// Server is the description
// of loginapp web server
type Server struct {
	client   *http.Client
	config   AppConfig
	provider *oidc.Provider
	router   *httprouter.Router
	verifier *oidc.IDTokenVerifier
	context  context.Context
}

// KubeUserInfo contains all values
// needed by a user for OIDC authentication
type KubeUserInfo struct {
	ClientID      string
	IDToken       string
	RefreshToken  string
	RedirectURL   string
	Claims        interface{}
	ClientSecret  string
	UsernameClaim string
	Name          string
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
		scopes      []string
		authCodeURL string
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
		authCodeURL = s.OAuth2Config(scopes).AuthCodeURL(s.config.Name)
	} else {
		authCodeURL = s.OAuth2Config(scopes).AuthCodeURL(s.config.Name)
	}
	logger.Debugf("Request token with the following scopes: %v", scopes)
	return authCodeURL
}

// HandleGetIndex serves
// requests to index.html page
func (s *Server) HandleGetIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	renderTemplate(w, indexTmpl, s.config)
}

// HandleGetHealthz serves
// healthchecks requests (mainly
// used by kubernetes healthchecks)
// 200: OK, 500 otherwise
func (s *Server) HandleGetHealthz(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Check if provider is setup
	if s.provider == nil {
		logger.Debug("Provider is not yet setup or unavailable")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	// Check if our application can still contact the provider
	wellKnown := strings.TrimSuffix(s.config.OIDC.Issuer.URL, "/") + "/.well-known/openid-configuration"
	_, err := s.client.Head(wellKnown)
	if err != nil {
		logger.Debugf("Error while checking provider access: %v", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	// Should we add more checks ?
	w.WriteHeader(http.StatusOK)
}

// HandleLogin redirect to
// our IdP
func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.Redirect(w, r, s.PrepareCallbackURL(), http.StatusSeeOther)
}

// HandleGetCallback serves
// callback requests (from our IdP)
func (s *Server) HandleGetCallback(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	kc, err := s.ProcessCallback(w, r)
	if err != nil {
		logger.Errorf("error handling cli callback: %v", err)
		return
	}
	renderTemplate(w, tokenTmpl, kc)
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
		msg := fmt.Sprintf("Failed to verify ID token: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return KubeUserInfo{}, fmt.Errorf(msg)
	}
	var claims json.RawMessage
	if err := idToken.Claims(&claims); err != nil {
		return KubeUserInfo{}, fmt.Errorf("Failed to unmarshal claims from idToken: %v", err)
	}
	buff := new(bytes.Buffer)
	if err := json.Indent(buff, []byte(claims), "", "  "); err != nil {
		return KubeUserInfo{}, fmt.Errorf("Failed to format claims output: %v", err)
	}
	if err := json.Unmarshal(claims, &jsonClaims); err != nil {
		panic(err)
	}
	var usernameClaim interface{}
	if usernameClaim = jsonClaims[s.config.WebOutput.MainUsernameClaim]; usernameClaim == nil {
		msg := fmt.Sprintf("Failed to find a claim matching the main_username_claim '%v'", s.config.WebOutput.MainUsernameClaim)
		http.Error(w, msg, http.StatusInternalServerError)
		return KubeUserInfo{}, fmt.Errorf(msg)
	}
	logger.Debugf("Token issued with claims: %v", jsonClaims)
	return KubeUserInfo{
		IDToken:       rawIDToken,
		RefreshToken:  token.RefreshToken,
		RedirectURL:   oauth2Config.RedirectURL,
		Claims:        jsonClaims,
		ClientSecret:  s.config.OIDC.Client.Secret,
		ClientID:      s.config.WebOutput.MainClientID,
		UsernameClaim: usernameClaim.(string),
		Name:          s.config.Name,
	}, nil
}

// loggingHandler catch requests,
// add metadata and log user requests
func loggingHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		logger.WithFields(logrus.Fields{
			"method":           r.Method,
			"path":             r.URL.String(),
			"request_duration": t2.Sub(t1).String(),
			"protocol":         r.Proto,
			"remote_address":   r.RemoteAddr,
		}).Info()
	}
	return http.HandlerFunc(fn)
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
			logger.Errorf("Failed to query provider %q: %v", s.config.OIDC.Issuer.URL, backoffErr)
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
		return fmt.Errorf("Failed to parse provider scopes_supported: %v", err)
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

	// Run
	if s.config.TLS.Enabled {
		logger.Infof("listening on https://%s", s.config.Listen)
		if err := fmt.Errorf("%v", http.ListenAndServeTLS(s.config.Listen, s.config.TLS.Cert, s.config.TLS.Key, loggingHandler(s.router))); err != nil {
			return err
		}
	} else {
		logger.Infof("listening on http://%s", s.config.Listen)
		if err := fmt.Errorf("%v", http.ListenAndServe(s.config.Listen, loggingHandler(s.router))); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	app := NewCli()
	if err := app.Run(os.Args); err != nil {
		logger.Fatal(err)
	}
}

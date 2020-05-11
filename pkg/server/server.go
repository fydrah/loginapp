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

package server

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/fydrah/loginapp/pkg/client"
	"github.com/fydrah/loginapp/pkg/config"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

// Server is the description
// of loginapp web server
type Server struct {
	Config     *config.App
	client     *client.Client
	router     *httprouter.Router
	promrouter *httprouter.Router
}

// New initialize a new server
func New(cfg *config.App) *Server {
	s := new(Server)
	s.Config = cfg
	s.router = httprouter.New()
	s.Routes()
	return s
}

// ProcessCallback check callback
// from our IdP after a successful login
// and return user login information (token, claims, issuer)
func (s *Server) ProcessCallback(w http.ResponseWriter, r *http.Request) (KubeUserInfo, error) {
	// Authorization redirect callback from OAuth2 auth flow.
	if err := callbackFormCheck(w, r, s.Config.Secret); err != nil {
		return KubeUserInfo{}, err
	}
	token, rawIDToken, idToken, aErr := s.client.AuthCodeToIDToken(r.Context(), r.FormValue("code"))
	if aErr != nil {
		return KubeUserInfo{}, aErr
	}
	jsonClaims, cErr := client.ExtractClaims(idToken)
	if cErr != nil {
		return KubeUserInfo{}, cErr
	}
	// FORMAT: check if "usernameclaim" configured by user exist in response (should be done during init)
	var usernameClaim interface{}
	if usernameClaim = jsonClaims[s.Config.Web.MainUsernameClaim]; usernameClaim == nil {
		msg := fmt.Sprintf("failed to find a claim matching the main_username_claim '%v'", s.Config.Web.MainUsernameClaim)
		http.Error(w, msg, http.StatusInternalServerError)
		return KubeUserInfo{}, fmt.Errorf(msg)
	}
	log.Debugf("token issued with claims: %v", jsonClaims)
	return KubeUserInfo{
		IDToken:       rawIDToken,
		RefreshToken:  token.RefreshToken,
		RedirectURL:   s.Config.OIDC.Issuer.URL,
		Claims:        jsonClaims,
		UsernameClaim: usernameClaim.(string),
		AppConfig:     s.Config,
	}, nil
}

func callbackFormCheck(w http.ResponseWriter, r *http.Request, secret string) error {
	if errMsg := r.FormValue("error"); errMsg != "" {
		msg := fmt.Sprintf("%v: %v", errMsg, r.FormValue("error_description"))
		http.Error(w, msg, http.StatusBadRequest)
		return fmt.Errorf(msg)
	}
	code := r.FormValue("code")
	if code == "" {
		msg := fmt.Sprintf("no code in request: %q", r.Form)
		http.Error(w, msg, http.StatusBadRequest)
		return fmt.Errorf(msg)
	}
	if state, err := strconv.Unquote(r.FormValue("state")); client.VerifyState(r, state, secret) && err != nil {
		var msg string
		if err != nil {
			msg = fmt.Sprintf("unexpected quote error: %v", err)
		} else {
			msg = fmt.Sprintf("expected state %v got %q", client.GenerateState(r, secret), state)
		}
		http.Error(w, msg, http.StatusBadRequest)
		return fmt.Errorf(msg)
	}
	return nil
}

// RenderTemplate renders
// go-template formatted html page
func (s *Server) RenderTemplate(w http.ResponseWriter, tmpl *template.Template, data interface{}) {
	err := tmpl.Execute(w, data)
	if err == nil {
		return
	}
	switch err := err.(type) {
	case *template.Error:
		log.Errorf("error rendering template %s: %s", tmpl.Name(), err)

		http.Error(w, "internal server error", http.StatusInternalServerError)
	default:
		// An error with the underlying write, such as the connection being
		// dropped. Ignore for now.
		log.Errorf("error rendering template %s: %s", tmpl.Name(), err)
	}
}

// Run launch app
func (s *Server) Run() error {
	s.client = client.New(&s.Config.OIDC)
	if err := s.client.TLSSetup(); err != nil {
		return err
	}
	if err := s.client.ProviderSetup(); err != nil {
		return err
	}
	s.client.VerifierSetup()

	// Start prometheus metric exporter
	log.Infof("export metric on http://0.0.0.0:%v", s.Config.Metrics.Port)
	go PrometheusMetrics(s.Config.Metrics.Port)

	// Run
	if s.Config.TLS.Enabled {
		log.Infof("listening on https://%s", s.Config.Listen)
		if err := fmt.Errorf("%v", http.ListenAndServeTLS(s.Config.Listen, s.Config.TLS.Cert, s.Config.TLS.Key, LoggingHandler(s.router))); err != nil {
			return err
		}
	} else {
		log.Infof("listening on http://%s", s.Config.Listen)
		if err := fmt.Errorf("%v", http.ListenAndServe(s.Config.Listen, LoggingHandler(s.router))); err != nil {
			return err
		}
	}
	return nil
}

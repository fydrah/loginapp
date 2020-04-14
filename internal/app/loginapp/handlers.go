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
	"fmt"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"html/template"
	"net/http"
	"strings"
)

// HandleGetHealthz serves
// healthchecks requests (mainly
// used by kubernetes healthchecks)
// 200: OK, 500 otherwise
func (s *Server) HandleGetHealthz(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Check if provider is setup
	if s.provider == nil {
		log.Debug("provider is not yet setup or unavailable")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	// Check if our application can still contact the provider
	wellKnown := strings.TrimSuffix(s.config.OIDC.Issuer.URL, "/") + "/.well-known/openid-configuration"
	_, err := s.client.Head(wellKnown)
	if err != nil {
		log.Debugf("error while checking provider access: %v", err)
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
		log.Errorf("error handling cli callback: %v", err)
		return
	}
	var tokenTmpl = template.Must(template.ParseFiles(fmt.Sprintf("%v/token.html", s.config.WebOutput.TemplatesDir)))
	s.RenderTemplate(w, tokenTmpl, kc)
}

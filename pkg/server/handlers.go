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
	"os"
	"fmt"
	"io/ioutil"
	"html/template"
	"net/http"

	"github.com/gobuffalo/packr/v2"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

// HandleGetHealthz serves
// healthchecks requests (mainly
// used by kubernetes healthchecks)
// 200: OK, 500 otherwise
func (s *Server) HandleGetHealthz(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if !s.client.Healthz() {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	// Should we add more checks ?
	w.WriteHeader(http.StatusOK)
}

// HandleLogin redirects client to the IdP
func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.Redirect(w, r, s.client.AuthCodeURL(r, s.Config.Secret), http.StatusSeeOther)
}

// GetTemplateStr returns string representation of a template
func (s *Server) GetTemplateStr(templateName string) (string, error) {
	TmplFile, err := os.Stat(fmt.Sprintf("%v/%v.html", s.Config.Web.TemplatesDir, templateName))
	if (err != nil || !TmplFile.Mode().IsRegular()) {
		tBox := packr.New("templates", "../../web/templates")
		// Get the string representation of a file, or an error if it doesn't exist:
		tmpl, err := tBox.FindString(fmt.Sprintf("%v.html", templateName))
		if err != nil {
			log.Errorf("template loading failed: %v", err)
			return "", err
		}
		return tmpl, nil
	} else {
		// Read the string representation of a file, or an error if it can not be read:
		tmpl, err := ioutil.ReadFile(fmt.Sprintf("%v/%v.html", s.Config.Web.TemplatesDir, templateName))
		if err != nil {
			log.Errorf("template loading from file failed: %v", err)
			return "", err
		}
		return string(tmpl), nil
	}
}

// HandleGetCallback serves callback requests from the IdP
func (s *Server) HandleGetCallback(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	kc, err := s.ProcessCallback(w, r)
	if err != nil {
		log.Errorf("error handling callback: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	tokenTmplStr, err := s.GetTemplateStr("token")
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	var tokenTmpl = template.New("token")
	tokenTmpl.Parse(tokenTmplStr)
	s.RenderTemplate(w, tokenTmpl, kc)
}

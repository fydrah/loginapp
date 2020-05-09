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

package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Init load configuration,
// and run error/warning checks
func (a *App) Init() error {
	/*
		Extract data from yaml configuration file
	*/

	// dirty patch until https://github.com/spf13/viper/issues/608 is solved
	viper.Set("oidc.extra.authcodeopts", viper.GetStringMapString("oidc.extra.authcodeopts"))
	if err := viper.Unmarshal(&a); err != nil {
		return err
	}

	/*
		Error checks: list of checks which make loginapp failed
	*/
	errorChecks := []Check{
		{a.Name == "", "no name specified", nil},
		{a.Listen == "", "no listen 'ip:port' specified", nil},
		{a.OIDC.Client.ID == "", "no oidc.client.id specified", nil},
		{a.OIDC.Client.Secret == "", "no oidc.client.secret specified", nil},
		{a.OIDC.Client.RedirectURL == "", "no oidc.client.redirectURL specified", nil},
		{a.OIDC.Issuer.URL == "", "no oidc.issuer.url specified", nil},
		{a.OIDC.Issuer.RootCA == "", "no oidc.issuer.rootCA specified", nil},
		{a.TLS.Enabled && a.TLS.Cert == "", "no tls.cert specified", nil},
		{a.TLS.Enabled && a.TLS.Key == "", "no tls.key specified", nil},
	}

	if configCheck(errorChecks) {
		return fmt.Errorf("error while loading configuration")
	}

	/*
		Default checks: list of checks which makes loginapp setup default values

		Even if logger report this as an error log, this is not handle as an error.
		This issue could help to use loglevel as a parameter once merged:
		https://github.com/sirupsen/logrus/issues/646
	*/
	defaultChecks := []Check{
		{a.Secret == "", "no secret defined, using a random secret but it is strongly advised to add a secret since without it requests cannot be load balanced between multiple server", func() {
			a.Secret = randomString()
			log.Info(a.Secret)
		}},
		{a.Web.MainClientID == "", fmt.Sprintf("no output web.mainClientID specified, using default: %v", a.OIDC.Client.ID), func() {
			a.Web.MainClientID = a.OIDC.Client.ID
		}},
		{a.Web.MainUsernameClaim == "", "no output web.mainUsernameClaim specified, using default: 'name'", func() {
			a.Web.MainUsernameClaim = "name"
		}},
		{a.Metrics.Port == 0, "no metrics.port setup, using default: 9090", func() {
			a.Metrics.Port = 9090
		}},
	}
	if ok := configCheck(defaultChecks); !ok {
		log.Info("Non-blocking configuration missing, using defaults")
	}

	log.Debugf("Configuration loaded: %+v", a)

	return nil
}

func configCheck(checks []Check) bool {
	checkFailed := false
	for _, c := range checks {
		if !c.Check() {
			checkFailed = true
		}
	}
	return checkFailed
}

func randomString() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(b)
}

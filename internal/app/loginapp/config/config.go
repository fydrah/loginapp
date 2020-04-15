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
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// check run each each Check condition, if one
// check fails, return true
func configCheck(checks []Check) bool {
	checkFailed := false
	for _, c := range checks {
		if !c.Check() {
			checkFailed = true
		}
	}
	return checkFailed
}

// Init load configuration,
// setup logger and run
// error/warning checks
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
		Configuration checks
		(inspired from https://github.com/dexidp/dex/blob/master/cmd/dex/serve.go)
	*/
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}
	defaultAssetsDir := fmt.Sprintf("%v/web/assets", currentDir)
	defaultTemplatesDir := fmt.Sprintf("%v/web/templates", currentDir)

	/*
		Error checks: list of checks which make loginapp failed
	*/
	errorChecks := []Check{
		{a.Name == "", "no name specified", nil},
		{a.Listen == "", "no bind 'ip:port' specified", nil},
		{a.OIDC.Client.ID == "", "no client id specified", nil},
		{a.OIDC.Client.Secret == "", "no client secret specified", nil},
		{a.OIDC.Client.RedirectURL == "", "no redirect url specified", nil},
		{a.OIDC.Issuer.URL == "", "no issuer url specified", nil},
		{a.OIDC.Issuer.RootCA == "", "no issuer root_ca specified", nil},
		{a.TLS.Enabled && a.TLS.Cert == "", "no tls cert specified", nil},
		{a.TLS.Enabled && a.TLS.Key == "", "no tls key specified", nil},
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
		{a.Web.MainClientID == "", fmt.Sprintf("no output main_client_id specified, using default: %v", a.OIDC.Client.ID), func() {
			a.Web.MainClientID = a.OIDC.Client.ID
		}},
		{a.Web.AssetsDir == "", fmt.Sprintf("no assets_dir specified, using default: %v", defaultAssetsDir), func() {
			a.Web.AssetsDir = defaultAssetsDir
		}},
		{a.Web.TemplatesDir == "", fmt.Sprintf("no templates_dir specified, using default: %v", defaultTemplatesDir), func() {
			a.Web.TemplatesDir = defaultTemplatesDir
		}},
		{a.Web.MainUsernameClaim == "", "no output main_username_claim specified, using default: 'name'", func() {
			a.Web.MainUsernameClaim = "name"
		}},
		{a.Metrics.Port == 0, "no prometheus scrap port setup, using default: 9090", func() {
			a.Metrics.Port = 9090
		}},
	}
	_ = configCheck(defaultChecks)

	log.Debugf("Configuration loaded: %+v", a)

	return nil
}

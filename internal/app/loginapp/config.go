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

package loginapp

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"strings"
)

// AppConfig contains all configuration options. Options
// are overrided by environement variables with
// the prefix LOGINAPP_ followed by the path of the option.
// Ex: LOGINAPP_OIDC_CLIENT_ID=customid
type AppConfig struct {
	Name   string `yaml:"name"`
	Listen string `yaml:"listen"`
	OIDC   struct {
		Client struct {
			ID          string `yaml:"id"`
			Secret      string `yaml:"secret"`
			RedirectURL string `yaml:"redirectURL"`
		} `yaml:"client"`
		Issuer struct {
			URL    string `yaml:"url"`
			RootCA string `yaml:"rootCA"`
		} `yaml:"issuer"`
		ExtraScopes       []string          `yaml:"extraScopes"`
		ExtraAuthCodeOpts map[string]string `yaml:"extraAuthCodeOpts"`
		OfflineAsScope    *bool             `yaml:"offlineAsScope"`
		CrossClients      []string          `yaml:"crossClients"`
	} `yaml:"oidc"`
	TLS struct {
		Enabled bool   `yaml:"enabled"`
		Cert    string `yaml:"cert"`
		Key     string `yaml:"key"`
	} `yaml:"tls"`
	Log struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"log"`
	WebOutput struct {
		MainUsernameClaim string `yaml:"mainUsernameClaim"`
		MainClientID      string `yaml:"mainClientID"`
		AssetsDir         string `yaml:"assetsDir"`
		TemplatesDir      string `yaml:"templatesDir"`
	} `yaml:"webOutput"`
	Prometheus struct {
		Port int `yaml:"port"`
	} `yaml:"prometheus"`
	Clusters []Cluster `yaml:"clusters"`
}

// appCheck struct
// used by check function
type appCheck struct {
	FailedCondition bool
	Message         string
	DefaultAction   func()
}

// check checks each appCheck, if one
// check fails, return true
func check(checks []appCheck) bool {
	checkFailed := false
	for _, c := range checks {
		if c.FailedCondition {
			if c.DefaultAction != nil {
				c.DefaultAction()
				log.Info(c.Message)
			} else {
				log.Error(c.Message)
			}
			checkFailed = true
		}
	}
	return checkFailed
}

// configLogger setup application logger
// Default loglevel is info
func configLogger(format string, logLevel string) {
	switch f := format; f {
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
	case "text":
		log.SetFormatter(&log.TextFormatter{})
	default:
		log.SetFormatter(&log.JSONFormatter{})
		log.Warningf("format %q not available, use json|text. Using json format", f)
		format = "json"
	}
	log.Debugf("Using %s log format", format)
	switch l := logLevel; l {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warning":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
		log.Warningf("log level %q not available, use debug|info|warning|error. Using Info log level", l)
		logLevel = "info"
	}
	log.Debugf("Using %s log level", logLevel)
}

// Init load configuration,
// setup logger and run
// error/warning checks
func (a *AppConfig) Init() error {
	/*
		Extract data from yaml configuration file
	*/
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	if err := viper.Unmarshal(&a); err != nil {
		return err
	}

	/*
		Configure log level
	*/
	configLogger(strings.ToLower(a.Log.Format), strings.ToLower(a.Log.Level))

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
	errorChecks := []appCheck{
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
	if check(errorChecks) {
		return fmt.Errorf("error while loading configuration")
	}
	/*
		Default checks: list of checks which makes loginapp setup default values

		Even if logger report this as an error log, this is not handle as an error.
		This issue could help to use loglevel as a parameter once merged:
		https://github.com/sirupsen/logrus/issues/646
	*/
	defaultChecks := []appCheck{
		{a.WebOutput.MainClientID == "", fmt.Sprintf("no output main_client_id specified, using default: %v", a.OIDC.Client.ID), func() {
			a.WebOutput.MainClientID = a.OIDC.Client.ID
		}},
		{a.WebOutput.AssetsDir == "", fmt.Sprintf("no assets_dir specified, using default: %v", defaultAssetsDir), func() {
			a.WebOutput.AssetsDir = defaultAssetsDir
		}},
		{a.WebOutput.TemplatesDir == "", fmt.Sprintf("no templates_dir specified, using default: %v", defaultTemplatesDir), func() {
			a.WebOutput.TemplatesDir = defaultTemplatesDir
		}},
		{a.WebOutput.MainUsernameClaim == "", "no output main_username_claim specified, using default: 'name'", func() {
			a.WebOutput.MainUsernameClaim = "name"
		}},
		{a.Prometheus.Port == 0, "no prometheus scrap port setup, using default: 9090", func() {
			a.Prometheus.Port = 9090
		}},
	}
	_ = check(defaultChecks)

	log.Debugf("Configuration loaded: %+v", a)
	return nil
}

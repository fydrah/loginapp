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

package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
)

type AppConfig struct {
	Name   string `yaml:"name"`
	Listen string `yaml:"listen"`
	OIDC   struct {
		Client struct {
			ID          string `yaml:"id"`
			Secret      string `yaml:"secret"`
			RedirectURL string `yaml:"redirect_url"`
		} `yaml:"client"`
		Issuer struct {
			URL    string `yaml:"url"`
			RootCA string `yaml:"root_ca"`
		} `yaml:"issuer"`
		ExtraScopes       []string          `yaml:"extra_scopes"`
		ExtraAuthCodeOpts map[string]string `yaml:"extra_auth_code_opts"`
		OfflineAsScope    *bool             `yaml:"offline_as_scope"`
		CrossClients      []string          `yaml:"cross_clients"`
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
		MainUsernameClaim string `yaml:"main_username_claim"`
		MainClientID      string `yaml:"main_client_id"`
		AssetsDir         string `yaml:"assets_dir"`
	} `yaml:"web_output"`
	Prometheus struct {
		Port int `yaml:"port"`
	} `yaml:"prometheus"`
	Clusters []Cluster `yaml:"clusters"`
}

// appCheck struct
// used by check function
type appCheck struct {
	Condition     bool
	Message       string
	DefaultAction func()
}

// check checks each appCheck, if one
// check fails, return true
func check(checks []appCheck) bool {
	checkFailed := false
	for _, c := range checks {
		if c.Condition {
			logger.Error(c.Message)
			checkFailed = true
			if c.DefaultAction != nil {
				c.DefaultAction()
			}
		}
	}
	return checkFailed
}

// configLogger setup application logger
// Default loglevel is info
func configLogger(format string, logLevel string) {
	switch f := format; f {
	case "json":
		logger.Formatter = &logrus.JSONFormatter{}
	case "text":
		logger.Formatter = &logrus.TextFormatter{}
	default:
		logger.Formatter = &logrus.JSONFormatter{}
		logger.Warningf("format %q not available, use json|text. Using json format", f)
		format = "json"
	}
	logger.Debugf("Using %s log format", format)
	switch l := logLevel; l {
	case "debug":
		logger.Level = logrus.DebugLevel
	case "info":
		logger.Level = logrus.InfoLevel
	case "warning":
		logger.Level = logrus.WarnLevel
	case "error":
		logger.Level = logrus.ErrorLevel
	default:
		logger.Level = logrus.InfoLevel
		logger.Warningf("log level %q not available, use debug|info|warning|error. Using Info log level", l)
		logLevel = "info"
	}
	logger.Debugf("Using %s log level", logLevel)
}

// Init load configuration,
// setup logger and run
// error/warning checks
func (a *AppConfig) Init(config string) error {
	/*
		Extract data from yaml configuration file
	*/
	logger.Debugf("loading configuration file: %v", config)
	configData, err := ioutil.ReadFile(config)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %v", config, err)
	}
	logger.Debugf("unmarshal data: %v", configData)
	if err := yaml.Unmarshal(configData, &a); err != nil {
		return fmt.Errorf("error parse config file %s: %v", config, err)
	}

	/*
		Configure log level
	*/
	configLogger(strings.ToLower(a.Log.Format), strings.ToLower(a.Log.Level))

	/*
		Configuration checks
		(inspired from https://github.com/coreos/dex/blob/master/cmd/dex/serve.go)
	*/
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}
	defaultAssetsDir := fmt.Sprintf("%v/assets", currentDir)
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
		{a.WebOutput.MainUsernameClaim == "", "no output main_username_claim specified, using default: 'name'", func() {
			a.WebOutput.MainUsernameClaim = "name"
		}},
		{a.Prometheus.Port == 0, "no prometheus scrap port setup, using default: 9090", func() {
			a.Prometheus.Port = 9090
		}},
	}
	_ = check(defaultChecks)

	logger.Debugf("Configuration loaded: %+v", a)
	return nil
}

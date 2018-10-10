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
			Id          string `yaml:"id"`
			Secret      string `yaml:"secret"`
			RedirectURL string `yaml:"redirect_url"`
		} `yaml:"client"`
		Issuer struct {
			Url    string `yaml:"url"`
			RootCA string `yaml:"root_ca"`
		} `yaml:"issuer"`
		ExtraScopes    []string `yaml:"extra_scopes"`
		OfflineAsScope *bool    `yaml:"offline_as_scope"`
		CrossClients   []string `yaml:"cross_clients"`
	} `yaml:"oidc"`
	Tls struct {
		Enabled bool   `yaml:"enabled"`
		Cert    string `yaml:"cert"`
		Key     string `yaml:"key"`
	} `yaml:"tls"`
	Log struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"log"`
	WebOutput struct {
		MainClientID string `yaml:"main_client_id"`
		AssetsDir    string `yaml:"assets_dir"`
		SkipMainPage bool   `yaml:"skip_main_page"`
	} `yaml:"web_output"`
}

func ConfigLogger(format string, logLevel string) {
	switch f := format; f {
	case "json":
		logger.Formatter = &logrus.JSONFormatter{}
	case "text":
		logger.Formatter = &logrus.TextFormatter{}
	default:
		logger.Formatter = &logrus.JSONFormatter{}
		logger.Warningf("Format %q not available, use json|text. Using json format", f)
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
		logger.Warningf("Log level %q not available, use debug|info|warning|error. Using Info log level", l)
		logLevel = "info"
	}
	logger.Debugf("Using %s log level", logLevel)
}

func (a *AppConfig) Init(config string) error {
	/*
		Extract data from yaml configuration file
	*/
	logger.Debugf("Loading configuration file: %v", config)
	configData, err := ioutil.ReadFile(config)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %v", config, err)
	}
	logger.Debugf("Unmarshal data: %v", configData)
	if err := yaml.Unmarshal(configData, &a); err != nil {
		return fmt.Errorf("error parse config file %s: %v", config, err)
	}

	/*
		Configure log level
	*/
	ConfigLogger(strings.ToLower(a.Log.Format), strings.ToLower(a.Log.Level))

	/*
		Configuration checks
		(inspired from https://github.com/coreos/dex/blob/master/cmd/dex/serve.go)
	*/
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}
	if a.WebOutput.AssetsDir == "" {
		a.WebOutput.AssetsDir = fmt.Sprintf("%v/assets", currentDir)
		logger.Infof("no assets dir specified, using default: %v", a.WebOutput.AssetsDir)
	}
	if a.WebOutput.MainClientID == "" {
		a.WebOutput.MainClientID = a.OIDC.Client.Id
		logger.Infof("no output main_client_id specified, using default: %v", a.WebOutput.MainClientID)
	}
	configChecks := []struct {
		failed bool
		msg    string
	}{
		{a.Name == "", "no name specified"},
		{a.Listen == "", "no bind 'ip:port' specified"},
		{a.OIDC.Client.Id == "", "no client id specified"},
		{a.OIDC.Client.Secret == "", "no client secret specified"},
		{a.OIDC.Client.RedirectURL == "", "no redirect url specified"},
		{a.OIDC.Issuer.Url == "", "no issuer url specified"},
		{a.OIDC.Issuer.RootCA == "", "no issuer root_ca specified"},
		{a.Tls.Enabled && a.Tls.Cert == "", "no tls cert specified"},
		{a.Tls.Enabled && a.Tls.Key == "", "no tls key specified"},
	}
	checksFailed := func() bool {
		checkFailed := false
		for _, c := range configChecks {
			if c.failed {
				logger.Errorf("Check failed: %v", c.msg)
				checkFailed = true
			}
		}
		return checkFailed
	}()
	if checksFailed {
		return fmt.Errorf("Error while loading configuration")
	}

	logger.Debugf("Configuration loaded: %+v", a)
	return nil
}

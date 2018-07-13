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
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
	"github.com/sirupsen/logrus"
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
		OfflineAsScope bool     `yaml:"offline_as_scope"`
		CrossClients   []string `yaml:"cross_clients"`
	} `yaml:"oidc"`
	Tls struct {
		Enabled	bool	`yaml:"enabled"`
		Cert	string	`yaml:"cert"`
		Key	string	`yaml:"key"`
	} `yaml:"tls"`
	Log struct {
		Level	string	`yaml:"level"`
		Format	string	`yaml:"format"`
	} `yaml:"log"`
}

func (a *AppConfig) Init(config string) error {
	configData, err := ioutil.ReadFile(config)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %v", config, err)
	}
	if err := yaml.Unmarshal(configData, &a); err != nil {
		return fmt.Errorf("error parse config file %s: %v", config, err)
	}
	switch f := strings.ToLower(a.Log.Format); f {
		case "json":
			logger.Formatter = &logrus.JSONFormatter{}
		case "text":
			logger.Formatter = &logrus.TextFormatter{}
		default:
			logger.Formatter = &logrus.JSONFormatter{}
			logger.Warningf("Format %q not available, use json|text. Using json format", f)
	}
	logger.Debugf("Using %s log format", a.Log.Format)
	switch f := strings.ToLower(a.Log.Level); f {
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
			logger.Warningf("Log level %q not available, use debug|info|warning|error. Using Info log level", f)
	}
	logger.Debugf("Using %s log level", a.Log.Format)
	logger.Debugf("Configuration loaded: %+v", a)
	return nil
}

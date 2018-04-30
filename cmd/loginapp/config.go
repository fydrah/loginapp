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
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"fmt"
)

type AppConfig struct {
	AppName			string	`yaml:"app_name"`
	ClientID		string	`yaml:"client_id"`
	ClientSecret		string	`yaml:"client_secret"`
	CrossClients		string	`yaml:"cross_clients"`
	Debug			bool	`yaml:"debug"`
	DisableChoices		bool	`yaml:"disable_choices"`
	ExtraScopes		string	`yaml:"extra_scopes"`
	IssuerRootCA		string	`yaml:"issuer_root_ca"`
	IssuerURL		string	`yaml:"issuer_url"`
	Listen			string	`yaml:"listen"`
	OfflineAsScope		bool	`yaml:"offline_as_scope"`
	RedirectURL		string	`yaml:"redirect_url"`
	TlsCert			string	`yaml:"tls_cert"`
	TlsEnabled		bool	`yaml:"tls_enabled"`
	TlsKey			string	`yaml:"tls_key"`
}

func (a *AppConfig) Init(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("Configuration file missing")
	}
	configFile := args[1]
	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %v", configFile, err)
	}
	if err := yaml.Unmarshal(configData, &a); err != nil {
		return fmt.Errorf("error parse config file %s: %v", configFile, err)
	}
	return nil
}

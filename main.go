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

// Loginapp is an OIDC authentication web interface.
// It is mainly designed to render the token issued by an IdP (like Dex) in
// a kubernetes kubeconfig format.
package main

import (
	"github.com/sirupsen/logrus"
	"os"
)

var (
	logger = logrus.New()
)

// KubeUserInfo contains all values
// needed by a user for OIDC authentication
type KubeUserInfo struct {
	ClientID          string
	IDToken           string
	RefreshToken      string
	RedirectURL       string
	ExtraAuthCodeOpts interface{}
	Claims            interface{}
	ClientSecret      string
	UsernameClaim     string
	Name              string
}

func main() {
	app := NewCli()
	if err := app.Run(os.Args); err != nil {
		logger.Fatal(err)
	}
}

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
	"github.com/spf13/cobra"
)

// OIDC is the OpenID configuration
type OIDC struct {
	Client         OIDCClient
	Issuer         OIDCIssuer
	Extra          OIDCExtra
	OfflineAsScope bool
	CrossClients   []string
}

// OIDCClient is the client OpenID configuration
type OIDCClient struct {
	ID          string
	Secret      string
	RedirectURL string
}

// OIDCIssuer is the issuer OpenID configuration
type OIDCIssuer struct {
	URL    string
	RootCA string
}

// OIDCIssuer is the extra OpenID configuration supported
type OIDCExtra struct {
	Scopes       []string
	AuthCodeOpts map[string]string
}

func (o *OIDC) AddFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("oidc-offlineasscope", false, "Issue a refresh token for offline access")
	cmd.Flags().StringSlice("oidc-crossclients", nil, "Issue token on behalf of this list of client IDs")
	o.Client.AddFlags(cmd)
	o.Issuer.AddFlags(cmd)
	o.Extra.AddFlags(cmd)
}

func (oc *OIDCClient) AddFlags(cmd *cobra.Command) {
	cmd.Flags().String("oidc-client-id", "loginapp", "Client ID")
	cmd.Flags().String("oidc-client-secret", "", "Client secret")
	cmd.Flags().String("oidc-client-redirecturl", "", "Redirect URL for callback. This must be the same than the one provided to the IDP. Must end with '/callback'")
}

func (oi *OIDCIssuer) AddFlags(cmd *cobra.Command) {
	cmd.Flags().String("oidc-issuer-url", "", "Full URL of issuer before '/.well-known/openid-configuration' path")
	cmd.Flags().String("oidc-issuer-rootca", "", "Certificate authority of the issuer")
}

func (oe *OIDCExtra) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringSlice("oidc-extra-scopes", nil, "List of extra scopes to ask")
	cmd.Flags().StringToString("oidc-extra-authcodeopts", nil, "K/V list of extra authorisation code to include in token request")
}

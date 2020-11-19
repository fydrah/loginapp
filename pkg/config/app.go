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

// App is the loginapp configuration set
type App struct {
	Name     string
	Listen   string
	Secret   string
	OIDC     OIDC
	TLS      TLS
	Web      Web
	Metrics  Metrics
	Clusters []Cluster
}

// AddFlags init common App flags
func (a *App) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("name", "n", "Loginapp", "Application name. Used for web title.")
	cmd.Flags().StringP("listen", "l", "0.0.0.0:8080", "Listen interface and port")
	cmd.Flags().StringP("secret", "s", "", "Application secret. Must be identical across all loginapp server replicas (this is not the OIDC Client secret)")
	a.OIDC.AddFlags(cmd)
	a.TLS.AddFlags(cmd)
	a.Web.AddFlags(cmd)
	a.Metrics.AddFlags(cmd)
}

// OIDC is the OpenID configuration
type OIDC struct {
	Client         OIDCClient
	Issuer         OIDCIssuer
	Extra          OIDCExtra
	OfflineAsScope bool
	CrossClients   []string
	Scopes         []string
}

// AddFlags init oidc flags
func (o *OIDC) AddFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("oidc-offlineasscope", false, "Issue a refresh token for offline access")
	cmd.Flags().StringSlice("oidc-crossclients", nil, "Issue token on behalf of this list of client IDs")
	cmd.Flags().StringSlice("oidc-scopes", []string{"openid", "profile", "email", "groups"}, "List of scopes to request. Updating this parameter will override existing scopes.")
	o.Client.AddFlags(cmd)
	o.Issuer.AddFlags(cmd)
	o.Extra.AddFlags(cmd)
}

// OIDCClient is the client OpenID configuration
type OIDCClient struct {
	ID          string
	Secret      string
	RedirectURL string
}

// AddFlags init oidc client flags
func (oc *OIDCClient) AddFlags(cmd *cobra.Command) {
	cmd.Flags().String("oidc-client-id", "loginapp", "Client ID")
	cmd.Flags().String("oidc-client-secret", "", "Client secret")
	cmd.Flags().String("oidc-client-redirecturl", "", "Redirect URL for callback. This must be the same than the one provided to the IDP. Must end with '/callback'")
}

// OIDCIssuer is the issuer OpenID configuration
type OIDCIssuer struct {
	URL                string
	RootCA             string
	InsecureSkipVerify bool
}

// AddFlags init oidc issuer flags
func (oi *OIDCIssuer) AddFlags(cmd *cobra.Command) {
	cmd.Flags().String("oidc-issuer-url", "", "Full URL of issuer before '/.well-known/openid-configuration' path")
	cmd.Flags().String("oidc-issuer-rootca", "", "Certificate authority of the issuer")
	cmd.Flags().Bool("oidc-issuer-insecureskipverify", false, "Skip issuer certificate validation (usefull for testing). It is not advised to use this option in production")
}

// OIDCIssuer is the extra OpenID configuration supported
type OIDCExtra struct {
	Scopes       []string
	AuthCodeOpts map[string]string
}

// AddFlags init oidc extra flags
func (oe *OIDCExtra) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringSlice("oidc-extra-scopes", nil, "[DEPRECATED] List of extra scopes to ask. Use oidc.scopes option instead. Option will be removed in next release.")
	cmd.Flags().StringToString("oidc-extra-authcodeopts", nil, "K/V list of extra authorisation code to include in token request")
}

// Metrics is the exported metrics configuration
type Metrics struct {
	Port int
}

// AddFlags init metrics flags
func (m *Metrics) AddFlags(cmd *cobra.Command) {
	cmd.Flags().Int("metrics-port", 9090, "Port to export metrics")
}

// TLS is the tls configuration, required to configure HTTPS endpoint for Loginapp
type TLS struct {
	Enabled bool
	Cert    string
	Key     string
}

// AddFlags init tls flags
func (t *TLS) AddFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("tls-enabled", false, "Enable TLS")
	cmd.Flags().String("tls-cert", "", "TLS certificate path")
	cmd.Flags().String("tls-key", "", "TLS private key path")
}

// Web is the web output configuration, mainly used to customize output
type Web struct {
	MainUsernameClaim string
	MainClientID      string
	TemplatesDir      string
	AssetsDir         string
	Kubeconfig        WebKubeconfig
}

// AddFlags init web flags
func (w *Web) AddFlags(cmd *cobra.Command) {
	cmd.Flags().String("web-mainusernameclaim", "email", "Claim to use for username (depends on IDP available claims")
	cmd.Flags().String("web-mainclientid", "", "Application client ID")
	cmd.Flags().String("web-templatesdir", "/web/templates", "Directory to look for templates, which are overriding embedded")
	cmd.Flags().String("web-assetsdir", "/web/assets", "Directory to look for assets, which are overriding embedded")
	w.Kubeconfig.AddFlags(cmd)
}

// WebKubeconfig manages default web output for kubeconfig
type WebKubeconfig struct {
	DefaultCluster   string
	DefaultNamespace string
	DefaultContext   string
}

// AddFlags init web kubeconfig flags
func (wk *WebKubeconfig) AddFlags(cmd *cobra.Command) {
	cmd.Flags().String("web-kubeconfig-defaultcluster", "", "Default cluster name to use for full kubeconfig output")
	cmd.Flags().String("web-kubeconfig-defaultnamespace", "default", "Default namespace to use for full kubeconfig output")
	cmd.Flags().String("web-kubeconfig-defaultcontext", "", "Default context to use for full kubeconfig output. Use the following format by default: 'defaultcluster'/'usernameclaim'")
}

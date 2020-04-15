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
	OIDC     OIDC
	TLS      TLS
	Web      Web
	Metrics  Metrics
	Clusters []Cluster
}

func (a *App) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("name", "n", "Loginapp", "Application name. Used for web title.")
	cmd.Flags().StringP("listen", "l", "0.0.0.0:8080", "Listen interface and port")
	a.OIDC.AddFlags(cmd)
	a.TLS.AddFlags(cmd)
	a.Web.AddFlags(cmd)
	a.Metrics.AddFlags(cmd)
}

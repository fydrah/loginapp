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

// Web is the web output configuration, mainly used to customize output
type Web struct {
	MainUsernameClaim string
	MainClientID      string
	AssetsDir         string
	TemplatesDir      string
}

func (w *Web) AddFlags(cmd *cobra.Command) {
	cmd.Flags().String("web-mainusernameclaim", "email", "Claim to use for username (depends on IDP available claims")
	cmd.Flags().String("web-mailclientid", "loginapp", "Application client ID")
	cmd.Flags().String("web-assertsdir", "", "Custom asserts directory")
	cmd.Flags().String("web-templatedir", "", "Custom templates directory")
}

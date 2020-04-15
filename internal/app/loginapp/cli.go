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

package loginapp

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var (
	// GitVersion returns latest tag
	GitVersion = "X.X.X"
	// GitHash return hash of latest commit
	GitHash = "XXXXXXX"

	loginappCmd = &cobra.Command{
		Use:     "loginapp",
		Short:   "Web application for Kubernetes CLI configuration with OIDC",
		Version: fmt.Sprintf("%v build %v\n", GitVersion, GitHash),
	}

	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Run loginapp application",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := &Server{}
			if err := s.config.Init(); err != nil {
				cmd.SilenceUsage = true
				return err
			}
			if err := s.Run(); err != nil {
				cmd.SilenceUsage = true
				return err
			}
			return nil
		},
	}

	configFile string
)

func init() {
	// Configure flags
	serveCmd.Flags().StringVar(&configFile, "config", "", "Configuration file")
	serveCmd.MarkFlagRequired("config")

	// Configure Sub-commands
	loginappCmd.AddCommand(serveCmd)

	// Configure init, for us read configuration file and env vars
	cobra.OnInitialize(func() {
		viper.SetConfigType("yaml")
		viper.SetConfigFile(configFile)
		viper.SetEnvPrefix("loginapp")
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		viper.AutomaticEnv()
	})
}

func Execute() {
	if err := loginappCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

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
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/fydrah/loginapp/internal/app/loginapp/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	// GitVersion returns latest tag
	GitVersion = "X.X.X"
	// GitHash return hash of latest commit
	GitHash = "XXXXXXX"

	appConfig = &config.App{}

	loginappCmd = &cobra.Command{
		Use:     "loginapp",
		Short:   "Web application for Kubernetes CLI configuration with OIDC",
		Version: fmt.Sprintf("%v build %v\n", GitVersion, GitHash),
	}

	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Run loginapp application",
		Long: `
Perform configuration checks and run Loginapp.

Loginapp supports three configuration formats:
* Configuration file: '--config' flag
* Flags: '--oidc-xxx' flags for example
* Environment vars: each flag provides an environment var with
  'LOGINAPP_' prefix.
  Ex: '--oidc-client-secret' --> 'LOGINAPP_OIDC_CLIENT_SECRET'

Configuration precedence: flags > environment vars > configuration file`,
		Run: func(cmd *cobra.Command, args []string) {
			s := NewServer(appConfig)
			if err := s.config.Init(); err != nil {
				log.Fatal(err)
			}
			// ConfigChange is trigger more than once sometimes,
			// check issue https://github.com/spf13/viper/issues/609
			viper.OnConfigChange(func(e fsnotify.Event) {
				log.Info("Configuration changed, reloading...")
				if err := s.config.Init(); err != nil {
					cmd.SilenceUsage = true
					log.Errorf("Configuration init failed: %v", err)
					log.Info("Still using previous configuration")
				}
			})
			if err := s.Run(); err != nil {
				cmd.SilenceUsage = true
				log.Fatal(err)
			}
		},
	}

	configFile string
	verbose    bool
)

func init() {
	// Create flags
	appConfig.AddFlags(serveCmd)
	serveCmd.Flags().VisitAll(func(f *pflag.Flag) {
		if err := viper.BindPFlag(strings.Replace(f.Name, "-", ".", -1), f); err != nil {
			log.Fatal(err)
		}
	})
	// Configure flags
	serveCmd.Flags().StringVarP(&configFile, "config", "c", "", "Configuration file")
	// serveCmd.MarkFlagRequired("config")
	// Configure Sub-commands
	loginappCmd.AddCommand(serveCmd)
	loginappCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	// Configuration init, read configuration file and env vars
	cobra.OnInitialize(func() {
		log.SetFormatter(&log.JSONFormatter{})
		log.SetLevel(log.InfoLevel)
		if verbose {
			log.SetLevel(log.DebugLevel)
		}
		viper.SetConfigType("yaml")
		if configFile != "" {
			viper.SetConfigFile(configFile)
			if err := viper.ReadInConfig(); err != nil {
				log.Errorf("error while reading configuration file '%s': %v", configFile, err)
			}
			viper.WatchConfig()
		}
		viper.SetEnvPrefix("loginapp")
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
		viper.AutomaticEnv()
	})
}

func Execute() {
	if err := loginappCmd.Execute(); err != nil {
		log.Exit(1)
	}
}

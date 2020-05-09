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

package cmd

import (
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/fydrah/loginapp/pkg/config"
	"github.com/fydrah/loginapp/pkg/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	ServeCmd = &cobra.Command{
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
			s := server.New(serveCfg)
			if err := s.Config.Init(); err != nil {
				log.Fatal(err)
			}
			if err := s.Run(); err != nil {
				cmd.SilenceUsage = true
				log.Fatal(err)
			}
			// ConfigChange is trigger more than once sometimes,
			// check issue https://github.com/spf13/viper/issues/609
			viper.OnConfigChange(func(e fsnotify.Event) {
				log.Info("Configuration changed, reloading...")
				if err := s.Config.Init(); err != nil {
					cmd.SilenceUsage = true
					log.Errorf("Configuration init failed: %v", err)
					log.Info("Still using previous configuration")
				}
			})
		},
	}
	serveCfg     *config.App
	serveCfgFile string
)

func init() {
	serveCfg = new(config.App)
	flagsSetup()
	// Configure flags
	cobra.OnInitialize(func() {
		configSetup()
		envSetup()
	})
}

func flagsSetup() {
	// App flags
	serveCfg.AddFlags(ServeCmd)
	ServeCmd.Flags().VisitAll(func(f *pflag.Flag) {
		// replaces "-" by ".", ex: "oidc-client-secret" is stored as "oidc.client.secret"
		// this is required for overrides by env vars and config
		if err := viper.BindPFlag(strings.Replace(f.Name, "-", ".", -1), f); err != nil {
			log.Fatal(err)
		}
	})
	// Static flags
	ServeCmd.Flags().StringVarP(&serveCfgFile, "config", "c", "", "Configuration file")
}

func configSetup() {
	viper.SetConfigType("yaml")
	if serveCfgFile != "" {
		viper.SetConfigFile(serveCfgFile)
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("error while reading configuration file '%s': %v", serveCfgFile, err)
		}
		viper.WatchConfig()
	}
}

func envSetup() {
	viper.SetEnvPrefix("LOGINAPP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()
}

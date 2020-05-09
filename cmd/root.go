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
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// GitVersion returns latest tag
	GitVersion = "X.X.X"
	// GitHash return hash of latest commit
	GitHash = "XXXXXXX"

	rootCmd = &cobra.Command{
		Use:     "loginapp",
		Short:   "Web application for Kubernetes CLI configuration with OIDC",
		Version: fmt.Sprintf("%v build %v\n", GitVersion, GitHash),
	}
	verbose = false
)

func init() {
	rootCmd.AddCommand(ServeCmd)
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
	cobra.OnInitialize(func() {
		if verbose {
			log.SetLevel(log.DebugLevel)
		}
	})
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		log.Exit(1)
	}
}

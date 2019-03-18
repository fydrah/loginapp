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

package main

import (
	"fmt"

	"github.com/prometheus/common/version"
	"github.com/urfave/cli"
)

// NewCli configure loginapp CLI
func NewCli() *cli.App {
	app := cli.NewApp()
	cli.AppHelpTemplate = `
NAME:
    {{.Name}} - {{.UsageText}}
{{if len .Authors}}
AUTHOR:
    {{range .Authors}}{{ . }}{{end}}
{{end}}
USAGE:
    {{.HelpName}}{{if .VisibleFlags}} [global options]{{end}}{{if .Commands}} command [command options]{{end}}
{{if .Commands}}
COMMANDS:
{{range .Commands}}{{if not .HideHelp}}    {{join .Names ", "}}{{ "\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{end}}{{if .VisibleFlags}}
GLOBAL OPTIONS:
    {{range .VisibleFlags}}{{.}}
    {{end}}{{end}}
`
	app.UsageText = "Web application for Kubernetes CLI configuration with OIDC"
	app.Version = version.Print("loginapp")
	app.Authors = []cli.Author{
		{
			Name:  "fydrah",
			Email: "flav.hardy@gmail.com",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:            "serve",
			Usage:           "Run loginapp application",
			SkipFlagParsing: true,
			ArgsUsage:       "[configuration file]",
			Before: func(c *cli.Context) error {
				return nil
			},
			Action: func(c *cli.Context) error {
				if len(c.Args()) == 0 {
					if err := cli.ShowCommandHelp(c, c.Command.Name); err != nil {
						return fmt.Errorf("Error while rendering command help: %v", err)
					}
					return fmt.Errorf("Missing argument")
				}
				s := &Server{}
				if err := s.config.Init(c.Args().First()); err != nil {
					return err
				}
				if err := s.Run(); err != nil {
					return err
				}
				return nil
			},
		},
	}
	return app
}

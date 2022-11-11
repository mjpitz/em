// Copyright (C) 2022 Mya Pitzeruse
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package jenkins

import (
	"github.com/urfave/cli/v2"
	"go.pitz.tech/em/internal/index"
	"go.pitz.tech/lib/flagset"
)

type AnalyzeConfig struct {
	Index   index.Config `json:"index"`
	Jenkins Config       `json:"jenkins"`
}

var (
	analyzeConfig = &AnalyzeConfig{
		Jenkins: Config{
			Jobs: cli.NewStringSlice(),
		},
	}

	Command = &cli.Command{
		Name:            "jenkins",
		Usage:           "Common operations for working with Jenkins deployments.",
		HideHelpCommand: true,
		Subcommands: []*cli.Command{
			{
				Name:            "builds",
				Usage:           "Common operations for working with Jenkins builds.",
				HideHelpCommand: true,
				Subcommands: []*cli.Command{
					{
						Name:            "analyze",
						Usage:           "Analyze builds in a Jenkins instance",
						Flags:           flagset.ExtractPrefix("em", analyzeConfig),
						HideHelpCommand: true,
						Action: func(ctx *cli.Context) error {
							idx, err := index.Open(analyzeConfig.Index)
							if err != nil {
								return err
							}
							defer idx.Close()

							return Run(ctx.Context, analyzeConfig.Jenkins, idx)
						},
					},
				},
			},
		},
	}
)

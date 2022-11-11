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

package admin

import (
	"github.com/urfave/cli/v2"
)

var (
	Command = &cli.Command{
		Name:            "admin",
		Usage:           "Administrative functions for em.",
		HideHelpCommand: true,
		Subcommands: []*cli.Command{
			{
				Name:            "docs",
				Usage:           "Documentation for em.",
				HideHelpCommand: true,
				Subcommands: []*cli.Command{
					{
						Name:            "markdown",
						Usage:           "Prints markdown documentation for em.",
						HideHelpCommand: true,
						Action: func(ctx *cli.Context) error {
							markdown, err := ctx.App.ToMarkdown()
							if err != nil {
								return err
							}

							_, err = ctx.App.Writer.Write([]byte(markdown))
							return err
						},
					},
					{
						Name:            "man",
						Usage:           "Prints the man page for em.",
						HideHelpCommand: true,
						Action: func(ctx *cli.Context) error {
							man, err := ctx.App.ToMan()
							if err != nil {
								return err
							}

							_, err = ctx.App.Writer.Write([]byte(man))
							return err
						},
					},
				},
			},
		},
	}
)

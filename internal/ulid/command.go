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

package ulid

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/urfave/cli/v2"
	"go.pitz.tech/lib/flagset"
	"go.pitz.tech/lib/ulid"
)

type GenConfig struct {
	Size int    `json:"size" usage:"specify the size of the ulid being generated" default:"256"`
	Out  string `json:"out" alias:"o" usage:"specify the output format (string, bytes)"`
}

type FormatConfig struct {
	In  string `json:"in" alias:"i" usage:"specify the input format (string, bytes)"`
	Out string `json:"out" alias:"o" usage:"specify the output format (json, string, bytes)"`
}

var (
	genConfig = &GenConfig{
		Out: "string",
	}

	formatConfig = &FormatConfig{
		In:  "string",
		Out: "json",
	}

	Command = &cli.Command{
		Name:  "ulid",
		Usage: "Generate or format myago/ulids.",
		Flags: flagset.ExtractPrefix("em", genConfig),
		Subcommands: []*cli.Command{
			{
				Name:  "format",
				Usage: "Parse and format provided myago/ulids.",
				Flags: flagset.ExtractPrefix("", formatConfig),
				Action: func(ctx *cli.Context) error {
					in, err := ioutil.ReadAll(ctx.App.Reader)
					if err != nil {
						return err
					}

					var parsed ulid.ULID

					switch formatConfig.In {
					case "string":
						parsed, err = ulid.Parse(string(in))
					case "bytes":
						parsed = in
					default:
						err = fmt.Errorf("unrecognized input type: %s (available: string, bytes)", formatConfig.In)
					}

					if err != nil {
						return err
					}

					switch formatConfig.Out {
					case "json":
						enc := json.NewEncoder(ctx.App.Writer)
						enc.SetIndent("", "  ")

						err = enc.Encode(map[string]any{
							"skew":    parsed.Skew(),
							"time":    parsed.Timestamp().Local(),
							"payload": parsed.Payload(),
						})
					case "string":
						_, err = ctx.App.Writer.Write([]byte(parsed.String()))
					case "bytes":
						_, err = ctx.App.Writer.Write(parsed.Bytes())
					default:
						err = fmt.Errorf("unrecognized output type: %s (available: json, string, bytes)", formatConfig.Out)
					}

					return err
				},
				HideHelpCommand: true,
			},
		},
		Action: func(ctx *cli.Context) error {
			c := ctx.Context
			ulid, err := ulid.Extract(c).Generate(c, genConfig.Size)
			if err != nil {
				return err
			}

			switch genConfig.Out {
			case "string":
				_, err = ctx.App.Writer.Write([]byte(ulid.String()))
			case "bytes":
				_, err = ctx.App.Writer.Write(ulid.Bytes())
			default:
				err = fmt.Errorf("unrecognized output type: %s (available: string, bytes)", genConfig.Out)
			}

			return err
		},
		HideHelpCommand: true,
	}
)

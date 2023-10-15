// Copyright (C) 2021 Mya Pitzeruse
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

package project

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"go.pitz.tech/em/internal/project/scaffold"
	"go.pitz.tech/em/internal/project/scaffold/licenses"

	"go.pitz.tech/lib/flagset"
	"go.pitz.tech/lib/logger"
	"go.pitz.tech/lib/vfs"
)

const scaffoldHelpTemplate = `
  Features:
    {{- range $feature, $files := .features }}
    - {{ $feature }}
    {{- end }}

  Aliases:
    {{- range $alias, $targets := .aliases }}
    - {{ $alias }}: {{ join $targets ", " }}
    {{- end }}

`

type ScaffoldConfig struct {
	Mkdir    bool             `json:"mkdir"    usage:"specify if we should make the target project directory"`
	License  string           `json:"license"  usage:"specify which license should be applied to the project" default:"agpl3"`
	Features *cli.StringSlice `json:"features" usage:"specify the features to generate"`
}

type IgnoreConfig struct {
	Global     bool   `json:"global" usage:""`
	IgnoreFile string `json:"ignore_file" usage:"" default:".gitignore"`
}

var (
	scaffoldConfig = &ScaffoldConfig{}
	ignoreConfig   = &IgnoreConfig{}

	Command = &cli.Command{
		Name:  "project",
		Usage: "Common operations for working with projects.",
		Subcommands: []*cli.Command{
			{
				Name:            "scaffold",
				Usage:           "Scaffold out a new project or add onto an existing one.",
				Flags:           flagset.ExtractPrefix("em", scaffoldConfig),
				HideHelpCommand: true,
				UsageText: strings.Join([]string{
					"em project scaffold [options] <name>",
					"em project scaffold features    # will output a list of features and aliases",
					"em project scaffold --mkdir --license mpl --features init <name>",
					"em project scaffold --mkdir --license mpl --features init --features bin <name>",
				}, "\n"),
				Action: func(ctx *cli.Context) error {
					if ctx.NArg() == 0 {
						return fmt.Errorf("name not specified")
					}

					name := ctx.Args().Get(0)
					if name == "features" {
						return template.Must(
							template.New("scaffold-help").
								Funcs(map[string]interface{}{
									"join": func(elems []string, sep string) string {
										return strings.Join(elems, sep)
									},
								}).
								Parse(scaffoldHelpTemplate),
						).Execute(ctx.App.Writer, map[string]interface{}{
							"features": scaffold.FilesByFeature,
							"aliases":  scaffold.FeatureAliases,
						})
					}

					license, ok := licenses.ByTemplateName[scaffoldConfig.License]
					if !ok {
						license, ok = licenses.BySPDX[scaffoldConfig.License]
						if !ok {
							return fmt.Errorf("unsupported license: %s", scaffoldConfig.License)
						}
					}

					if scaffoldConfig.Mkdir {
						logger.Extract(ctx.Context).Info("making directory")
						if err := os.MkdirAll(name, 0755); err != nil {
							return errors.Wrap(err, "failed to make project directory")
						}

						if err := os.Chdir(name); err != nil {
							return errors.Wrap(err, "failed to change into directory")
						}
					}

					logger.Extract(ctx.Context).Info("rendering files")
					files := scaffold.Template(
						scaffold.Data{
							Name:     name,
							License:  license.TemplateName,
							SPDX:     license.Identifier,
							Features: scaffoldConfig.Features.Value(),
						},
					).Render(ctx.Context)

					logger.Extract(ctx.Context).Info("writing files")
					afs := vfs.Extract(ctx.Context)
					for _, file := range files {
						dir := filepath.Dir(file.Name)
						_ = afs.MkdirAll(dir, 0755)

						if exists, _ := afero.Exists(afs, file.Name); exists {
							// don't overwrite existing files
							continue
						}

						logger.Extract(ctx.Context).Info("writing file", zap.String("file", file.Name))
						err := afero.WriteFile(afs, file.Name, file.Contents, 0644)
						if err != nil {
							return err
						}
					}

					if scaffoldConfig.Mkdir {
						if exists, _ := afero.Exists(afs, "go.mod"); exists {
							_, err := exec.Command("go", "mod", "tidy").CombinedOutput()
							if err != nil {
								return err
							}
						}
					}

					return nil
				},
			},
			{
				Name:            "ignore",
				Usage:           "Ignore files within a given project.",
				UsageText:       "em project ignore [options] <...patterns>",
				Flags:           flagset.ExtractPrefix("em", ignoreConfig),
				HideHelpCommand: true,
				Action: func(ctx *cli.Context) error {
					log := logger.Extract(ctx.Context)
					cfg := ignoreConfig

					var gitIgnore string
					if cfg.Global {
						homedir, err := os.UserHomeDir()
						if err != nil {
							return err
						}

						gitIgnore = filepath.Join(homedir, cfg.IgnoreFile)
					} else {
						workdir, err := os.Getwd()
						if err != nil {
							return err
						}

						var lastdir string
						for {
							if lastdir == workdir {
								return fmt.Errorf("failed to locate gitignore")
							}

							maybeIgnore := filepath.Join(workdir, cfg.IgnoreFile)
							if _, err := os.Stat(maybeIgnore); err == nil {
								gitIgnore = maybeIgnore
								break
							}

							lastdir = workdir
							workdir = filepath.Dir(filepath.Clean(workdir))
						}
					}

					log.Info("updating gitignore", zap.String("path", gitIgnore))
					f, err := os.OpenFile(gitIgnore, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
					if err != nil {
						return err
					}
					defer f.Close()

					for _, pattern := range ctx.Args().Slice() {
						_, err = f.Write([]byte("\n" + pattern))
						if err != nil {
							return err
						}
					}

					_, err = f.Write([]byte("\n"))

					return err
				},
			},
		},
	}
)

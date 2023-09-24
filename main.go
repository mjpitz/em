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

package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"go.pitz.tech/em/internal/admin"
	"go.pitz.tech/em/internal/ballistics"
	"go.pitz.tech/em/internal/crypto"
	"go.pitz.tech/em/internal/encoding"
	"go.pitz.tech/em/internal/oidc"
	"go.pitz.tech/em/internal/pass"
	"go.pitz.tech/em/internal/project"
	itime "go.pitz.tech/em/internal/time"
	"go.pitz.tech/em/internal/ulid"
	"go.pitz.tech/em/internal/version"

	"go.pitz.tech/lib/flagset"
	"go.pitz.tech/lib/logger"
)

type BuildInfo struct {
	OS           string
	Architecture string

	GoVersion  string
	CGoEnabled bool

	Version  string
	VCS      string
	Revision string
	Compiled time.Time
	Modified bool
}

func (info BuildInfo) Metadata() map[string]any {
	return map[string]any{
		"os":   info.OS,
		"arch": info.Architecture,
		"go":   info.GoVersion,
		"cgo":  strconv.FormatBool(info.CGoEnabled),
		"vcs":  info.VCS,
		"rev":  info.Revision,
		"time": info.Compiled.Format(time.RFC3339),
		"mod":  strconv.FormatBool(info.Modified),
	}
}

func parseBuildInfo() (info BuildInfo) {
	info.OS = runtime.GOOS
	info.Architecture = runtime.GOARCH
	info.GoVersion = strings.TrimPrefix(runtime.Version(), "go")
	info.Compiled = time.Now()

	build, ok := debug.ReadBuildInfo()
	if ok {
		info.Version = build.Main.Version

		for _, setting := range build.Settings {
			switch setting.Key {
			case "CGO_ENABLED":
				info.CGoEnabled, _ = strconv.ParseBool(setting.Value)
			case "vcs":
				info.VCS = setting.Value
			case "vcs.revision":
				info.Revision = setting.Value
			case "vcs.time":
				info.Compiled, _ = time.Parse(time.RFC3339, setting.Value)
			case "vcs.modified":
				info.Modified, _ = strconv.ParseBool(setting.Value)
			}
		}
	}

	return info
}

type GlobalConfig struct {
	Log logger.Config `json:"log"`
}

func main() {
	info := parseBuildInfo()

	config := &GlobalConfig{
		Log: logger.DefaultConfig(),
	}

	app := &cli.App{
		Name:      "em",
		Usage:     "mya's general purpose command line utilities",
		UsageText: "em [options] <command>",
		Version:   info.Version,
		Flags:     flagset.Extract(config),
		Commands: []*cli.Command{
			// order package by abc
			admin.Command,
			ballistics.Command,
			crypto.Command,
			encoding.Command,
			oidc.Command,
			pass.Command,
			project.Command,
			itime.Command,
			ulid.Command,
			version.Command,
		},
		Before: func(ctx *cli.Context) error {
			ctx.Context = logger.Setup(ctx.Context, config.Log)
			ctx.Context, _ = signal.NotifyContext(ctx.Context, os.Interrupt, os.Kill)

			return nil
		},
		Compiled:             info.Compiled,
		HideVersion:          true,
		EnableBashCompletion: true,
		BashComplete:         cli.DefaultAppComplete,
		Suggest:              true,
		Metadata:             info.Metadata(),
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}

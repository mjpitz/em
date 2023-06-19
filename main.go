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
	"runtime"
	"strings"

	"github.com/urfave/cli/v2"
	"go.pitz.tech/em/internal/admin"
	"go.pitz.tech/em/internal/ballistics"
	"go.pitz.tech/em/internal/crypto"
	"go.pitz.tech/em/internal/encoding"
	"go.pitz.tech/em/internal/jenkins"
	"go.pitz.tech/em/internal/oidc"
	"go.pitz.tech/em/internal/project"
	"go.pitz.tech/em/internal/storj"
	"go.pitz.tech/em/internal/time"
	"go.pitz.tech/em/internal/ulid"
	"go.pitz.tech/em/internal/version"

	"go.pitz.tech/lib/lifecycle"

	"go.pitz.tech/lib/flagset"
	"go.pitz.tech/lib/logger"
)

type GlobalConfig struct {
	Log logger.Config `json:"log"`
}

func main() {
	config := &GlobalConfig{
		Log: logger.DefaultConfig(),
	}

	app := &cli.App{
		Name:      "em",
		Usage:     "mya's general purpose command line utilities",
		UsageText: "em [options] <command>",
		Flags:     flagset.Extract(config),
		Commands: []*cli.Command{
			// order package by abc
			admin.Command,
			ballistics.Command,
			crypto.Command,
			encoding.Command,
			jenkins.Command,
			oidc.Command,
			project.Command,
			storj.Command,
			time.Command,
			ulid.Command,
			version.Command,
		},
		Before: func(ctx *cli.Context) error {
			ctx.Context = logger.Setup(ctx.Context, config.Log)
			ctx.Context = lifecycle.Setup(ctx.Context)

			return nil
		},
		HideVersion:          true,
		EnableBashCompletion: true,
		BashComplete:         cli.DefaultAppComplete,
		Suggest:              true,
		Metadata: map[string]interface{}{
			"arch":       runtime.GOARCH,
			"go_version": strings.TrimPrefix(runtime.Version(), "go"),
			"os":         runtime.GOOS,
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}

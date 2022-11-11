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

package oidc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/urfave/cli/v2"
	oidcauth "go.pitz.tech/lib/auth/oidc"
	"go.pitz.tech/lib/browser"
	"go.pitz.tech/lib/flagset"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
)

var (
	authConfig = &oidcauth.Config{
		Scopes: cli.NewStringSlice(),
	}

	Command = &cli.Command{
		Name:            "oidc",
		Usage:           "Common operations for working with OIDC providers.",
		HideHelpCommand: true,
		Subcommands: []*cli.Command{
			{
				Name:            "auth",
				Usage:           "Authenticate with an OIDC provider.",
				Flags:           flagset.ExtractPrefix("em", authConfig),
				HideHelpCommand: true,
				Action: func(ctx *cli.Context) error {
					uri, err := url.Parse(authConfig.RedirectURL)
					if err != nil {
						return err
					}

					svr := &http.Server{
						Addr: uri.Host,
					}

					if len(authConfig.Scopes.Value()) == 0 {
						authConfig.Scopes = cli.NewStringSlice("openid", "profile", "email")
					}

					cctx, cancel := context.WithCancel(ctx.Context)
					defer cancel()

					svr.Handler = oidcauth.ServeMux(*authConfig, func(token *oauth2.Token) {
						defer cancel()

						enc := json.NewEncoder(ctx.App.Writer)
						enc.SetIndent("", "  ")
						_ = enc.Encode(token)
					})

					group := &errgroup.Group{}

					group.Go(func() error {
						time.Sleep(time.Second)
						return browser.Open(ctx.Context, uri.Scheme+"://"+uri.Host+"/login")
					})

					group.Go(svr.ListenAndServe)

					<-cctx.Done()
					err = svr.Shutdown(ctx.Context)
					_ = group.Wait()

					return nil
				},
			},
		},
	}
)

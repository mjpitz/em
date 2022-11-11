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

package storj

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/urfave/cli/v2"
	"go.pitz.tech/em/internal/storj/auth"
	oidcauth "go.pitz.tech/lib/auth/oidc"
	"go.pitz.tech/lib/browser"
	"go.pitz.tech/lib/logger"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"

	"go.pitz.tech/lib/flagset"

	"storj.io/common/uuid"
)

type uuidGenConfig struct {
	Out string `json:"out" alias:"o" usage:"specify the output format (string or bytes)"`
}

type uuidFormatConfig struct {
	uuidGenConfig

	In string `json:"in" alias:"i" usage:"specify the input format (string or bytes)"`
}

var (
	uuidGen = &uuidGenConfig{
		Out: "string",
	}

	uuidFormat = &uuidFormatConfig{
		In: "string",
		uuidGenConfig: uuidGenConfig{
			Out: "bytes",
		},
	}

	authConfig = &oidcauth.Config{
		Scopes: cli.NewStringSlice(),
	}

	Command = &cli.Command{
		Name:            "storj",
		Usage:           "Common operations for working with Storj resources.",
		HideHelpCommand: true,
		Subcommands: []*cli.Command{
			{
				Name:            "auth",
				Usage:           "Authenticate with a Storj OIDC provider.",
				Flags:           flagset.ExtractPrefix("em", authConfig),
				HideHelpCommand: true,
				Action: func(ctx *cli.Context) error {
					uri, err := url.Parse(authConfig.RedirectURL)
					if err != nil {
						return err
					}

					svr := &http.Server{
						Addr: uri.Host,
						BaseContext: func(_ net.Listener) context.Context {
							return ctx.Context
						},
					}

					if len(authConfig.Scopes.Value()) == 0 {
						authConfig.Scopes = cli.NewStringSlice("openid", "profile", "email", "object:list", "object:read", "object:write", "object:delete")
					}

					cctx, cancel := context.WithCancel(ctx.Context)
					defer cancel()

					svr.Handler = auth.ServeMux(*authConfig, func(token *oauth2.Token, rootKey []byte) {
						defer cancel()

						enc := json.NewEncoder(ctx.App.Writer)
						enc.SetIndent("", "  ")
						_ = enc.Encode(struct {
							Token   *oauth2.Token `json:"token"`
							RootKey []byte        `json:"root_key"`
						}{
							Token:   token,
							RootKey: rootKey,
						})
					})

					group := &errgroup.Group{}

					group.Go(func() error {
						time.Sleep(time.Second)
						url := uri.Scheme + "://" + uri.Host + "/login"

						logger.Extract(ctx.Context).Info("Opening " + url)
						return browser.Open(ctx.Context, url)
					})

					group.Go(svr.ListenAndServe)

					<-cctx.Done()
					_ = svr.Shutdown(ctx.Context)
					_ = group.Wait()

					return nil
				},
			},
			{
				Name:            "uuid",
				Usage:           "Format storj-specific UUID.",
				Flags:           flagset.ExtractPrefix("em", uuidGen),
				HideHelpCommand: true,
				Subcommands: []*cli.Command{
					{
						Name:            "format",
						Usage:           "Swap between different formats of the UUID (string and bytes)",
						Flags:           flagset.ExtractPrefix("em", uuidFormat),
						HideHelpCommand: true,
						Action: func(ctx *cli.Context) error {
							in, err := io.ReadAll(ctx.App.Reader)
							if err != nil {
								return err
							}

							var parsed uuid.UUID

							switch uuidFormat.In {
							case "string":
								parsed, err = uuid.FromString(string(in))
							case "bytes":
								parsed, err = uuid.FromBytes(in)
							default:
								err = fmt.Errorf("unrecognized input type: %s (available: string, bytes)", uuidFormat.In)
							}

							if err != nil {
								return err
							}

							switch uuidFormat.Out {
							case "string":
								_, err = ctx.App.Writer.Write([]byte(parsed.String()))
							case "bytes":
								_, err = ctx.App.Writer.Write(parsed.Bytes())
							default:
								err = fmt.Errorf("unrecognized output type: %s (available: string, bytes)", uuidFormat.Out)
							}

							return err
						},
					},
				},
				Action: func(ctx *cli.Context) error {
					uuid, err := uuid.New()
					if err != nil {
						return err
					}

					switch uuidGen.Out {
					case "string":
						_, err = ctx.App.Writer.Write([]byte(uuid.String()))
					case "bytes":
						_, err = ctx.App.Writer.Write(uuid.Bytes())
					default:
						err = fmt.Errorf("unrecognized output type: %s (available: string, bytes)", uuidFormat.Out)
					}

					return err
				},
			},
		},
	}
)

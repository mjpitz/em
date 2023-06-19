package time

import (
	"fmt"
	"strconv"
	"time"

	"github.com/urfave/cli/v2"
)

var (
	day = 24 * time.Hour

	supportedFormats = []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02",
		time.Kitchen,
	}

	Command = &cli.Command{
		Name:  "time",
		Usage: "Common operations handling time.",
		Subcommands: []*cli.Command{
			{
				Name:  "since",
				Usage: "Compute how much time has elapsed since the provided time.",
				Action: func(ctx *cli.Context) error {
					if ctx.NArg() < 1 {
						return fmt.Errorf("missing time as first argument")
					}

					arg0 := ctx.Args().Get(0)

					var t time.Time
					var err error

					for _, format := range supportedFormats {
						t, err = time.Parse(format, arg0)
						if err == nil {
							break
						}
					}

					if err != nil {
						return fmt.Errorf("failed to parse input time using known formats")
					}

					elapsed := time.Since(t)
					out := ""

					if elapsed > day {
						out += strconv.FormatInt(int64(elapsed/day), 10) + "d"
						elapsed = elapsed % day
					}

					out += elapsed.String()
					_, err = ctx.App.Writer.Write([]byte(out))

					return err
				},
			},
		},
		Action: func(ctx *cli.Context) error {
			now := time.Now().Format(time.RFC3339)
			_, err := ctx.App.Writer.Write([]byte(now))

			return err
		},
	}
)

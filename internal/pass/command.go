package pass

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/urfave/cli/v2"
	"go.pitz.tech/lib/dirset"
	"go.pitz.tech/lib/flagset"
	"go.pitz.tech/lib/pass"
	"os"
	"path/filepath"
)

type DeriveConfig struct {
	Passphrase string `json:"passphrase" usage:"the root passphrase all subsequent passphrases are derived from"`
	User       string `json:"user" usage:"the name of the individual authenticating with the remote system"`
	Site       string `json:"site" usage:"specify which site we're deriving a password for"`
	Template   string `json:"template" usage:"which template pattern to use [max,long,medium,short,basic,pin,code]" default:"max"`
}

type RotateConfig struct {
	Passphrase string `json:"passphrase" usage:"the root passphrase all subsequent passphrases are derived from"`
	User       string `json:"user" usage:"the name of the individual authenticating with the remote system"`
	Site       string `json:"site" usage:"specify which site we're deriving a password for"`
}

type PurgeConfig struct {
	Confirm bool `json:"confirm" usage:"confirm that you want to purge the local database"`
}

var (
	deriveConfig = &DeriveConfig{}
	rotateConfig = &RotateConfig{}
	purgeConfig  = &PurgeConfig{}

	templates = map[string]pass.TemplateClass{
		string(pass.MaximumSecurity):  pass.MaximumSecurity,
		string(pass.Long):             pass.Long,
		string(pass.Medium):           pass.Medium,
		string(pass.Short):            pass.Short,
		string(pass.Basic):            pass.Basic,
		string(pass.PIN):              pass.PIN,
		string(pass.VerificationCode): pass.VerificationCode,
	}

	Command = &cli.Command{
		Name:            "pass",
		Usage:           "Common operations for working with passwords.",
		HideHelpCommand: true,
		Subcommands: []*cli.Command{
			{
				Name:            "derive",
				Usage:           "Derive the current password for the given site.",
				Flags:           flagset.ExtractPrefix("em", deriveConfig),
				HideHelpCommand: true,
				Action: func(ctx *cli.Context) error {
					cfg := deriveConfig

					dirset := dirset.Must(ctx.App.Name)
					err := os.MkdirAll(dirset.StateDir, 0755)
					if err != nil {
						return err
					}

					generations, err := loadGenerations(dirset.StateDir)
					if err != nil {
						return err
					}

					template, ok := templates[cfg.Template]
					if !ok {
						return fmt.Errorf("unrecognized template: %s", cfg.Template)
					}

					key := hashKey(cfg.Passphrase, cfg.User, cfg.Site)
					generation, ok := generations[key]
					if !ok {
						generation = 1
					}

					identity, err := pass.Identity(pass.Authentication, []byte(cfg.Passphrase), cfg.User)
					if err != nil {
						return err
					}

					siteKey := pass.SiteKey(pass.Authentication, identity, cfg.Site, generation)
					password := pass.SitePassword(siteKey, template)

					_, err = ctx.App.Writer.Write(password)

					return err
				},
			},
			{
				Name:            "rotate",
				Usage:           "Rotate the current password for the given site.",
				Flags:           flagset.ExtractPrefix("em", rotateConfig),
				HideHelpCommand: true,
				Action: func(ctx *cli.Context) error {
					cfg := rotateConfig

					dirset := dirset.Must(ctx.App.Name)
					err := os.MkdirAll(dirset.StateDir, 0755)
					if err != nil {
						return err
					}

					generations, err := loadGenerations(dirset.StateDir)
					if err != nil {
						return err
					}

					key := hashKey(cfg.Passphrase, cfg.User, cfg.Site)

					if _, ok := generations[key]; !ok {
						generations[key] = 1
					}

					generations[key]++

					err = writeGenerations(dirset.StateDir, generations)
					if err != nil {
						return err
					}

					ctx.App.Writer.Write([]byte("rotated!\n"))

					return nil
				},
			},
			{
				Name:            "purge",
				Usage:           "Purges the local database.",
				Flags:           flagset.ExtractPrefix("em", purgeConfig),
				HideHelpCommand: true,
				Action: func(ctx *cli.Context) error {
					cfg := purgeConfig

					if !cfg.Confirm {
						return fmt.Errorf("please confirm you want to perform this action by passing the `--confirm` flag")
					}

					dirset := dirset.Must(ctx.App.Name)
					_ = os.Remove(filepath.Join(dirset.StateDir, "pass-generations.json"))

					return nil
				},
			},
		},
	}
)

// hashKey computes a dictionary key using the passphrase as an hmac seed, and the user and site as the hash key.
func hashKey(passphrase, user, site string) string {
	hmacKey := sha256.Sum256([]byte(passphrase))
	hmac := hmac.New(sha256.New, hmacKey[:])

	_, _ = hmac.Write([]byte(user + "@" + site))

	return base64.URLEncoding.EncodeToString(hmac.Sum(nil))
}

func loadGenerations(directory string) (map[string]uint32, error) {
	generations := make(map[string]uint32)

	if handle, err := os.Open(filepath.Join(directory, "pass-generations.json")); err == nil {
		defer handle.Close()

		err = json.NewDecoder(handle).Decode(&generations)
		if err != nil {
			return nil, err
		}
	}

	return generations, nil
}

func writeGenerations(directory string, generations map[string]uint32) error {
	handle, err := os.OpenFile(filepath.Join(directory, "pass-generations.json"), os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer handle.Close()

	err = json.NewEncoder(handle).Encode(generations)

	return err
}

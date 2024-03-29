package flags

import (
	"github.com/urfave/cli/v2"

	"github.com/warehouse-13/hammertime/pkg/config"
	"github.com/warehouse-13/hammertime/pkg/defaults"
)

// WithFlagsFunc can be used with CLIFlags to build a list of flags for a
// command.
type WithFlagsFunc func() []cli.Flag

// CLIFlags takes a list of WithFlagsFunc options and returns a list of flags
// for a command.
func CLIFlags(options ...WithFlagsFunc) []cli.Flag {
	flags := []cli.Flag{}

	for _, group := range options {
		flags = append(flags, group()...)
	}

	return flags
}

// WithGRPCAddressFlag adds the flintlock GRPC address flag to the command.
func WithGRPCAddressFlag() WithFlagsFunc {
	return func() []cli.Flag {
		return []cli.Flag{
			&cli.StringFlag{
				Name:    "grpc-address",
				Value:   defaults.DialTarget,
				Aliases: []string{"a"},
				Usage:   "flintlock server address + port",
			},
		}
	}
}

// WithNameAndNamespaceFlags adds the name and namespace flags to the command.
func WithNameAndNamespaceFlags(withDefaults bool) WithFlagsFunc {
	nameFlag := &cli.StringFlag{
		Name:    "name",
		Aliases: []string{"n"},
		Usage:   "microvm name",
	}
	namespaceFlag := &cli.StringFlag{
		Name:    "namespace",
		Aliases: []string{"ns"},
		Usage:   "microvm namespace",
	}

	if withDefaults {
		nameFlag.Value = defaults.MvmName
		namespaceFlag.Value = defaults.MvmNamespace
	}

	return func() []cli.Flag {
		return []cli.Flag{
			nameFlag,
			namespaceFlag,
		}
	}
}

// WithJSONSpecFlag adds the json file flag to the command.
func WithJSONSpecFlag() WithFlagsFunc {
	return func() []cli.Flag {
		return []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "path to json file containing full flintlock spec. will override other flags",
			},
		}
	}
}

// WithSSHKeyFlag adds the public-key-path flag to the command.
func WithSSHKeyFlag() WithFlagsFunc {
	return func() []cli.Flag {
		return []cli.Flag{
			&cli.StringFlag{
				Name:    "public-key-path",
				Aliases: []string{"k"},
				Usage:   "path to file containing public SSH key to be added to root user",
			},
		}
	}
}

// WithIDFlag adds the id flag to the command.
func WithIDFlag() WithFlagsFunc {
	return func() []cli.Flag {
		return []cli.Flag{
			&cli.StringFlag{
				Name:    "id",
				Aliases: []string{"i"},
				Usage:   "microvm uuid",
			},
		}
	}
}

// WithStateFlag adds the boolean state flag to the command.
func WithStateFlag() WithFlagsFunc {
	return func() []cli.Flag {
		return []cli.Flag{
			&cli.BoolFlag{
				Name:    "state",
				Value:   false,
				Aliases: []string{"s"},
				Usage:   "print just the state of the microvm",
			},
		}
	}
}

// WithAllFlag adds the boolean all flag to the command.
func WithAllFlag() WithFlagsFunc {
	return func() []cli.Flag {
		return []cli.Flag{
			&cli.BoolFlag{
				Name:  "all",
				Usage: "delete all microvms (filter with --name and --namespace)",
			},
		}
	}
}

// WithQuietFlag adds a silent flag to the command.
func WithQuietFlag() WithFlagsFunc {
	return func() []cli.Flag {
		return []cli.Flag{
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "silence the output on the command",
			},
		}
	}
}

func WithBasicAuthFlag() WithFlagsFunc {
	return func() []cli.Flag {
		return []cli.Flag{
			&cli.StringFlag{
				Name:    "token",
				Aliases: []string{"t"},
				Usage:   "provide a token if basic auth is set on the server",
			},
		}
	}
}

// ParseFlags processes all flags on the CLI context and builds a config object
// which will be used in the command's action.
func ParseFlags(cfg *config.Config) cli.BeforeFunc {
	return func(ctx *cli.Context) error {
		cfg.GRPCAddress = ctx.String("grpc-address")
		cfg.Token = ctx.String("token")

		cfg.MvmName = ctx.String("name")
		cfg.MvmNamespace = ctx.String("namespace")

		cfg.JSONFile = ctx.String("file")
		cfg.SSHKeyPath = ctx.String("public-key-path")

		cfg.State = ctx.Bool("state")
		cfg.DeleteAll = ctx.Bool("all")
		cfg.Silent = ctx.Bool("quiet")

		cfg.UUID = ctx.String("id")

		return nil
	}
}

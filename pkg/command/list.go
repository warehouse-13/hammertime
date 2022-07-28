package command

import (
	"github.com/urfave/cli/v2"

	"github.com/warehouse-13/hammertime/pkg/client"
	"github.com/warehouse-13/hammertime/pkg/config"
	"github.com/warehouse-13/hammertime/pkg/flags"
	"github.com/warehouse-13/hammertime/pkg/utils"
)

func listCommand() *cli.Command {
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: client.New,
		},
	}

	return &cli.Command{
		Name:    "list",
		Usage:   "list microvms",
		Aliases: []string{"l"},
		Before:  flags.ParseFlags(cfg),
		Flags: flags.CLIFlags(
			flags.WithGRPCAddressFlag(),
			flags.WithNameAndNamespaceFlags(false),
		),
		Action: func(c *cli.Context) error {
			return ListFn(cfg)
		},
	}
}

// TODO: add tests as part of #54.
func ListFn(cfg *config.Config) error {
	client, err := cfg.ClientBuilderFunc(cfg.GRPCAddress)
	if err != nil {
		return err
	}

	defer client.Close()

	res, err := client.List(cfg.MvmName, cfg.MvmNamespace)
	if err != nil {
		return err
	}

	return utils.PrettyPrint(res)
}

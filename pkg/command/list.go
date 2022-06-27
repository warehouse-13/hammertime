package command

import (
	"github.com/urfave/cli/v2"
	"github.com/weaveworks/flintlock/api/services/microvm/v1alpha1"

	"github.com/warehouse-13/hammertime/pkg/client"
	"github.com/warehouse-13/hammertime/pkg/config"
	"github.com/warehouse-13/hammertime/pkg/dialler"
	"github.com/warehouse-13/hammertime/pkg/flags"
	"github.com/warehouse-13/hammertime/pkg/utils"
)

func listCommand() *cli.Command {
	cfg := &config.Config{}

	return &cli.Command{
		Name:    "list",
		Usage:   "list microvms",
		Aliases: []string{"l"},
		Before:  flags.ParseFlags(cfg),
		Flags: flags.CLIFlags(
			flags.WithGRPCAddressFlag(),
			flags.WithNameAndNamespaceFlags(),
		),
		Action: func(c *cli.Context) error {
			return listFn(cfg)
		},
	}
}

func listFn(cfg *config.Config) error {
	conn, err := dialler.New(cfg.GRPCAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := client.New(v1alpha1.NewMicroVMClient(conn))

	res, err := client.List(cfg.MvmName, cfg.MvmNamespace)
	if err != nil {
		return err
	}

	return utils.PrettyPrint(res)
}

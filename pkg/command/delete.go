package command //nolint: dupl // duplication here is fine

import (
	"github.com/urfave/cli/v2"
	"github.com/weaveworks/flintlock/api/services/microvm/v1alpha1"

	"github.com/warehouse-13/hammertime/pkg/client"
	"github.com/warehouse-13/hammertime/pkg/config"
	"github.com/warehouse-13/hammertime/pkg/dialler"
	"github.com/warehouse-13/hammertime/pkg/flags"
	"github.com/warehouse-13/hammertime/pkg/microvm"
	"github.com/warehouse-13/hammertime/pkg/utils"
)

func deleteCommand() *cli.Command {
	cfg := &config.Config{}

	return &cli.Command{
		Name:    "delete",
		Usage:   "delete a microvmd",
		Aliases: []string{"d"},
		Before:  flags.ParseFlags(cfg),
		Flags: flags.CLIFlags(
			flags.WithGRPCAddressFlag(),
			flags.WithNameAndNamespaceFlags(),
			flags.WithIDFlag(),
			flags.WithJSONSpecFlag(),
			flags.WithAllFlag(),
		),
		Action: func(c *cli.Context) error {
			return deleteFn(cfg)
		},
	}
}

func deleteFn(cfg *config.Config) error {
	conn, err := dialler.New(cfg.GRPCAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	mngr := microvm.NewManager(
		client.New(
			v1alpha1.NewMicroVMClient(conn),
		),
	)

	res, err := mngr.Delete(cfg)
	if err != nil {
		return err
	}

	return utils.PrettyPrint(res)
}

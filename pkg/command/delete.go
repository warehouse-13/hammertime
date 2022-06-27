package command

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"github.com/weaveworks/flintlock/api/services/microvm/v1alpha1"

	"github.com/warehouse-13/hammertime/pkg/client"
	"github.com/warehouse-13/hammertime/pkg/config"
	"github.com/warehouse-13/hammertime/pkg/dialler"
	"github.com/warehouse-13/hammertime/pkg/flags"
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

func deleteFn(cfg *config.Config) error { //nolint: cyclop // we are refactoring this file
	conn, err := dialler.New(cfg.GRPCAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := client.New(v1alpha1.NewMicroVMClient(conn))

	if utils.IsSet(cfg.JSONFile) {
		var err error

		cfg.UUID, cfg.MvmName, cfg.MvmNamespace, err = utils.ProcessFile(cfg.JSONFile)
		if err != nil {
			return err
		}
	}

	if utils.IsSet(cfg.UUID) {
		res, err := client.Delete(cfg.UUID)
		if err != nil {
			return err
		}

		return utils.PrettyPrint(res)
	}

	if cfg.DeleteAll {
		if utils.IsSet(cfg.MvmName) && !utils.IsSet(cfg.MvmNamespace) {
			return fmt.Errorf("required: --namespace")
		}
	} else {
		if utils.IsSet(cfg.MvmName) && !utils.IsSet(cfg.MvmNamespace) {
			return fmt.Errorf("required: --namespace")
		}
		if !utils.IsSet(cfg.MvmName) && utils.IsSet(cfg.MvmNamespace) {
			return fmt.Errorf("required: --name")
		}
	}

	list, err := client.List(cfg.MvmName, cfg.MvmNamespace)
	if err != nil {
		return err
	}

	if utils.IsSet(cfg.MvmName) && utils.IsSet(cfg.MvmNamespace) && !cfg.DeleteAll {
		if len(list.Microvm) > 1 {
			fmt.Printf("%d MicroVMs found under %s/%s:\n", len(list.Microvm), cfg.MvmNamespace, cfg.MvmName)

			for _, mvm := range list.Microvm {
				fmt.Println(*mvm.Spec.Uid)
			}

			return nil
		}
	}

	for _, mvm := range list.Microvm {
		res, err := client.Delete(*mvm.Spec.Uid)
		if err != nil {
			return err
		}

		if err := utils.PrettyPrint(res); err != nil {
			return err
		}
	}

	return nil
}

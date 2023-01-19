package command

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/warehouse-13/hammertime/pkg/client"
	"github.com/warehouse-13/hammertime/pkg/config"
	"github.com/warehouse-13/hammertime/pkg/flags"
	"github.com/warehouse-13/hammertime/pkg/utils"
)

func deleteCommand() *cli.Command {
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: client.New,
		},
	}

	w := utils.NewWriter(os.Stdout)

	return &cli.Command{
		Name:    "delete",
		Usage:   "delete a microvmd",
		Aliases: []string{"d"},
		Before:  flags.ParseFlags(cfg),
		Flags: flags.CLIFlags(
			flags.WithGRPCAddressFlag(),
			flags.WithNameAndNamespaceFlags(false),
			flags.WithIDFlag(),
			flags.WithJSONSpecFlag(),
			flags.WithAllFlag(),
			flags.WithQuietFlag(),
			flags.WithBasicAuthFlag(),
		),
		Action: func(c *cli.Context) error {
			return DeleteFn(w, cfg)
		},
	}
}

func DeleteFn(w utils.Writer, cfg *config.Config) error { //nolint: cyclop // we are refactoring this file
	client, err := cfg.ClientBuilderFunc(cfg.GRPCAddress, cfg.Token)
	if err != nil {
		return err
	}

	defer client.Close()

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

		if cfg.Silent {
			return nil
		}

		return w.PrettyPrint(res)
	}

	if !cfg.DeleteAll {
		if !utils.IsSet(cfg.MvmName) || !utils.IsSet(cfg.MvmNamespace) {
			return fmt.Errorf("required: --namespace, --name")
		}
	}

	list, err := client.List(cfg.MvmName, cfg.MvmNamespace)
	if err != nil {
		return err
	}

	if utils.IsSet(cfg.MvmName) && utils.IsSet(cfg.MvmNamespace) && !cfg.DeleteAll {
		if len(list.Microvm) > 1 {
			w.Printf("%d MicroVMs found under %s/%s:\n", len(list.Microvm), cfg.MvmNamespace, cfg.MvmName)

			for _, mvm := range list.Microvm {
				w.Print(*mvm.Spec.Uid)
			}

			return nil
		}
	}

	for _, mvm := range list.Microvm {
		res, err := client.Delete(*mvm.Spec.Uid)
		if err != nil {
			return err
		}

		if !cfg.Silent {
			if err := w.PrettyPrint(res); err != nil {
				return err
			}
		}
	}

	return nil
}

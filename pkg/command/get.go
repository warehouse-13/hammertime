package command

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/warehouse-13/hammertime/pkg/client"
	"github.com/warehouse-13/hammertime/pkg/config"
	"github.com/warehouse-13/hammertime/pkg/flags"
	"github.com/warehouse-13/hammertime/pkg/utils"
)

func getCommand() *cli.Command {
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: client.New,
		},
	}

	return &cli.Command{
		Name:    "get",
		Usage:   "get an existing microvm",
		Aliases: []string{"g"},
		Before:  flags.ParseFlags(cfg),
		Flags: flags.CLIFlags(
			flags.WithGRPCAddressFlag(),
			flags.WithNameAndNamespaceFlags(true),
			flags.WithJSONSpecFlag(),
			flags.WithStateFlag(),
			flags.WithIDFlag(),
		),
		Action: func(c *cli.Context) error {
			return GetFn(cfg)
		},
	}
}

// TODO: add tests as part of #54.
func GetFn(cfg *config.Config) error {
	client, err := cfg.ClientBuilderFunc(cfg.GRPCAddress)
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
		res, err := client.Get(cfg.UUID)
		if err != nil {
			return err
		}

		if cfg.State {
			fmt.Println(res.Microvm.Status.State)

			return nil
		}

		return utils.PrettyPrint(res)
	}

	res, err := client.List(cfg.MvmName, cfg.MvmNamespace)
	if err != nil {
		return err
	}

	if len(res.Microvm) > 1 {
		fmt.Printf("%d MicroVMs found under %s/%s:\n", len(res.Microvm), cfg.MvmNamespace, cfg.MvmName)

		for _, mvm := range res.Microvm {
			fmt.Println(*mvm.Spec.Uid)
		}

		return nil
	}

	if len(res.Microvm) == 1 {
		if cfg.State {
			fmt.Println(res.Microvm[0].Status.State)

			return nil
		}

		return utils.PrettyPrint(res.Microvm[0])
	}

	return fmt.Errorf("MicroVM %s/%s not found", cfg.MvmName, cfg.MvmNamespace)
}

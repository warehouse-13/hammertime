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

func getCommand() *cli.Command {
	cfg := &config.Config{
		ClientConfig: config.ClientConfig{
			ClientBuilderFunc: client.New,
		},
	}

	w := utils.NewWriter(os.Stdout)

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
			flags.WithBasicAuthFlag(),
		),
		Action: func(c *cli.Context) error {
			return GetFn(w, cfg)
		},
	}
}

func GetFn(w utils.Writer, cfg *config.Config) error {
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
		res, err := client.Get(cfg.UUID)
		if err != nil {
			return err
		}

		if cfg.State {
			w.Print(res.Microvm.Status.State)

			return nil
		}

		return w.PrettyPrint(res)
	}

	res, err := client.List(cfg.MvmName, cfg.MvmNamespace)
	if err != nil {
		return err
	}

	if len(res.Microvm) > 1 {
		w.Printf("%d MicroVMs found under %s/%s:\n", len(res.Microvm), cfg.MvmNamespace, cfg.MvmName)

		for _, mvm := range res.Microvm {
			w.Print(*mvm.Spec.Uid)
		}

		return nil
	}

	if len(res.Microvm) == 1 {
		if cfg.State {
			w.Print(res.Microvm[0].Status.State)

			return nil
		}

		return w.PrettyPrint(res.Microvm[0])
	}

	return fmt.Errorf("MicroVM %s/%s not found", cfg.MvmNamespace, cfg.MvmName)
}

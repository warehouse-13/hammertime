package command

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/weaveworks-liquidmetal/flintlock/api/types"

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
	if utils.IsSet(cfg.JSONFile) {
		var err error

		cfg.UUID, cfg.MvmName, cfg.MvmNamespace, err = utils.ProcessFile(cfg.JSONFile)
		if err != nil {
			return err
		}
	}

	res, err := findMicrovm(cfg)
	if err != nil {
		return err
	}

	if len(res) == 1 {
		if cfg.State {
			w.Print(res[0].Status.State)

			return nil
		}

		return w.PrettyPrint(res[0])
	}

	if len(res) > 1 {
		w.Printf("%d MicroVMs found under %s/%s:\n", len(res), cfg.MvmNamespace, cfg.MvmName)

		for _, mvm := range res {
			w.Print(*mvm.Spec.Uid)
		}

		return nil
	}

	return fmt.Errorf("MicroVM %s/%s not found", cfg.MvmNamespace, cfg.MvmName)
}

func findMicrovm(cfg *config.Config) ([]*types.MicroVM, error) {
	client, err := cfg.ClientBuilderFunc(cfg.GRPCAddress, cfg.Token)
	if err != nil {
		return nil, err
	}

	defer client.Close()

	if utils.IsSet(cfg.UUID) {
		res, err := client.Get(cfg.UUID)
		if err != nil {
			return nil, err
		}

		return []*types.MicroVM{res.Microvm}, nil
	}

	res, err := client.List(cfg.MvmName, cfg.MvmNamespace)
	if err != nil {
		return nil, err
	}

	return res.Microvm, nil
}

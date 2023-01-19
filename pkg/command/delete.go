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

func DeleteFn(w utils.Writer, cfg *config.Config) error {
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

	// If it is possible to delete by set UUID, do that and exit
	if utils.IsSet(cfg.UUID) {
		return deleteMvm(w, client, cfg.UUID, cfg.Silent)
	}

	// If UUID is not present, make sure that required spec is set
	if missingSpec(cfg) {
		return fmt.Errorf("required: --namespace, --name")
	}

	// Get all microvms
	list, err := client.List(cfg.MvmName, cfg.MvmNamespace)
	if err != nil {
		return err
	}

	// Do not auto-delete multple mvms, inform and exit
	if len(list.Microvm) > 1 && doNotDeleteAll(cfg) {
		w.Printf("%d MicroVMs found under %s/%s:\n", len(list.Microvm), cfg.MvmNamespace, cfg.MvmName)

		for _, mvm := range list.Microvm {
			w.Print(*mvm.Spec.Uid)
		}

		w.Print("\nTo delete all microvms in this list, re-run command with `--all`.")

		return nil
	}

	// By this point we assume the user wants everything dead
	for _, mvm := range list.Microvm {
		if err := deleteMvm(w, client, *mvm.Spec.Uid, cfg.Silent); err != nil {
			return err
		}
	}

	return nil
}

func deleteMvm(w utils.Writer, c client.FlintlockClient, u string, s bool) error {
	res, err := c.Delete(u)
	if err != nil {
		return err
	}

	if s {
		return nil
	}

	return w.PrettyPrint(res)
}

func missingSpec(cfg *config.Config) bool {
	return !cfg.DeleteAll && (!utils.IsSet(cfg.MvmName) || !utils.IsSet(cfg.MvmNamespace))
}

func doNotDeleteAll(cfg *config.Config) bool {
	return utils.IsSet(cfg.MvmName) && utils.IsSet(cfg.MvmNamespace) && !cfg.DeleteAll
}

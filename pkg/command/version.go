package command

import (
	"os"

	"github.com/urfave/cli/v2"

	"github.com/warehouse-13/hammertime/pkg/utils"
	"github.com/warehouse-13/hammertime/pkg/version"
)

func versionCommand() *cli.Command {
	return &cli.Command{
		Name:    "version",
		Usage:   "print the version number for hammertime",
		Aliases: []string{"v"},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "long",
				Value:   false,
				Aliases: []string{"l"},
				Usage:   "print the long version information",
			},
		},
		Action: VersionFn,
	}
}

type versionInfo struct {
	PackageName string
	Version     string
	CommitHash  string
	BuildDate   string
}

func VersionFn(ctx *cli.Context) error {
	w := utils.NewWriter(os.Stdout)

	if ctx.Bool("long") {
		info := versionInfo{
			version.PackageName,
			version.Version,
			version.CommitHash,
			version.BuildDate,
		}

		return w.PrettyPrint(info)
	}

	w.Print(version.Version)

	return nil
}

package command

import (
	"fmt"

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
	if ctx.Bool("long") {
		info := versionInfo{
			version.PackageName,
			version.Version,
			version.CommitHash,
			version.BuildDate,
		}

		return utils.PrettyPrint(info)
	}

	fmt.Println(version.Version)

	return nil
}

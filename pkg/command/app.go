package command

import (
	"io"

	"github.com/urfave/cli/v2"
)

// NewApp is a builder which returns a cli.App.
func NewApp(out io.Writer) *cli.App {
	app := cli.NewApp()

	if out != nil {
		app.Writer = out
	}

	app.Name = "hammertime"
	// TODO let's have a usage builder func here #48
	app.Usage = "a basic cli client to flintlock"
	app.EnableBashCompletion = true
	app.Commands = commands()

	return app
}

func commands() []*cli.Command {
	return []*cli.Command{
		createCommand(),
		getCommand(),
		listCommand(),
		deleteCommand(),
	}
}

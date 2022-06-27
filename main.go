package main

import (
	"log"
	"os"

	"github.com/warehouse-13/hammertime/pkg/command"
)

func main() {
	app := command.NewApp(os.Stdout)

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

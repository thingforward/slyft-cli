package main

import (
	"os"
	"time"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Name = "Slyft"
	app.Version = "0.0.0"
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Dr. Sandeep Sadandandan",
			Email: "grep@whybenormal.org",
		},
	}
	app.Copyright = "(c) Digi Inc."
	app.Usage = "Help you to connect to Slyft Server"
	app.Commands = CommandStructure // see command_structure.go

	app.Run(os.Args)
}

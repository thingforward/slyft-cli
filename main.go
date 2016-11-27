package main

import (
	"os"
	"time"

	"github.com/urfave/cli"
)

// To be honest, I am not really happy about having these globals here.
// But I am going to keep them here - at least until the alpha is ready.
// The *preferred* way can be seen here: https://github.com/urfave/cli#flags
var ProjectName string
var Everything bool

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
	app.Flags = CommandFlags        // see command_flags.go
	app.Commands = CommandStructure // see command_structure.go

	app.Run(os.Args)
}

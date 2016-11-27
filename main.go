package main

import (
	"fmt"
	"os"
	"time"

	"github.com/op/go-logging"
	"github.com/urfave/cli"
)

// To be honest, I am not really happy about having these globals here.
// But I am going to keep them here - at least until the alpha is ready.
// The *preferred* way can be seen here: https://github.com/urfave/cli#flags
var ProjectName string
var Everything bool
var BackendBaseUrl = os.Getenv("SLYFTBACKEND")

var log = logging.MustGetLogger("ibtlogger")
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

func setupLogger() {
	logBackend := logging.NewLogBackend(os.Stderr, "", 0)
	logBackendWithFormat := logging.NewBackendFormatter(logBackend, format)
	loggerLeveled := logging.AddModuleLevel(logBackendWithFormat)

	level, err := logging.LogLevel(os.Getenv("DEBUGLEVEL"))
	if err != nil {
		fmt.Println("Logging level is set to %s", level)
	}

	loggerLeveled.SetLevel(level, "")
	logging.SetBackend(loggerLeveled)
}

func init() {
	setupLogger()
	if BackendBaseUrl == "" {
		log.Fatal("Backend URL missing, please contact tech support")
	}
}

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

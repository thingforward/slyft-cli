package main

import (
	"net/http"
	"os"
	"time"

	"github.com/jawher/mow.cli"
	"github.com/op/go-logging"
)

var BackendBaseUrl = os.Getenv("SLYFTBACKEND")

var Log = logging.MustGetLogger("ibtlogger")
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

func setupLogger() {
	logBackend := logging.NewLogBackend(os.Stderr, "", 0)
	logBackendWithFormat := logging.NewBackendFormatter(logBackend, format)
	loggerLeveled := logging.AddModuleLevel(logBackendWithFormat)

	level, _ := logging.LogLevel(os.Getenv("DEBUGLEVEL"))

	loggerLeveled.SetLevel(level, "")
	logging.SetBackend(loggerLeveled)
}

func init() {
	// https://www.reddit.com/r/golang/comments/45mzie/dont_use_gos_default_http_client/
	http.DefaultClient.Timeout = 10 * time.Second

	setupLogger()

	// If Environment variable SLYFTBACKEND is present, take it. Must be a full URL
	if BackendBaseUrl == "" {
		// If not, set standard production backend
		BackendBaseUrl = "https://api.slyft.io/"
	}
}

func main() {
	app := cli.App("Slyft", "")

	app.Version("v version", "0.0.1")
	//app.Name = "Slyft"
	//app.Version = "0.0.0"
	//app.Compiled = time.Now()
	//app.Authors = []cli.Author{
	//cli.Author{
	//Name:  "Dr. Sandeep Sadandandan",
	//Email: "grep@whybenormal.org",
	//},
	//}
	//app.Copyright = "(c) Digi Inc."

	app.Command("user account", "Account management", RegisterUserRoutes)
	app.Command("project p", "Project management", RegisterProjectRoutes)
	app.Command("asset a", "Asset management", RegisterAssetRoutes)

	app.Run(os.Args)
}

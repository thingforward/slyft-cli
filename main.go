//
//   Copyright 2016, 2017 Digital Incubation and Growth GmbH
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
//
package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jawher/mow.cli"
	"github.com/op/go-logging"
)

var VERSION = "0.3.1"

var BackendBaseUrl = os.Getenv("SLYFTBACKEND")

var Log = logging.MustGetLogger("ibtlogger")
var format_dbg = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)
var format = logging.MustStringFormatter(
	`%{level:.4s} %{id:03x} %{message}`,
)
var fDebug *bool

func getLogFormat() logging.Formatter {
	// if debug, use timestamps to correlate with server actions
	if os.Getenv("DEBUGLEVEL") == "DEBUG" {
		return format_dbg
	}
	return format
}

func SetupLogger() {
	if fDebug != nil && *fDebug {
		fmt.Println("Setting --debug")
		os.Setenv("DEBUGLEVEL", "DEBUG")
	}
	logBackend := logging.NewLogBackend(os.Stderr, "", 0)
	logBackendWithFormat := logging.NewBackendFormatter(logBackend, getLogFormat())
	loggerLeveled := logging.AddModuleLevel(logBackendWithFormat)

	level, _ := logging.LogLevel(os.Getenv("DEBUGLEVEL"))

	loggerLeveled.SetLevel(level, "")
	logging.SetBackend(loggerLeveled)
}

func init() {
	// https://www.reddit.com/r/golang/comments/45mzie/dont_use_gos_default_http_client/
	http.DefaultClient.Timeout = 10 * time.Second

	SetupLogger()

	// If Environment variable SLYFTBACKEND is present, take it. Must be a full URL
	if BackendBaseUrl == "" {
		// If not, set standard production backend
		BackendBaseUrl = "https://api.slyft.io/"
	}
}

func showBanner() {
	fmt.Printf(`
       .__           _____  __
  _____|  | ___.__._/ ____\/  |_	slyft.io
 /  ___/  |<   |  |\   __\\   __\	The Service Layer for Things
 \___ \|  |_\___  | |  |   |  |  	Licensed under the Apache License, Version 2.0
/____  >____/ ____| |__|   |__|
     \/     \/
`)
}

func showInfo(cmd *cli.Cmd) {
	cmd.Action = func() {
		showBanner()
		fmt.Printf(`
slyft, slyft.io is (C)opright 2017 Digital Incubation and Growth GmbH
info@slyft.io
`)
		fmt.Printf("Version %s", VERSION)
		fmt.Printf(`

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Please see CONTRIBUTORS https://github.com/thingforward/slyft-cli/raw/master/CONTRIBUTORS

Contains libraries from
* https://github.com/jawher/mow.cli		Copyright (c) 2014 Jawher Moussa
* https://github.com/mattn/go-runewidth		Copyright (c) 2016 Yasuhiro Matsumoto	
* https://github.com/op/go-logging		Copyright (c) 2013 Örjan Persson
* https://github.com/siddontang/go		Copyright (c) 2014 siddontang
* https://github.com/ghodss/yaml                Copyright (c) 2014 Sam Ghods
`)
	}
}

func main() {
	if len(os.Args) <= 1 {
		showBanner()
	}
	err := UpdateCheck(VERSION)
	if err != nil {
		Log.Error(err)
		return
	}

	app := cli.App("slyft", "")

	fDebug = app.BoolOpt("debug d", false, "Show debug output")

	app.Version("v version", VERSION)

	app.Command("user u", "User/Account management", RegisterUserRoutes)
	app.Command("project p", "Project management", RegisterProjectRoutes)
	app.Command("asset a", "Asset management", RegisterAssetRoutes)
	app.Command("info", "Show program info", showInfo)

	app.Run(os.Args)
}

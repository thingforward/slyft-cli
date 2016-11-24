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

	app.Commands = []cli.Command{
		{
			Name:     "register",
			Usage:    "register yourself to the backend",
			Category: "\n   Your Account",
			Action:   RegisterUser,
		},
		{
			Name:     "login",
			Aliases:  []string{"signin"},
			Usage:    "login into your account",
			Category: "\n   Your Account",
			Action:   LogUserIn,
		},
		{
			Name:     "logout",
			Aliases:  []string{"quit"},
			Usage:    "logout from your session",
			Category: "\n   Your Account",
			Action:   LogUserOut,
		},
		{
			Name:     "destroy",
			Aliases:  []string{"implode"},
			Usage:    "destroy your account",
			Category: "\n   Your Account",
			Action:   LogUserOut,
		},
		{
			Name:     "create",
			Usage:    "create a project or a build",
			Category: "\n   Bring to life",
			Subcommands: []cli.Command{
				{
					Name:   "build",
					Usage:  "create a build for a project",
					Action: DummyFunction,
				},
				{
					Name:   "project",
					Usage:  "create a new proejct",
					Action: DummyFunction,
				},
			},
		},
		{
			Name:     "add",
			Aliases:  []string{"a"},
			Usage:    "add assets or settings",
			Category: "\n   Extend the existing",
			Subcommands: []cli.Command{
				{
					Name:   "asset",
					Usage:  "add asset to a project",
					Action: DummyFunction,
				},
				{
					Name:   "settings",
					Usage:  "add/change settings of a project",
					Action: DummyFunction,
				},
			},
		},
		{
			Name:     "list",
			Aliases:  []string{"ls"},
			Usage:    "list entities",
			Category: "\n   Gather information",
			Subcommands: []cli.Command{
				{
					Name:    "projects",
					Aliases: []string{"p"},
					Usage:   "list projects",
					Action:  DummyFunction,
				},
				{
					Name:    "assets",
					Aliases: []string{"a"},
					Usage:   "list assets",
					Action:  DummyFunction,
				},
				{
					Name:    "builds",
					Aliases: []string{"b"},
					Usage:   "list assets",
					Action:  DummyFunction,
				},
			},
		},
		{
			Name:     "show",
			Aliases:  []string{"sh"},
			Usage:    "show entity",
			Category: "\n   Gather information",
			Subcommands: []cli.Command{
				{
					Name:    "project",
					Aliases: []string{"p"},
					Usage:   "show project details",
					Action:  DummyFunction,
				},
				{
					Name:    "asset",
					Aliases: []string{"a"},
					Usage:   "show asset details",
					Action:  DummyFunction,
				},
				{
					Name:    "build",
					Aliases: []string{"b"},
					Usage:   "show build details",
					Action:  DummyFunction,
				},
			},
		},
		{
			Name:     "run",
			Aliases:  []string{"r"},
			Usage:    "execute your wish",
			Category: "\n   Make things happen",
			Subcommands: []cli.Command{
				{
					Name:    "project",
					Aliases: []string{"p"},
					Usage:   "run a specific project (after creating a build for it)",
					Action:  DummyFunction,
				},
				{
					Name:    "build",
					Aliases: []string{"b"},
					Usage:   "run a specific build (after confirming its validity)",
					Action:  DummyFunction,
				},
			},
		},
		{
			Name:     "validate",
			Aliases:  []string{"v"},
			Usage:    "prepare to run",
			Category: "\n   Get, Set... and...",
			Subcommands: []cli.Command{
				{
					Name:    "project",
					Aliases: []string{"p"},
					Usage:   "validate a projects status",
					Action:  DummyFunction,
				},
				{
					Name:    "build",
					Aliases: []string{"b"},
					Usage:   "validate a build",
					Action:  DummyFunction,
				},
			},
		},
		{
			Name:     "delete",
			Aliases:  []string{"rm"},
			Usage:    "delete entity",
			Category: "\n   Delete into oblivion",
			Subcommands: []cli.Command{
				{
					Name:    "project",
					Aliases: []string{"p"},
					Usage:   "remove an existing project",
					Action:  DummyFunction,
				},
				{
					Name:    "build",
					Aliases: []string{"b"},
					Usage:   "remove a build configuration",
					Action:  DummyFunction,
				},
				{
					Name:    "asset",
					Aliases: []string{"a"},
					Usage:   "remove an existing asset",
					Action:  DummyFunction,
				},
			},
		},
	}

	app.Run(os.Args)
}

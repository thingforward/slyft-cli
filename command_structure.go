package main

import "github.com/urfave/cli"

var CommandStructure = []cli.Command{
	{
		Name:     "register",
		Usage:    "register to the Slyft service",
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
		Usage:    "destroy your account [think twice]",
		Category: "\n   Your Account",
		Action:   DeleteUser,
	},
	{
		Name:     "project",
		Aliases:  []string{"p"},
		Usage:    "Project Management",
		Category: "Operations",
		Subcommands: []cli.Command{
			{
				Name:   "create",
				Usage:  "create a new project",
				Action: ProjectCreationHandler,
			},
			{
				Name:   "list",
				Usage:  "list all your projects",
				Action: ProjectListHandler,
			},
			{
				Name:   "show",
				Usage:  "show details of your project",
				Action: ProjectShowHandler,
			},
			//{
			//Name:   "validate",
			//Usage:  "validate your project",
			//Action: ProjectValidationHandler,
			//},
			{
				Name:   "delete",
				Usage:  "create new",
				Action: ProjectDeletionHandler,
			},
		},
	},
	{
		Name:     "build",
		Aliases:  []string{"b"},
		Usage:    "Build Management",
		Category: "Operations",
		Subcommands: []cli.Command{
			{
				Name:   "add",
				Usage:  "add a build to an existing project",
				Action: DummyFunction,
			},
			{
				Name:   "list",
				Usage:  "list all the builds of a project",
				Action: DummyFunction,
			},
			{
				Name:   "show",
				Usage:  "show the details of a specific build",
				Action: DummyFunction,
			},
			{
				Name:   "validate",
				Usage:  "validate a specific build",
				Action: DummyFunction,
			},
			{
				Name:   "start",
				Usage:  "start a build",
				Action: DummyFunction,
			},
			{
				Name:   "delete",
				Usage:  "delete a specific build",
				Action: DummyFunction,
			},
		},
	},
	{
		Name:     "asset",
		Aliases:  []string{"a"},
		Usage:    "All that to do with the assets",
		Category: "Operations",
		Subcommands: []cli.Command{
			{
				Name:   "add",
				Usage:  "add a new asset to a project",
				Action: DummyFunction,
			},
			{
				Name:   "list",
				Usage:  "list the asssets",
				Action: DummyFunction,
			},
			{
				Name:   "show",
				Usage:  "show the details of an asset",
				Action: DummyFunction,
			},
			{
				Name:   "delete",
				Usage:  "remove an asset",
				Action: DummyFunction,
			},
		},
	},
	{
		Name:     "settings",
		Aliases:  []string{"s"},
		Usage:    "All that to do with the assets",
		Category: "Operations",
		Subcommands: []cli.Command{
			{
				Name:   "add",
				Usage:  "add new settings to a proejct",
				Action: DummyFunction,
			},
			{
				Name:   "list",
				Usage:  "list the settings of a project",
				Action: DummyFunction,
			},
			{
				Name:   "delete",
				Usage:  "remove specific settins of a project",
				Action: DummyFunction,
			},
		},
	},
	{
		Name:  "server",
		Usage: "Server checks and such",
		Subcommands: []cli.Command{
			{
				Name:   "ping",
				Usage:  "Ping the Slyft service",
				Action: PingServer,
			},
		},
	},
}

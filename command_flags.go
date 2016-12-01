package main

import "github.com/urfave/cli"

var CommandFlags = []cli.Flag{
	cli.StringFlag{
		Name:        "project, p",
		Value:       "",
		Usage:       "Get the `Project Name`",
		Destination: &ProjectName,
	},
	cli.BoolFlag{
		Name:        "all, a",
		Usage:       "get everything",
		Destination: &Everything,
	},
}

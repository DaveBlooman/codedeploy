package main

import (
	"fmt"
	"os"

	"github.com/DaveBlooman/codedeploy/command"
	"github.com/codegangsta/cli"
)

var GlobalFlags = []cli.Flag{}

var Commands = []cli.Command{

	{
		Name:   "deploy",
		Usage:  "",
		Action: command.CmdDeploy,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "region, r",
				Usage: "Region, e.g: eu-west-1",
			},
			cli.StringFlag{
				Name:  "bucket, b",
				Usage: "Bucket Name",
			},
			cli.StringFlag{
				Name:  "filename, f",
				Usage: "Filename, e.g: app.zip",
			},
			cli.StringFlag{
				Name:  "deployment-group, d",
				Usage: "deployment-group",
			},
			cli.StringFlag{
				Name:  "app-name, a",
				Usage: "app-name",
			},
			cli.StringFlag{
				Name:  "awsprofile, p",
				Usage: "awsprofile",
			},
		},
	},
}

func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}

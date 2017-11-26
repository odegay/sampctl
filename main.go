package main

import (
	"fmt"
	"os"

	"gopkg.in/urfave/cli.v1"

	"github.com/Southclaws/sampctl/download"
)

var version = "master"

func main() {
	app := cli.NewApp()

	app.Author = "Southclaws"
	app.Email = "southclaws@gmail.com"
	app.Name = "sampctl"
	app.Description = "A small utility for starting and managing SA:MP servers with better settings handling and crash resiliency."
	app.Version = version

	cli.VersionFlag = cli.BoolFlag{
		Name:  "appVersion, V",
		Usage: "sampctl version",
	}

	cacheDir, err := download.GetCacheDir()
	if err != nil {
		fmt.Println("Failed to retrieve cache directory path (attempted <user folder>/.samp) ", err)
		return
	}
	err = os.MkdirAll(cacheDir, 0665)
	if err != nil {
		fmt.Println("Failed to create cache directory at ", cacheDir, ": ", err)
		return
	}

	app.Commands = []cli.Command{
		{
			Name:        "server",
			Aliases:     []string{"s"},
			Usage:       "sampctl server <subcommand>",
			Description: "For managing servers and runtime configurations.",
			Subcommands: []cli.Command{
				{
					Name:        "init",
					Usage:       "sampctl server init",
					Description: "Bootstrap a new SA:MP server and generates a `samp.json` configuration based on user input. If `gamemodes`, `filterscripts` or `plugins` directories are present, you will be prompted to select relevant files.",
					Action:      serverInit,
					Flags:       serverInitFlags,
				},
				{
					Name:        "download",
					Usage:       "sampctl server download",
					Description: "Downloads the files necessary to run a SA:MP server to the current directory (unless `--dir` specified). Will download the latest stable (non RC) server version unless `--version` is specified.",
					Action:      serverDownload,
					Flags:       serverDownloadFlags,
				},
				{
					Name:        "run",
					Usage:       "sampctl server run",
					Description: "Generates a `server.cfg` file based on the configuration inside `samp.json` then executes the server process and automatically restarts it on crashes.",
					Action:      serverRun,
					Flags:       serverRunFlags,
				},
			},
		},
		{
			Name:        "package",
			Aliases:     []string{"p"},
			Usage:       "sampctl package <subcommand>",
			Description: "For managing Pawn packages such as gamemodes and libraries.",
			Subcommands: []cli.Command{
				{
					Name:        "ensure",
					Usage:       "sampctl package ensure",
					Description: "Ensures dependencies are up to date based on the `dependencies` field in `pawn.json`.",
					Action:      packageEnsure,
					Flags:       packageEnsureFlags,
				},
				{
					Name:        "build",
					Usage:       "sampctl package build",
					Description: "Builds a package defined by a `pawn.json` or `pawn.yaml` file.",
					Action:      packageBuild,
					Flags:       packageBuildFlags,
				},
				{
					Name:        "run",
					Usage:       "sampctl package run",
					Description: "Compiles and runs a package defined by a `pawn.json` or `pawn.yaml` file.",
					Action:      packageRun,
					Flags:       packageRunFlags,
				},
			},
		},
		{
			Name:        "version",
			Description: "Show version number - this is also the version of the container image that will be used for `--container` runtimes.",
			Action:      cli.VersionPrinter,
		},
		{
			Name:        "docs",
			Usage:       "sampctl docs > documentation.md",
			Description: "Generate documentation in markdown format and print to standard out.",
			Action: func(c *cli.Context) error {
				docs := GenerateDocs(c.App)
				fmt.Print(docs)
				return nil
			},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		fmt.Printf("Exited with error: %v\n", err)
	}
}

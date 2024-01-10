package plugin

import "github.com/urfave/cli/v2"

func CMD() *cli.Command {
	return &cli.Command{
		Name:  "plugin",
		Usage: "Create and manage Kite plugins",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path",
				Usage: "Path to the Kite plugin",
				Value: ".",
			},
		},
		Subcommands: []*cli.Command{
			initCMD(),
			buildCMD(),
			testCMD(),
		},
	}
}

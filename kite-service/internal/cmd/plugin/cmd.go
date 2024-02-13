package plugin

import "github.com/urfave/cli/v2"

func CMD() *cli.Command {
	return &cli.Command{
		Name:  "plugin",
		Usage: "Create and manage Kite plugins",
		Subcommands: []*cli.Command{
			initCMD(),
			buildCMD(),
			testCMD(),
			deployCMD(),
		},
	}
}

package main

import (
	"log"
	"os"

	"github.com/urfave/cli"

	"github.com/mokiat/softgfx/cmd/softgfx-lvlgen/internal/conversion"
)

func main() {
	app := cli.NewApp()
	app.Name = "softgfx-lvlgen"
	app.Usage = "generate levels for softgfx from wavefront obj files"
	app.UsageText = "softgfx-lvlgen [--in obj_file] [--out level_file]"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "in",
			Usage: "specify an obj file to read model from (by default STDIN is used)",
		},
		cli.StringFlag{
			Name:  "out",
			Usage: "specify a file to write json level to (by default STDOUT is used)",
		},
		cli.Float64Flag{
			Name:  "scale",
			Usage: "specify a scaling factor for the level",
			Value: 64.0,
		},
	}
	app.Version = "0.1.0"
	app.Action = conversion.Command()
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

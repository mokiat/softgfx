package conversion

import (
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v2"
)

func Command() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		var in io.Reader = os.Stdin
		if inFilePath := ctx.String("in"); inFilePath != "" {
			inFile, err := os.Open(inFilePath)
			if err != nil {
				return fmt.Errorf("failed ot open input file: %w", err)
			}
			defer inFile.Close()
			in = inFile
		}

		var out io.Writer = os.Stdout
		if outFilePath := ctx.String("out"); outFilePath != "" {
			outFile, err := os.Create(outFilePath)
			if err != nil {
				return fmt.Errorf("failed ot create output file: %w", err)
			}
			defer outFile.Close()
			out = outFile
		}

		scale := ctx.Float64("scale")

		return run(in, out, scale)
	}
}

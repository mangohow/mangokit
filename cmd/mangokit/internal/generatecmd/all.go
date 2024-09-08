package generatecmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var CmdGenAll = &cobra.Command{
	Use:     "all",
	Example: "mangokit generate proto api",
	Short:   "Generate proto, openapi and wire",
	Long:    "Generate proto, openapi and wire",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Fprintf(os.Stderr, "missing file path you wang to generate.")
			os.Exit(1)
		}
		dir := args[0]

		GenerateAll(dir)
	},
}

func GenerateAll(dir string) {
	if err := GenerateProtos(dir); err != nil {
		color.Red("generate proto failed")
	}

	if err := GenerateOpenAPI(dir); err != nil {
		color.Red("generate openapi failed")
	}

	if err := GenerateWire(); err != nil {
		color.Red("generate wire failed")
	}
}

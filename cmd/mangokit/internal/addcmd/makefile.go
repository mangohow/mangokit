package addcmd

import (
	_ "embed"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var CmdAddMakefile = &cobra.Command{
	Use:   "makefile",
	Short: "Generate makefile",
	Long:  "Generate makefile",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := os.Stat("makefile")
		if err == nil {
			color.Yellow("makefile is already exist!\n")
			return
		}

		if err = os.WriteFile("makefile", []byte(makefileContent), 0666); err != nil {
			color.Red("generate makefile error, %v\n", err)
			return
		}

		color.Green("makefile added!")
	},
}

//go:embed makefile.tpl
var makefileContent string

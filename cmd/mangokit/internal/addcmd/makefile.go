package addcmd

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var CmdAddMakefile = &cobra.Command{
	Use:   "makefile",
	Short: "Generate makefile",
	Long:  "Generate makefile",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := os.Stat("makefile")
		if err == nil {
			fmt.Printf("makefile is already exist!\n")
			return
		}

		if err = os.WriteFile("makefile", []byte(makefileContent), 0666); err != nil {
			fmt.Printf("generate makefile error, %v\n", err)
			return
		}
	},
}

//go:embed makefile.tpl
var makefileContent string

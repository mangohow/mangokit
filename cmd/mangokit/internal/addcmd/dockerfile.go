package addcmd

import (
	_ "embed"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var CmdAddDockerfile = &cobra.Command{
	Use:   "dockerfile",
	Short: "Generate dockerfile",
	Long:  "Generate dockerfile",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := os.Stat("Dockerfile")
		if err == nil {
			color.Yellow("Dockerfile is already exist!\n")
			return
		}

		if err = os.WriteFile("Dockerfile", []byte(dockerfileContent), 0666); err != nil {
			color.Red("generate Dockerfile error, %v\n", err)
			return
		}

		color.Green("Dockerfile added!")
	},
}

//go:embed dockerfile.tpl
var dockerfileContent string

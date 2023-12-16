package addcmd

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var CmdAddDockerfile = &cobra.Command{
	Use:   "dockerfile",
	Short: "Generate dockerfile",
	Long:  "Generate dockerfile",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := os.Stat("Dockerfile")
		if err == nil {
			fmt.Printf("Dockerfile is already exist!\n")
			return
		}

		if err = os.WriteFile("Dockerfile", []byte(dockerfileContent), 0666); err != nil {
			fmt.Printf("generate Dockerfile error, %v\n", err)
			return
		}
	},
}

//go:embed dockerfile.tpl
var dockerfileContent string

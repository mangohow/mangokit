package generatecmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var CmdGenWire = &cobra.Command{
	Use:   "wire",
	Short: "generate wire injection files",
	Long:  "generate wire injection files",
	Run: func(cmd *cobra.Command, args []string) {
		c := exec.Command("go", "mod", "tidy")
		if err := c.Run(); err != nil {
			fmt.Printf("go mod tidy, %v\n", err)
		}

		command := exec.Command("go", "generate", "./...")
		if err := command.Run(); err != nil {
			fmt.Printf("go generate error, %v\n", err)
		}
	},
}

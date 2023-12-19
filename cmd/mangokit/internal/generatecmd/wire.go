package generatecmd

import (
	"os/exec"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var CmdGenWire = &cobra.Command{
	Use:   "wire",
	Short: "generate wire injection files",
	Long:  "generate wire injection files",
	Run: func(cmd *cobra.Command, args []string) {
		GenerateWire()
	},
}

func GenerateWire()  error{
	c := exec.Command("go", "mod", "tidy")
	if err := c.Run(); err != nil {
		return err
		color.Red("go mod tidy, %v\n", err)
	}

	command := exec.Command("go", "generate", "./...")
	if err := command.Run(); err != nil {
		return err
		color.Red("go generate error, %v\n", err)
	}

	return nil
}
package addcmd

import "github.com/spf13/cobra"

var CmdAdd = &cobra.Command{
	Use:   "add",
	Short: "Add proto, dockerfile or makefile",
	Long:  "Add proto, dockerfile or makefile",
}

func init() {
	CmdAdd.AddCommand(CmdAddApi)
	CmdAdd.AddCommand(CmdAddError)
	CmdAdd.AddCommand(CmdAddDockerfile)
	CmdAdd.AddCommand(CmdAddMakefile)
}

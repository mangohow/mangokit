package generatecmd

import "github.com/spf13/cobra"

var CmdGenerate = &cobra.Command{
	Use:     "generate",
	Short:   "Generate files, such as go and openapi",
	Long:    "Generate files, include go files from proto files and wire inject and openapi files",
	Example: "mangokit generate [api, wire, openapi]",
}

func init() {
	CmdGenerate.AddCommand(CmdGenProto)
	CmdGenerate.AddCommand(CmdGenWire)
	CmdGenerate.AddCommand(CmdGenOpenApi)
	CmdGenerate.AddCommand(CmdGenAll)
}

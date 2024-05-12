package generatecmd

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var CmdGenOpenApi = &cobra.Command{
	Use:   "openapi",
	Short: "Generate openapi files",
	Long:  "Generate openapi files",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Fprintf(os.Stderr, "missing file path you wang to generate")
			os.Exit(1)
		}
		dir := args[0]

		GenerateOpenAPI(dir)
	},
}

func init() {
	CmdGenOpenApi.Flags().StringSliceVarP(&protoPath, "proto_path", "p", protoPath, "specify proto_path")
}

func GenerateOpenAPI(dir string) error {
	// 遍历目录, 获取所有proto文件
	protos := make([]string, 0)
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "generate openapi error, dir: %s, file %s, error: %v\n", dir, path, err)
			os.Exit(1)
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".proto") {
			protos = append(protos, path)
		}

		return nil
	})
	if err != nil {
		color.Red("walk protos error, %v\n", err)
		return err
	}

	args := []string{}
	for _, s := range protoPath {
		args = append(args, "--proto_path="+s)
	}
	args = append(args, "--openapi_out=fq_schema_naming=true,default_response=false:.")
	args = append(args, protos...)

	cmd := exec.Command("protoc", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err = cmd.Run(); err != nil {
		color.Red("generate openapi error, %v\n", err)
		return err
	}

	return nil
}

package generatecmd

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var CmdGenProto = &cobra.Command{
	Use:   "proto",
	Short: "Generate go files based on proto files",
	Long:  "Generate go files based on proto files",
	Run: func(cmd *cobra.Command, args []string) {
		dir := "api"
		if len(args) > 0 {
			dir = args[0]
		}

		GenerateProtos(dir)
	},
}

var (
	protoPath = []string{"third_party", "api"}
)

func init() {
	CmdGenProto.Flags().StringSliceVarP(&protoPath, "proto_path", "p", protoPath, "specify proto_path")
}

//  protoc --proto_path=third_party --proto_path=api --go_out=. --go-gin_out=. --go-error_out=. api/errors/v1/proto/errors.proto api/helloworld/v1/proto/greeter.proto

func GenerateProtos(dir string) error {
	// 遍历目录, 获取所有proto文件
	protos := make([]string, 0)
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
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
	args = append(args, "--go_out=.")
	args = append(args, "--go-gin_out=.")
	args = append(args, "--go-error_out=.")
	args = append(args, protos...)

	cmd := exec.Command("protoc", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err = cmd.Run(); err != nil {
		color.Red("generate proto files error, %v\n", err)
		return err
	}

	return nil
}

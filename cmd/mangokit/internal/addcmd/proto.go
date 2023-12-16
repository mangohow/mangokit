package addcmd

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var CmdAddApi = &cobra.Command{
	Use:     "api",
	Short:   "Add proto api file",
	Long:    "Add proto api file",
	Example: "mangokit add api api/helloworld/v1/proto hello.proto",
	Run: func(cmd *cobra.Command, args []string) {
		dir, name := cmdFunc(args)
		GenerateProtoFile(dir, name, ProtoApiContent)
	},
}

var CmdAddError = &cobra.Command{
	Use:     "error",
	Short:   "Add proto error file",
	Long:    "Add proto error file",
	Example: "mangokit add api api/helloworld/v1/proto errReason.proto",
	Run: func(cmd *cobra.Command, args []string) {
		dir, name := cmdFunc(args)
		GenerateProtoFile(dir, name, ProtoErrorContent)
	},
}

func cmdFunc(args []string) (protoDir, protoName string) {
	if len(args) < 2 {
		prompt1 := &survey.Input{
			Message: "Which folder do you want the proto file in?",
			Help:    "Enter the name of the folder where you want to put the proto file",
		}
		err := survey.AskOne(prompt1, &protoDir)
		if err != nil || protoDir == "" {
			return
		}

		prompt2 := &survey.Input{
			Message: "What is the proto file name?",
			Help:    "Enter the name of the proto file",
		}
		err = survey.AskOne(prompt2, &protoName)
		if err != nil || protoName == "" {
			return
		}

		if !strings.HasSuffix(protoName, ".proto") {
			protoName += ".proto"
		}
	} else {
		protoDir = args[0]
		protoName = args[1]
	}

	return protoDir, protoName
}

//go:embed proto-api.tpl
var ProtoApiContent string

//go:embed proto-error.tpl
var ProtoErrorContent string

type TemplateInfo struct {
	Package string
	Name    string
}

func (t *TemplateInfo) execute(content string) string {
	buf := new(bytes.Buffer)
	tmpl, err := template.New("http").Parse(strings.TrimSpace(content))
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(buf, t); err != nil {
		panic(err)
	}
	return strings.Trim(buf.String(), "\r\n")
}

func GenerateProtoFile(dir, name string, content string) error {
	dir = strings.TrimRight(dir, string(filepath.Separator))
	path := filepath.Join(dir, name)
	pkg := dir
	// 如果生成的proto文件在proto文件夹下，则选择将生成的proto文件放在上一级目录
	// 例如 api/hello/v1/proto   .proto文件放在proto文件夹下，而生成的.go文件放在v1文件夹下
	if strings.HasSuffix(pkg, "/proto") {
		pkg = strings.TrimRight(pkg, "/proto")
	}
	info := &TemplateInfo{
		Package: pkg,
		Name:    strings.TrimSuffix(name, filepath.Ext(name)), // 去掉文件后缀
	}
	if filepath.IsAbs(dir) {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Printf("failed to get current working dir, %v\n", err)
			return err
		}

		rel, err := filepath.Rel(wd, dir)
		if err != nil {
			fmt.Printf("failed to get relative path, %v\n", err)
			return err
		}
		info.Package = rel
		if strings.HasSuffix(info.Package, "/proto") {
			info.Package = strings.TrimRight(info.Package, "/proto")
		}
	}
	// 在windows下是\\，但是在proto文件中都为/
	info.Package = strings.ReplaceAll(info.Package, "\\", "/")

	cont := info.execute(content)
	if err := os.WriteFile(path, []byte(cont), 0666); err != nil {
		fmt.Printf("write file error, file: %s, reason: %v\n", name, err)
		return err
	}

	return nil
}

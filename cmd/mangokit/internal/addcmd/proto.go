package addcmd

import (
	"bytes"
	_ "embed"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var CmdAddApi = &cobra.Command{
	Use:     "api",
	Short:   "Add proto api file",
	Long:    "Add proto api file",
	Example: "mangokit add api api/helloworld hello.proto",
	Run: func(cmd *cobra.Command, args []string) {
		dir, name := cmdFunc(args)
		GenerateProtoFile(dir, name, ProtoApiContent)
	},
}

var CmdAddError = &cobra.Command{
	Use:     "error",
	Short:   "Add proto error file",
	Long:    "Add proto error file",
	Example: "mangokit add error api/errors errReason.proto",
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
	Package  string
	Name     string
	FileName string
	DirName  string
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

func GenerateProtoFile(dir, name, content string) error {
	dir = strings.TrimRight(dir, string(filepath.Separator))
	path := filepath.Join(dir, name)
	pkg := dir

	filename := strings.TrimSuffix(name, filepath.Ext(name)) // 去掉文件后缀
	info := &TemplateInfo{
		Package:  pkg,
		FileName: filename,
		Name:     case2Camel(filename),
		DirName:  filepath.Base(dir),
	}
	if filepath.IsAbs(dir) {
		wd, err := os.Getwd()
		if err != nil {
			color.Red("failed to get current working dir, %v\n", err)
			return err
		}

		rel, err := filepath.Rel(wd, dir)
		if err != nil {
			color.Red("failed to get relative path, %v\n", err)
			return err
		}
		info.Package = rel
	}
	// 在windows下是\\，但是在proto文件中都为/
	info.Package = strings.ReplaceAll(info.Package, "\\", "/")

	// 如果目录不存在，创建目录
	if _, err := os.Stat(dir); err != nil || os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0666); err != nil {
			color.Red("create directory %s failed, %v", dir, err)
			return err
		}
	}

	cont := info.execute(content)
	if err := os.WriteFile(path, []byte(cont), 0666); err != nil {
		color.Red("write file error, file: %s, reason: %v\n", name, err)
		return err
	}

	color.Green("%s added!", path)

	return nil
}

var enCases = cases.Title(language.AmericanEnglish, cases.NoLower)

func case2Camel(name string) string {
	if !strings.Contains(name, "_") {
		if name == strings.ToUpper(name) {
			name = strings.ToLower(name)
		}
		return enCases.String(name)
	}
	strs := strings.Split(name, "_")
	words := make([]string, 0, len(strs))
	for _, w := range strs {
		hasLower := false
		for _, r := range w {
			if unicode.IsLower(r) {
				hasLower = true
				break
			}
		}
		if !hasLower {
			w = strings.ToLower(w)
		}
		w = enCases.String(w)
		words = append(words, w)
	}

	return strings.Join(words, "")
}

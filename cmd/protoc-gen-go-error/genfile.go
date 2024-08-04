package main

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/mangohow/mangokit/errors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

func generateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	if len(file.Enums) == 0 {
		return nil
	}
	filename := file.GeneratedFilenamePrefix + "_errors.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// Code generated by protoc-gen-go-error. DO NOT EDIT.")
	g.P("// versions:")
	g.P(fmt.Sprintf("// - protoc-gen-go-error %s", version))
	g.P("// - protoc              ", protocVersion(gen))
	if file.Proto.GetOptions().GetDeprecated() {
		g.P("// ", file.Desc.Path(), " is a deprecated file.")
	} else {
		g.P("// source: ", file.Desc.Path())
	}
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()

	// gen import
	g.P("import (")
	g.P(`"fmt"`)
	g.P()
	g.P(`"github.com/mangohow/mangokit/errors"`)
	g.P(")")

	generateFileContent(gen, file, g)

	return g
}

func generateFileContent(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	if len(file.Enums) == 0 {
		return
	}

	index := 0
	for _, enum := range file.Enums {
		if !genErrorsReason(gen, file, g, enum) {
			index++
		}
	}
	// If all enums do not contain 'mangokit.code', the current file is skipped
	if index == 0 {
		g.Skip()
	}
}

func genErrorsReason(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, enum *protogen.Enum) bool {
	defaultCode := proto.GetExtension(enum.Desc.Options(), errors.E_DefaultCode)

	code := 0
	if ok := defaultCode.(int32); ok != 0 {
		code = int(ok)
	}
	if code > 600 || code < 0 {
		panic(fmt.Sprintf("Enum '%s' range must be greater than 0 and less than or equal to 600", string(enum.Desc.Name())))
	}

	var ees EnumErrors
	for _, value := range enum.Values {
		d := proto.GetExtension(value.Desc.Options(), errors.E_Desc)
		desc, _ := d.(string)
		if desc != "" {
			ees.GenDesc = true
		}

		status := code
		eCode := proto.GetExtension(value.Desc.Options(), errors.E_Code)
		if ok := eCode.(int32); ok != 0 {
			status = int(ok)
		}
		// If the current enumeration does not contain 'mangokit.code'
		// or the code value exceeds the range, the current enum will be skipped
		if status > 600 || status < 0 {
			panic(fmt.Sprintf("Enum '%s' range must be greater than 0 and less than or equal to 600", string(value.Desc.Name())))
		}

		if status == 0 {
			continue
		}

		// 注释
		comment := value.Comments.Leading.String()
		if comment == "" {
			comment = value.Comments.Trailing.String()
		}

		e := &ErrorDesc{
			Comment:    comment,
			CamelName:  case2Camel(string(value.Desc.Name())),
			Name:       string(value.Desc.Name()),
			HTTPStatus: status,
			EnumName:   case2Camel(string(enum.Desc.Name())),
			Desc:       desc,
		}

		ees.Errors = append(ees.Errors, e)
	}

	if len(ees.Errors) == 0 {
		return true
	}

	g.P(ees.execute())

	return false
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

func protocVersion(gen *protogen.Plugin) string {
	v := gen.Request.GetCompilerVersion()
	if v == nil {
		return "(unknown)"
	}
	var suffix string
	if s := v.GetSuffix(); s != "" {
		suffix = "-" + s
	}
	return fmt.Sprintf("v%d.%d.%d%s", v.GetMajor(), v.GetMinor(), v.GetPatch(), suffix)
}

package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"text/template"
)

//go:embed gin-template.tpl
var TextTemplate string

type ServiceDesc struct {
	ServiceName      string
	LowerServiceName string
	Comment          string
	Methods          []*MethodDesc
	ImportSerialize  bool
}

type MethodDesc struct {
	Name           string // 方法名
	Request        string // 请求参数名
	Reply          string // 响应参数名
	ServiceName    string // 所属service名
	Comment        string // 注释
	InputFieldLen  int    // 输入参数字段数量
	OutputFieldLen int    // 输出参数字段数量

	// http rule
	Path   string // 请求路径
	Method string // 请求方法

	LowerServiceName string // 小写service名
	EncodeParam      bool
	EncodeForm       bool
}

func (s *ServiceDesc) execute() string {
	buf := new(bytes.Buffer)
	tmpl, err := template.New("http").Parse(strings.TrimSpace(TextTemplate))
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse template error: %v\n", err)
		os.Exit(1)
	}
	if err := tmpl.Execute(buf, s); err != nil {
		fmt.Fprintf(os.Stderr, "execute template error: %v\n", err)
		os.Exit(1)
	}
	return strings.Trim(buf.String(), "\r\n")
}

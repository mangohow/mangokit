package main

import (
	"bytes"
	_ "embed"
	"text/template"
)

//go:embed err-template.tpl
var ErrorTemplate string

type ErrorDesc struct {
	Comment    string // 注释
	CamelName  string // 大驼峰名称
	Name       string // 名称
	HTTPStatus int    // http响应码
	EnumName   string // 枚举名称
}

type EnumErrors struct {
	Errors []*ErrorDesc
}

func (e EnumErrors) execute() string {
	buf := new(bytes.Buffer)
	tmpl, err := template.New("errors").Parse(ErrorTemplate)
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(buf, e); err != nil {
		panic(err)
	}
	return buf.String()
}

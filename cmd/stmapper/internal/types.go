package internal

import (
	"go/ast"
	"reflect"
)

type Package struct {
	Name  string
	Files []*File
}

type File struct {
	// 除了要生成代码的声明之外的其他声明
	OtherDecls []ast.Decl
	Comments   []*ast.CommentGroup
	AbsPath    string
	Imports    [][2]string // file imports
	Package    string      // package name
	Funcs      []*Func     // funcs to generatePackage
	Errors     []error     // errors when parse file
}

type Func struct {
	Name     string // func name
	Comments *ast.CommentGroup
	Inputs   []*Param // input parameter
	Outputs  []*Param // output parameter
}

type Param struct {
	Name          string            // parameter name, may be empty
	TypeName      string            // parameter type name
	MappingKey    string            // field mapping is performed based on which key
	Tag           string            // tag if kind is struct or ""
	Kind          reflect.Kind      // parameter kind
	Package       string            // parameter package name
	AbsPackage    string            // absolute package name
	IsPointer     bool              // is pointer or not
	IsReturnParam bool              // is func return parameter
	IsSlice       bool              // is slice
	Fields        map[string]*Param // struct fields or nil if is not struct
	PrimitiveType *Param
}

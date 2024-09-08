package main

import (
	"flag"
	"fmt"
	"github.com/mangohow/mangokit/cmd/stmapper/internal"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"strings"
)

var (
	tag      string
	mode     string
	copyMode string
)

func init() {
	flag.StringVar(&tag, "t", "stmapper", "specifies the tag on which the struct copies the field")
	flag.StringVar(&mode, "m", "name", "set the basis for copying struct fields, based on the name or tag")
	flag.StringVar(&copyMode, "d", "shallow", "shallow or deep copies, deep copies only support basic types such as asintxx、uintxx、floatxx、string")
	flag.Parse()
}

func main() {
	generator := internal.NewGenerator(internal.Config{
		Mode:     mode,
		Tag:      tag,
		CopyMode: copyMode,
	})
	generator.Execute()
	//dir, err := os.Getwd()
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(dir)
	//fst := token.NewFileSet()
	//file, err := parser.ParseFile(fst, "cmd/stmapper/stmapper/example.go", nil, 0)
	//if err != nil {
	//	panic(err)
	//}
	//ast.Print(fst, file)
	//f2, err := parser.ParseFile(fst, "cmd/stmapper/types/types.go", nil, 0)
	//if err != nil {
	//	panic(err)
	//}
	//ast.Print(fst, f2)
	//test()
	//reflect.Struct.String()
	//strconv.ParseInt("", 10, 32)
	//test()
}

func test() {
	// 解析 Go 文件
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "example.go", nil, 0)
	if err != nil {
		panic(err)
	}

	// 遍历函数声明
	ast.Inspect(f, func(node ast.Node) bool {
		if funDecl, ok := node.(*ast.FuncDecl); ok {
			for _, param := range funDecl.Type.Params.List {
				pkgName, typeName := getPackageAndTypeName(param.Type)
				if pkgName != "" {
					// 找到导入声明
					importDecl := findImportDecl(f, pkgName)
					if importDecl != nil {
						// 加载包,并在包 AST 中查找类型声明
						pkg, err := build.Import(strings.Trim(importDecl.Path.Value, "\""), "", 0)
						if err == nil {
							typeDecl := findTypeDecl(pkg, typeName)
							if typeDecl != nil {
								fmt.Printf("Found type declaration for %s.%s\n", pkgName, typeName)
								ast.Print(fset, typeDecl)
							}
						}
					}
				}
			}
		}
		return true
	})
}

func getPackageAndTypeName(expr ast.Expr) (string, string) {
	switch t := expr.(type) {
	case *ast.Ident:
		return "", t.Name
	case *ast.SelectorExpr:
		if pkgIdent, ok := t.X.(*ast.Ident); ok {
			return pkgIdent.Name, t.Sel.Name
		}
	case *ast.StarExpr:
		return getPackageAndTypeName(t.X)
	}
	return "", ""
}

func findImportDecl(f *ast.File, pkgName string) *ast.ImportSpec {
	for _, decl := range f.Imports {
		if decl.Name != nil && decl.Name.Name == pkgName {
			return decl
		}
		if strings.HasSuffix(strings.TrimSuffix(decl.Path.Value, `"`), "/"+pkgName) {
			return decl
		}
	}
	return nil
}

func findTypeDecl(pkg *build.Package, typeName string) *ast.TypeSpec {
	fset := token.NewFileSet()
	pkgFiles, err := parser.ParseDir(fset, pkg.Dir, nil, 0)
	if err != nil {
		return nil
	}

	for _, f := range pkgFiles[pkg.Name].Files {
		for _, decl := range f.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
				for _, spec := range genDecl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok && typeSpec.Name.Name == typeName {
						return typeSpec
					}
				}
			}
		}
	}

	return nil
}

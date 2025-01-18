package parser

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"slices"
	"strings"
)

type TypeManager struct {
	fst   *token.FileSet
	types map[string]map[string]*ast.TypeSpec
}

func NewTypeManager(fst *token.FileSet) *TypeManager {
	return &TypeManager{
		fst:   fst,
		types: make(map[string]map[string]*ast.TypeSpec),
	}
}

// GetTypeSpec 获取类型定义的ast
// Param:
//
//		   pkgName 包名，如果不是绝对路径或空，则需要传入pkgs
//	    typeName 要获取的类型名
//	    pkgs [2]string{绝对路径、别名} 比如 import model "github.com/xxx/yyy"
//
// Return:
//
//	ast.TypeSpec ast定义
//	string 绝对包名
func (tm *TypeManager) GetTypeSpec(pkgName, typeName string, pkgs [][2]string) (*ast.TypeSpec, string, error) {
	// 如果不是绝对路径并且pkgs长度大于0，则从中寻找绝对路径
	if !isAbsPkgName(pkgName) && len(pkgs) > 0 {
		// 在其他包里寻找，需要绝对包路径
		idx := slices.IndexFunc(pkgs, func(val [2]string) bool {
			if val[1] == pkgName {
				return true
			}

			if strings.HasSuffix(val[0], pkgName) {
				return true
			}
			return false
		})
		if idx != -1 {
			pkgName = pkgs[idx][0]
		}
	}

	ts, ok := tm.types[pkgName]
	if !ok {
		pkg, err := build.Import(pkgName, "", 0)
		if err != nil {
			return nil, "", fmt.Errorf("load package %s error: %v", pkgName, err)
		}
		if err = tm.LoadPackageFromDir(pkg.Dir, pkgName); err != nil {
			return nil, "", fmt.Errorf("load dir %s error: %v", pkg.Dir, err)
		}
	}
	ts, ok = tm.types[pkgName]
	if !ok {
		return nil, "", fmt.Errorf("can't find type %s from pkg %s", typeName, pkgName)
	}
	typ, ok := ts[typeName]
	if !ok {
		return nil, "", fmt.Errorf("type not found")
	}

	return typ, pkgName, nil
}

func (tm *TypeManager) LoadPackageFromDir(dir, pkgName string) error {
	pkgFiles, err := parser.ParseDir(tm.fst, dir, nil, 0)
	if err != nil {
		return fmt.Errorf("parse dir %s error: %v", dir, err)
	}

	spn := shortPackageName(pkgName)
	if _, ok := pkgFiles[spn]; !ok {
		return fmt.Errorf("package %s not found", pkgName)
	}
	return tm.LoadPackage(pkgFiles[spn], pkgName)
}

func (tm *TypeManager) LoadPackage(pkg *ast.Package, absPkgName string) error {
	var (
		m  map[string]*ast.TypeSpec
		ok bool
	)
	if m, ok = tm.types[absPkgName]; !ok {
		m = make(map[string]*ast.TypeSpec)
		tm.types[absPkgName] = m
	}
	for _, f := range pkg.Files {
		for _, decl := range f.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				m[typeSpec.Name.Name] = typeSpec
			}
		}
	}

	return nil
}

func shortPackageName(pkgName string) string {
	if !strings.Contains(pkgName, "/") {
		return pkgName
	}
	idx := strings.LastIndex(pkgName, "/")

	return pkgName[idx+1:]
}

func isAbsPkgName(pkgName string) bool {
	return strings.Contains(pkgName, "/")
}

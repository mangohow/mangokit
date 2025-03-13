package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"slices"
	"strings"

	"github.com/mangohow/mangokit/cmd/vulcan/internal/log"
	"github.com/mangohow/mangokit/cmd/vulcan/internal/parser"
	"github.com/mangohow/mangokit/cmd/vulcan/internal/types"
	"github.com/mangohow/mangokit/tools/collection"
)

type Config struct {
	// 根据结构体名称或者tag来映射字段
	Mode string
	// 根据哪个tag来映射字段
	Tag string

	dir string
	// 浅拷贝还是深拷贝
	CopyMode string
}

type Generator struct {
	cfg Config

	fst    *token.FileSet
	parser *parser.AstParser
}

func NewGenerator(cfg Config) *Generator {
	var mkf types.MappingKeyFunc
	if cfg.Mode == "name" {
		mkf = types.NameMappingKeyFunc
	} else if cfg.Mode == "tag" {
		mkf = func(name, tag string) string {
			return types.TagMappingKeyFunc(name, tag, cfg.Tag)
		}
	}

	fst := token.NewFileSet()
	return &Generator{
		cfg:    cfg,
		fst:    fst,
		parser: parser.NewAstParser(mkf, fst),
	}
}

func (g *Generator) Execute() error {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get current working directory: %v", err)
	}

	pkgs, err := g.parser.ParseDir(wd)
	if err != nil {
		log.Fatalf("failed to parse directory: %v", err)
	}

	for _, pkg := range pkgs {
		if err := g.generatePackage(pkg); err != nil {
			log.Fatalf("failed to generate package: %v", err)
		}
	}

	return nil
}

func (g *Generator) generatePackage(pkg *types.Package) error {
	for _, file := range pkg.Files {
		if err := g.generateFile(file); err != nil {
			log.Errorf("failed to generate file: %v", err)
			return err
		}
	}

	return nil
}

func (g *Generator) generateFile(file *types.File) error {
	// 遍历生成每个func的ast
	for _, fn := range file.Funcs {
		astFn, err := g.GenerateFuncAst(fn)
		if err != nil {
			return err
		}

		// 插入文件的声明中，不会打乱原有的文件中的声明
		idx := slices.Index(file.OtherDecls, nil)
		if idx == -1 {
			file.OtherDecls = append(file.OtherDecls, astFn)
		} else {
			file.OtherDecls[idx] = astFn
		}
	}

	// 生成代码
	imports := make([]*ast.ImportSpec, 0, len(file.Imports))
	for _, v := range file.Imports {
		i := &ast.ImportSpec{
			Name: ast.NewIdent(v[1]),
			Path: &ast.BasicLit{
				Value: v[0],
			},
		}
		imports = append(imports, i)
	}

	astFile := &ast.File{
		Decls:    file.OtherDecls,
		Imports:  imports,
		Comments: file.Comments,
	}

	if err := g.writeSourceCode(astFile, file); err != nil {
		return err
	}

	return nil
}

func (g *Generator) writeSourceCode(astSource ast.Node, file *types.File) error {
	buf := &bytes.Buffer{}
	err := format.Node(buf, g.fst, astSource)
	if err != nil {
		log.Fatalf("generate source code error, err=%v file=%s", err, file.Name)
		return err
	}

	// 写入文件
	i := strings.Index(file.Name, ".")
	filename := file.Name[:i] + "_gen" + file.Name[i:]
	err = os.WriteFile(filename, buf.Bytes(), 0666)
	if err != nil {
		log.Fatalf("write source code to %s error, err=%v", filename, err)
		return err
	}

	return nil
}

func (g *Generator) GenerateFuncAst(fn *types.Func) (*ast.FuncDecl, error) {
	astFunc := &ast.FuncDecl{
		Doc:  fn.Comments,
		Recv: fn.AstReceiver,
		Name: ast.NewIdent(fn.Name),
		Type: fn.AstFuncType,
	}
	body := &ast.BlockStmt{}
	// 生成body
	var stmtList []ast.Stmt
	for i := range fn.Outputs {
		stmt, err := g.GenerateBodyStmtForOutput(fn.Outputs[i], fn)
		if err != nil {
			return nil, err
		}
		stmtList = append(stmtList, stmt...)
	}

	body.List = stmtList

	return astFunc, nil
}

func (g *Generator) GenerateBodyStmtForOutput(param *types.Param, fn *types.Func) ([]ast.Stmt, error) {
	// 生成结构体键值对赋值的方式
	// t := &x{
	//     a: y.a,
	//     b: y.b
	//   }
	if param.IsReturnParam {
		return g.generateKVAssign(param, fn)
	}

	// 生成等于号赋值的方式
	// x.a = y.a
	// x.b = y.b
	return g.generateEqualAssign(param.Name, param, fn)
}

func (g *Generator) generateKVAssign(param *types.Param, fn *types.Func) ([]ast.Stmt, error) {
	// TODO
}

func (g *Generator) generateEqualAssign(name string, param *types.Param, fn *types.Func) (res []ast.Stmt, e error) {
	// 遍历每个字段生成相应的ast
	collection.ForEach(param.FieldNames, func(key string) bool {
		var (
			stmt ast.Stmt
			err  error
		)

		field := param.Fields[key]
		// 去input中寻找
		input := g.findInputForEach(key, fn.Inputs)
		if len(input) == 0 {
			// 继续下一个字段
			return true
		}

		// 根据类型进行赋值
		switch {
		// 基础类型赋值
		case types.IsBasicKind(field.Kind):
			stmt, err = g.basicKindAssign([]*types.Param{field}, input)
			//TODO
		}

		if err != nil {
			res, e = nil, err
			return false
		}
	})
}

// 构建基础类型赋值ast表达式
func (g *Generator) basicKindAssign(field []*types.Param, input []*types.Param) ([]ast.Stmt, error) {
	left := field[len(field)-1]
	right := input[len(input)-1]

	// 处理类型转换
	if left.Kind != right.Kind {
		return g.basicKindConvertAssign(field, input)
	}

	lname := types.MapNames(field)
	rname := types.MapNames(input)

	switch {
	// 浅拷贝
	case left.IsPointer == right.IsPointer:
		return []ast.Stmt{BuildAssignStmt([][]string{lname}, rname)}, nil
	case left.IsPointer: // 左侧为指针，右侧非指针 left.a = &right.a
		return []ast.Stmt{BuildAssignStmt1([]ast.Expr{BuildAstSelectorExpr(lname)}, []ast.Expr{BuildUnaryExpr("&", BuildAstSelectorExpr(rname))})}, nil
	case !left.IsPointer: // 左侧非指针，右侧指针   left.a = *right.a
		return []ast.Stmt{BuildAssignStmt1([]ast.Expr{BuildAstSelectorExpr(lname)}, []ast.Expr{BuildUnaryExpr("*", BuildAstSelectorExpr(rname))})}, nil
	}

	return nil, fmt.Errorf("cannot assign %s to %s", lname, rname)
}

// 基础类型转换 TODO
func (g *Generator) basicKindConvertAssign(field []*types.Param, input []*types.Param) ([]ast.Stmt, error) {
	left := field[len(field)-1]
	right := input[len(input)-1]

	// 处理类型转换
	if left.Kind != right.Kind {
		return g.basicKindConvertAssign(field, input)
	}

	lname := types.MapNames(field)
	rname := types.MapNames(input)
	// 类型断言
	if left.Kind == types.Interface {
		return []ast.Stmt{BuildAssignStmt1([]ast.Expr{BuildAstSelectorExpr(lname)}, []ast.Expr{BuildTypeAssertExpr(BuildAstSelectorExpr(rname), left.TypeName, left.IsPointer)})}, nil
	}

	// 获取类型转换器
	converter, err := getTypeConverter(left.Kind, right.Kind)
	if err != nil {
		return nil, err
	}
}

// 获取原始类型
func (g *Generator) primitiveType(param *types.Param) (types.Kind, *types.Param) {
	p := param
	for ; p.PrimitiveType != nil; p = p.PrimitiveType {
	}
	return p.Kind, p
}

func (g *Generator) findInputForEach(key string, inputs []*types.Param) (res []*types.Param) {
	collection.ForEach(inputs, func(v *types.Param) bool {
		if res = g.findInput(key, v); res != nil {
			return false
		}

		return true
	})

	return
}

func (g *Generator) findInput(key string, param *types.Param) []*types.Param {
	// 递归查找,chain保存了父节点一直到目标节点
	chain := []*types.Param{param}
	found := false
	var findInputRecursive func(key string, input *types.Param, chain []*types.Param)
	findInputRecursive = func(key string, input *types.Param, chain []*types.Param) {
		// 如果能直接找到
		if v, ok := input.Fields[key]; ok {
			chain = append(chain, v)
			found = true
			return
		}

		// 可能存在于结构体的子结构体中
		collection.ForEach(input.FieldNames, func(name string) bool {
			field := input.Fields[name]
			if field.Kind != types.Struct && field.Kind != types.Invalid {
				return true
			}

			findInputRecursive(key, field, chain)

			return !found
		})

	}

	findInputRecursive(key, param, chain)
	if found {
		return chain
	}

	return nil
}

package internal

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"strings"
)

var UnimportantError = errors.New("unimportant")

// 变量类型比如 User或model.User
type fieldType struct {
	star     bool   // 是否指针类型,如果为切片，则表示的是切片中元素是否是指针
	name     string // 参数名称
	typeName string // 参数类型名称
	pkg      string // 参数类型包名
	absPkg   string // 包绝对路径
	isSlice  bool   // 是否为切片类型
	astType  *ast.TypeSpec
}

type AstParser struct {
	fst              *token.FileSet
	typeDeclarations *TypeManager
	mkf              MappingKeyFunc
}

func NewAstParser(mkf MappingKeyFunc) *AstParser {
	ap := &AstParser{
		fst: token.NewFileSet(),
		mkf: mkf,
	}
	ap.typeDeclarations = NewTypeManager(ap.fst)

	return ap
}

// ParseDir 解析目录下的go代码
// 目录下可能还有目录，因此可能存在多个package
func (p *AstParser) ParseDir(dir string) ([]*Package, error) {
	pkgs, err := parser.ParseDir(p.fst, dir, func(info fs.FileInfo) bool {
		// 所有生成的文件以xxx_gen.go结尾或者测试文件，忽略这些文件
		if strings.HasSuffix(info.Name(), "_gen.go") ||
			strings.HasSuffix(info.Name(), "_test.go") {
			return false
		}
		return true
	}, 0)
	if err != nil {
		return nil, err
	}

	res := make([]*Package, 0, len(pkgs))
	for name, pkg := range pkgs {
		Debugf("pkg: %+v", name)
		pp := NewPackageParser(pkg, p.fst, p.typeDeclarations, p.mkf)
		pk, err := pp.Parse()
		if err != nil {
			return nil, err
		}
		res = append(res, pk)
	}

	return res, nil
}

type PackageParser struct {
	pkgName string
	astPkg  *ast.Package
	fst     *token.FileSet
	// 共享的，用于存放其他包的类型声明
	sharedTypeManager *TypeManager
	// 用于存放本包的类型声明
	typeManager *TypeManager
	mkf         MappingKeyFunc
}

func NewPackageParser(astPkg *ast.Package, fst *token.FileSet, sharedTypeManager *TypeManager, mkf MappingKeyFunc) *PackageParser {
	return &PackageParser{
		pkgName:           astPkg.Name,
		astPkg:            astPkg,
		fst:               fst,
		sharedTypeManager: sharedTypeManager,
		typeManager:       NewTypeManager(fst),
		mkf:               mkf,
	}
}

func (p *PackageParser) Parse() (*Package, error) {
	_ = p.typeManager.LoadPackage(p.astPkg, "")

	pkg := &Package{
		Name: p.pkgName,
	}
	for path, file := range p.astPkg.Files {
		fi, err := p.parseFile(path, file)
		if err != nil {
			return nil, err
		}
		pkg.Files = append(pkg.Files, fi)
	}

	return pkg, nil
}

// 解析单个文件
func (p *PackageParser) parseFile(absPath string, file *ast.File) (*File, error) {
	if !p.check(file) {
		return nil, nil
	}

	fileDesc := &File{
		Comments: file.Comments,
		AbsPath:  absPath,
		Package:  p.pkgName,
	}
	for _, imp := range file.Imports {
		i := [2]string{strings.Trim(imp.Path.Value, `"`)}
		if imp.Name != nil {
			i[1] = imp.Name.Name
		}
		fileDesc.Imports = append(fileDesc.Imports, i)
	}

	// 遍历文件进行解析
	for _, decl := range file.Decls {
		switch info := decl.(type) {
		case *ast.GenDecl:
			fileDesc.OtherDecls = append(fileDesc.OtherDecls, decl)
		case *ast.FuncDecl:
			fn, err := p.parseFunc(file, info, fileDesc)
			if err != nil && !errors.Is(err, UnimportantError) {
				fileDesc.Errors = append(fileDesc.Errors, err)
			}
			// 如果fn为nil，则该函数不需要生成
			if fn == nil {
				fileDesc.OtherDecls = append(fileDesc.OtherDecls, decl)
			} else {
				fileDesc.Funcs = append(fileDesc.Funcs, fn)
			}
		}
	}

	return fileDesc, nil
}

func (p *PackageParser) parseFunc(file *ast.File, decl *ast.FuncDecl, fileDesc *File) (fn *Func, er error) {
	fn = &Func{
		Name:     decl.Name.Name,
		Comments: decl.Doc,
	}

	if decl.Recv != nil {
		// TODO 后续实现为结构体绑定转换方法
	}

	// 获取当前函数的输入参数
	inputs, err := p.parseParameter(file, decl, decl.Type.Params, true)
	if err != nil {
		return nil, err
	}
	// 获取当前函数的输出参数
	outputs, err := p.parseParameter(file, decl, decl.Type.Results, false)
	if err != nil {
		return nil, err
	}

	// 根据输入输出参数以及函数标记获取要进行映射拷贝的结构体或切片类型
	fnInfo, err := p.parseFuncLabel(file, decl, inputs, outputs)
	if err != nil {
		return nil, err
	}

	// 获取所有输入输出参数的类型信息
	return p.buildFunc(fnInfo, file, decl, fileDesc)
}

func (p *PackageParser) buildFunc(fnInfo *funcInfo, file *ast.File, fn *ast.FuncDecl, fileDesc *File) (*Func, error) {
	// 寻找定义
	for _, input := range fnInfo.input {
		tp, absPkg, err := p.findDeclaration(input, fileDesc)
		if err != nil {
			return nil, fmt.Errorf("failed to find declaration of %s: %w", input.typeName, err)
		}
		input.astType = tp
		input.absPkg = absPkg
	}
	for _, output := range fnInfo.output {
		tp, absPkg, err := p.findDeclaration(output, fileDesc)
		if err != nil {
			return nil, fmt.Errorf("failed to find declaration of %s: %w", output.typeName, err)
		}
		output.astType = tp
		output.absPkg = absPkg
	}

	res, err := NewFuncDescBuilder(
		p.sharedTypeManager,
		p.typeManager,
		p.mkf,
		fileDesc.Imports,
		fnInfo,
	).Build(fn.Name.Name)
	if err != nil {
		return nil, fmt.Errorf("%s, in file: %s, func: %s", err.Error(), file.Name.Name, fn.Name.Name)
	}

	return res, nil
}

// 查找类型定义
func (p *PackageParser) findDeclaration(ft *fieldType, fileDesc *File) (*ast.TypeSpec, string, error) {
	// 在当前包里寻找
	pkgName := ft.pkg
	if pkgName == "" {
		return p.typeManager.GetTypeSpec(pkgName, ft.typeName, nil)
	}

	// 在其他包里寻找
	return p.sharedTypeManager.GetTypeSpec(pkgName, ft.typeName, fileDesc.Imports)
}

// 解析函数标记，并获取参数信息
func (p *PackageParser) parseFuncLabel(file *ast.File, decl *ast.FuncDecl, inputs, outputs []*fieldType) (fl *funcInfo, e error) {
	args := make([]string, 0, len(inputs))
	found := false
	fl = &funcInfo{}
	// 获取函数标记
	ast.Inspect(decl.Body, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		// 如果发生错误，则停止遍历
		if e != nil {
			return false
		}
		// 如果已经找到，则不需要再找
		if found {
			return false
		}

		ce, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		se, ok := ce.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		x, ok := fl.Match(se.Sel.Name)
		if !ok {
			return true
		}
		sel, ok := se.X.(*ast.Ident)
		if !ok || sel.Name != x {
			return true
		}

		fl.name = FuncLabel(se.Sel.Name)
		found = true
		for _, arg := range ce.Args {
			// 函数中可能还有函数调用，这是非法的
			name, ok := arg.(*ast.Ident)
			if !ok {
				e = fmt.Errorf("func label parameter imvalid, in file %s, func %s", file.Name.Name, decl.Name.Name)
				return false
			}

			args = append(args, name.Name)
		}

		return false
	})

	// 没有找到函数标记
	if fl.name == "" {
		return nil, UnimportantError
	}

	e = fl.Build(args, inputs, outputs, file.Name.Name, decl.Name.Name)
	if e != nil {
		return nil, e
	}

	return
}

// 解析输入或输出参数
func (p *PackageParser) parseParameter(file *ast.File, fn *ast.FuncDecl, params *ast.FieldList, isInputParam bool) ([]*fieldType, error) {
	if params == nil {
		return nil, nil
	}

	var res []*fieldType
	for _, param := range params.List {
		// 下面的参数写法会导致多个名称
		// func AToB(a, b, c string)
		// 此时param.Names中有三个类型相同但名称不同的参数
		ips := make([]string, 0, len(param.Names))
		for _, name := range param.Names {
			ips = append(ips, name.Name)
		}

		// 如果是输入参数，且该参数没有名称，就无法使用
		// 比如 func AToB(UserProto) User
		if isInputParam && len(param.Names) == 0 {
			return nil, fmt.Errorf("input parameter has no name, in file %s, func %s", file.Name.Name, fn.Name.Name)
		}

		// 如果参数没有名称，则添加空字符串
		if len(ips) == 0 {
			ips = append(ips, "")
		}

		t := param.Type
		var ft fieldType

	loop: // 之所以用for循环，是因为可能为指针类型，则需要获取指针指向的类型，
		for {
			switch typ := t.(type) {
			case *ast.StarExpr: // 指针类型
				// 为多级指针，不支持
				if ft.star {
					return nil, NewUnsupportedTypeConversionError(file.Name.Name, fn.Name.Name, ips[0])
				}
				ft.star = true
				t = typ.X
			case *ast.Ident: // 普通类型
				ft.typeName = typ.Name
				break loop
			case *ast.SelectorExpr: // x.y类型
				ft.typeName = typ.Sel.Name
				ft.pkg = typ.X.(*ast.Ident).Name
				break loop
			case *ast.ArrayType: // 切片类型
				// 如果是*[]xxx类型，则不支持
				if ft.star {
					return nil, NewUnsupportedTypeConversionError(file.Name.Name, fn.Name.Name, ips[0])
				}
				ft.isSlice = true
				t = typ.Elt
			default: // 其他类型，暂时不支持
				return nil, NewUnsupportedTypeConversionError(file.Name.Name, fn.Name.Name, ips[0])
			}
		}

		// 多个参数的类型一致,拷贝一份
		for i := range ips {
			t := ft
			t.name = ips[i]
			res = append(res, &t)
		}
	}

	return res, nil
}

// 检查该文件是否需要生成代码
// 即该文件中是否存在函数标记，例如stmapping.BuildMapping等
func (p *PackageParser) check(file *ast.File) bool {
	for _, decl := range file.Decls {
		if _, ok := decl.(*ast.FuncDecl); !ok {
			continue
		}
		if p.checkFunc(decl.(*ast.FuncDecl)) {
			return true
		}
	}

	return false
}

// 检查函数中是否有函数标记
// 比如调用stmapping.BuildMapping
func (p *PackageParser) checkFunc(fn *ast.FuncDecl) (res bool) {
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if res {
			return false
		}

		ce, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		se, ok := ce.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		x, ok := (&funcInfo{}).Match(se.Sel.Name)
		if !ok {
			return true
		}
		if ident, ok := se.X.(*ast.Ident); ok && ident.Name == x {
			res = true
			return false
		}

		return true
	})

	return
}

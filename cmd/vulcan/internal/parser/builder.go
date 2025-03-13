package parser

import (
	"fmt"
	"go/ast"
	"slices"
	"strings"

	"github.com/mangohow/mangokit/cmd/vulcan/internal/types"
)

type FuncLabel string

const (
	BuildMapping     FuncLabel = "BuildMapping"
	BuildMappingFrom           = "BuildMappingFrom"
	ByName                     = "ByName"
	ByTag                      = "ByTag"
)

var (
	labels = map[FuncLabel]string{
		BuildMapping:     "vulcan",
		BuildMappingFrom: "vulcan",
		ByName:           "vulcan",
		ByTag:            "vulcan",
	}
)

type FuncInfo struct {
	absPkg string // 绝对包名
	name   FuncLabel
	mkf    types.MappingKeyFunc
	input  []*fieldType
	output []*fieldType
}

// BuildFuncInfo 解析函数标记，并获取参数信息
func BuildFuncInfo(file *ast.File, decl *ast.FuncDecl, inputs, outputs []*fieldType, fnDesc *types.Func) (funcInfo *FuncInfo, e error) {
	funcInfo = &FuncInfo{}
	type AstInfo struct {
		label FuncLabel
		args  []ast.Expr
	}

	ais := make([]AstInfo, 0, 2)
	// 获取函数标记
	ast.Inspect(decl.Body, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		// 如果发生错误，则停止遍历
		if e != nil {
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
		x, ok := MatchFuncLabel(se.Sel.Name)
		if !ok {
			return true
		}
		sel, ok := se.X.(*ast.Ident)
		if !ok || sel.Name != x {
			return true
		}

		ais = append(ais, AstInfo{
			label: FuncLabel(se.Sel.Name),
			args:  ce.Args,
		})

		return true
	})

	fillArgs := func(fi *FuncInfo, exprArgs []ast.Expr) ([]string, error) {
		args := make([]string, 0, len(exprArgs))
		for _, arg := range exprArgs {
			// 函数中可能还有函数调用，这是非法的
			name, ok := arg.(*ast.Ident)
			if !ok {
				return nil, fmt.Errorf("func label parameter invalid, in file %s, func %s", file.Name.Name, decl.Name.Name)
			}

			args = append(args, name.Name)
		}
		return args, nil
	}

	setKeyTag := func(fi *FuncInfo, exprArgs []ast.Expr) error {
		if len(exprArgs) != 1 {
			return fmt.Errorf("ByTag func label expect one parameter, but got %d, in file %s, func %s", len(exprArgs), file.Name.Name, decl.Name.Name)
		}

		arg, ok := exprArgs[0].(*ast.BasicLit)
		if !ok {
			return fmt.Errorf("func label parameter invalid, in file %s, func %s", file.Name.Name, decl.Name.Name)
		}
		ttag := strings.Trim(arg.Value, `"`)
		fi.mkf = func(name, tag string) string {
			return types.TagMappingKeyFunc(name, tag, ttag)
		}

		return nil
	}

	var args []string
	for _, ai := range ais {
		switch ai.label {
		case BuildMapping, BuildMappingFrom:
			funcInfo.name = ai.label
			args, e = fillArgs(funcInfo, ai.args)
			if e != nil {
				return nil, e
			}
		case ByName:
			funcInfo.mkf = types.NameMappingKeyFunc
		case ByTag:
			if e := setKeyTag(funcInfo, ai.args); e != nil {
				return nil, e
			}
		}
	}

	// 没有找到函数标记
	if funcInfo.name == "" {
		return nil, UnimportantError
	}

	e = funcInfo.build(args, inputs, outputs, fnDesc)
	if e != nil {
		return nil, e
	}

	return
}

func MatchFuncLabel(name string) (string, bool) {
	val, ok := labels[FuncLabel(name)]
	return val, ok
}

// build 构建输入输出参数
func (f *FuncInfo) build(args []string, inputs, outputs []*fieldType, fnDesc *types.Func) (err error) {
	switch f.name {
	case BuildMapping: // 第一个参数为输入参数，第二个为输出参数
		err = f.buildMapping(args, inputs, outputs, fnDesc.File.Name, fnDesc.Name)
	case BuildMappingFrom: // 全部为输入参数
		err = f.buildMappingFrom(args, inputs, outputs, fnDesc.File.Name, fnDesc.Name)
	}

	return
}

func (f *FuncInfo) buildMapping(args []string, inputs, outputs []*fieldType, file, fnName string) error {
	// 生成的函数类型为func(a typea, b type b)
	// 因此输入参数只能为2，输出参数为0
	if len(inputs) != 2 {
		return fmt.Errorf("invalid func declaration, input parmater count must be 2, in file %s, func %s", file, fnName)
	}
	if len(outputs) != 0 {
		return fmt.Errorf("invalid func declaration, must have no output parameter, in file %s, func %s", file, fnName)
	}
	if len(args) != 2 {
		return fmt.Errorf("invalid func label in file %s, func %s", file, fnName)
	}
	i := f.checkParam(args[0], inputs)
	if i == -1 {
		return fmt.Errorf("invalid func label input parameter, in file %s, func %s", file, fnName)
	}
	f.input = append(f.input, inputs[i])
	j := f.checkParam(args[1], inputs)
	if j == -1 || i == j {
		return fmt.Errorf("invalid func label input parameter, in file %s, func %s", file, fnName)
	}
	f.output = append(f.output, inputs[j])

	return nil
}

func (f *FuncInfo) buildMappingFrom(args []string, inputs, outputs []*fieldType, file, fnName string) error {
	// 可以有多个输入参数
	set := make(map[string]struct{})
	for _, arg := range args {
		idx := f.checkParam(arg, inputs)
		if idx == -1 {
			return fmt.Errorf("invalid func label input parameter, in file %s, func %s", file, fnName)
		}
		// 进行一个去重，防止一个参数被多次输入
		if _, ok := set[arg]; ok {
			continue
		}
		set[arg] = struct{}{}

		f.input = append(f.input, inputs[idx])
	}

	// 将所有返回值作为输出参数
	f.output = append(f.output, outputs...)

	return nil
}

func (f *FuncInfo) checkParam(name string, fts []*fieldType) int {
	return slices.IndexFunc(fts, func(f *fieldType) bool {
		return f.name == name
	})
}

type FuncDescBuilder struct {
	// 共享的，用于存放其他包的类型声明
	sharedTypeManager *TypeManager
	// 用于存放本包的类型声明
	typeManager *TypeManager
	mkf         types.MappingKeyFunc
	pkgs        [][2]string

	fli *FuncInfo

	typeFilters []TypeFilter
}

type TypeFilter func(absPkg, typeName string) bool

// StdTimeFilter 忽略标准库Time类型
func StdTimeFilter(absPkg, typeName string) bool {
	if absPkg == "time" && typeName == "Time" {
		return true
	}

	return false
}

func NewFuncDescBuilder(sharedTypeManager *TypeManager, typeManager *TypeManager, mkf types.MappingKeyFunc, pkgs [][2]string, fli *FuncInfo) *FuncDescBuilder {
	return &FuncDescBuilder{
		sharedTypeManager: sharedTypeManager,
		typeManager:       typeManager,
		mkf:               mkf,
		pkgs:              pkgs,
		fli:               fli,
		typeFilters: []TypeFilter{
			StdTimeFilter,
		},
	}
}

func (f *FuncDescBuilder) Build(fn string) (*types.Func, error) {
	switch f.fli.name {
	case BuildMapping:
		return f.buildMappingTwoInput(fn)
	case BuildMappingFrom:
		return f.buildMappingManyToMany(fn)
	}

	return nil, fmt.Errorf("invalid func label %s", f.fli.name)
}

func (f *FuncDescBuilder) buildMappingTwoInput(name string) (*types.Func, error) {
	fn := &types.Func{
		Name: name,
	}
	if err := f.buildKnownStruct(fn, f.fli.input[0], true, false); err != nil {
		return nil, err
	}
	if err := f.buildKnownStruct(fn, f.fli.output[0], false, false); err != nil {
		return nil, err
	}
	return fn, nil
}

func (f *FuncDescBuilder) buildMappingManyToMany(name string) (*types.Func, error) {
	fn := &types.Func{
		Name: name,
	}

	for _, input := range f.fli.input {
		if err := f.buildKnownStruct(fn, input, true, false); err != nil {
			return nil, err
		}
	}

	for _, output := range f.fli.output {
		if err := f.buildKnownStruct(fn, output, false, true); err != nil {
			return nil, err
		}
	}

	return fn, nil
}

func (f *FuncDescBuilder) buildKnownStruct(fn *types.Func, ft *fieldType, isInput, isReturnParam bool) error {
	st, ok := ft.astType.Type.(*ast.StructType)
	if !ok {
		return fmt.Errorf("invalid struct type, expected struct")
	}
	pam := &types.Param{
		Name:          ft.name,
		TypeName:      ft.astType.Name.Name,
		Kind:          types.Struct,
		Package:       ft.pkg,
		AbsPackage:    ft.absPkg,
		IsPointer:     ft.star,
		IsReturnParam: isReturnParam,
		Fields:        make(map[string]*types.Param),
	}

	// 解析每个字段
	for _, field := range st.Fields.List {
		parm := &types.Param{Fields: make(map[string]*types.Param)}
		fst := field.Type
	innerloop:
		for {
			switch t := fst.(type) {
			case *ast.Ident:
				parm.TypeName = t.Name
				if parm.TypeName == "any" && parm.IsPointer {
					return fmt.Errorf("unsupported type: *any")
				}
				break innerloop
			case *ast.ArrayType: // 数组、切片类型
				if parm.IsPointer {
					return fmt.Errorf("unsupported type: *[]xxx")
				}
				parm.IsSlice = true
				fst = t.Elt
			case *ast.StarExpr: // 指针类型
				if parm.IsPointer {
					return fmt.Errorf("multi level pointers are not supported")
				}
				fst = t.X
				parm.IsPointer = true
			case *ast.SelectorExpr:
				parm.Package = t.X.(*ast.Ident).Name
				parm.TypeName = t.Sel.Name
				break innerloop
			case *ast.InterfaceType:
				if parm.IsPointer {
					return fmt.Errorf("unsupported type: *interface{}")
				}
				parm.TypeName = "interface{}"
				break innerloop
			default:
				return fmt.Errorf("unsupported type")
			}
		}

		if parm.Package == "" && types.IsBasicType(parm.Package, parm.TypeName) {
			parm.Kind = types.GetKind(parm.Package, parm.TypeName)
		} else {
			// 如果不是基础类型，则需要继续解析
			// 该类型可能与当前类型在同一个包中，需要设置package以进行查询
			if parm.Package == "" {
				parm.Package = pam.Package
				parm.AbsPackage = pam.AbsPackage
			}
			if err := f.buildOtherType(parm); err != nil {
				return err
			}
		}

		for i := 0; i < len(field.Names); i++ {
			pp := *parm
			pp.Name = field.Names[i].Name
			if field.Tag != nil {
				pp.Tag = field.Tag.Value
			}
			pp.MappingKey = f.mkf(pp.Name, pp.Tag)
			pam.Fields[pp.MappingKey] = &pp
			// 按照字段声明顺序添加
			pam.FieldNames = append(pam.FieldNames, pp.MappingKey)
		}
	}

	if isInput {
		fn.Inputs = append(fn.Inputs, pam)
	} else {
		fn.Outputs = append(fn.Outputs, pam)
	}

	return nil
}

func fullTypeName(pkg, name string) string {
	if pkg == "" {
		return name
	}
	return pkg + "." + name
}

func (f *FuncDescBuilder) buildOtherType(param *types.Param) error {
	var (
		ts     *ast.TypeSpec
		absPkg string
		err    error
	)
	// 在本包查找类型定义
	if param.Package == "" {
		ts, absPkg, err = f.typeManager.GetTypeSpec("", param.TypeName, nil)
	} else {
		ts, absPkg, err = f.typeManager.GetTypeSpec(param.Package, param.TypeName, f.pkgs)
	}
	if err != nil {
		return fmt.Errorf("find type %s error, err=%v", fullTypeName(param.Package, param.TypeName), err)
	}

	param.AbsPackage = absPkg
	typ := ts.Type

loop:
	for {
		switch t := typ.(type) {
		case *ast.InterfaceType:
			if param.IsPointer {
				return fmt.Errorf("pointers of interface types are not supported")
			}
			param.Kind = types.Interface
			break loop
		case *ast.Ident:
			param.PrimitiveType = &types.Param{TypeName: t.Name}
			if !types.IsBasicType("", t.Name) {
				return f.buildOtherType(param.PrimitiveType)
			}
			param.PrimitiveType.Kind = types.GetKind("", param.TypeName)
			break loop
		case *ast.StructType:
			param.Kind = types.Struct
			return f.buildStruct(t, param)
		case *ast.SelectorExpr:
			param.PrimitiveType = &types.Param{
				TypeName: t.Sel.Name,
				Package:  t.X.(*ast.Ident).Name,
			}
			if !types.IsBasicType(t.X.(*ast.Ident).Name, t.Sel.Name) {
				return f.buildOtherType(param.PrimitiveType)
			}
			param.PrimitiveType.Kind = types.GetKind(t.X.(*ast.Ident).Name, param.TypeName)
			break loop
		case *ast.StarExpr:
			if param.IsPointer {
				return fmt.Errorf("multi level pointers are not supported")
			}
			param.IsPointer = true
			typ = t.X
		default:
			return fmt.Errorf("invalid type %s", param.TypeName)
		}
	}

	// 判断是否存在多级指针
	isPointer := false
	for tmp := param; tmp != nil; tmp = tmp.PrimitiveType {
		if isPointer && tmp.IsPointer {
			return fmt.Errorf("multi level pointers are not supported")
		}
		if tmp.IsPointer {
			isPointer = true
		}
	}

	return nil
}

func (f *FuncDescBuilder) buildStruct(st *ast.StructType, param *types.Param) error {
	// 过滤一些特殊的结构体，比如time.Time
	for _, fn := range f.typeFilters {
		if fn(param.Package, param.TypeName) {
			param.Kind = types.GetKind(param.Package, param.TypeName)
			return nil
		}
	}
	if param.Fields == nil {
		param.Fields = make(map[string]*types.Param)
	}

	// 解析每个字段
	for _, field := range st.Fields.List {
		parm := &types.Param{}
		fst := field.Type
	innerloop:
		for {
			switch t := fst.(type) {
			case *ast.Ident:
				parm.TypeName = t.Name
				break innerloop
			case *ast.ArrayType: // 数组、切片类型
				if parm.IsPointer {
					return fmt.Errorf("unsupported type: *[]xxx")
				}
				parm.IsSlice = true
				fst = t.Elt
			case *ast.StarExpr: // 指针类型
				if parm.IsPointer {
					return fmt.Errorf("multi level pointers are not supported")
				}
				fst = t
				parm.IsPointer = true
			case *ast.SelectorExpr:
				parm.Package = t.X.(*ast.Ident).Name
				parm.TypeName = t.Sel.Name
				break innerloop
			case *ast.InterfaceType:
				if parm.IsPointer {
					return fmt.Errorf("unsupported type: *interface{}")
				}
				parm.TypeName = "interface{}"
				break innerloop
			default:
				return fmt.Errorf("unsupported type")
			}
		}

		if parm.Package != "" || !types.IsBasicType(param.Package, parm.TypeName) {
			// 如果不是基础类型，则需要继续解析
			if err := f.buildOtherType(parm); err != nil {
				return err
			}
		} else {
			parm.Kind = types.GetKind(parm.Package, parm.TypeName)
		}

		for i := 0; i < len(field.Names); i++ {
			pp := *parm
			pp.Name = field.Names[i].Name
			if field.Tag != nil {
				pp.Tag = field.Tag.Value
			}
			pp.MappingKey = f.mkf(pp.Name, pp.Tag)
			param.Fields[pp.MappingKey] = &pp
			// 按照字段声明顺序添加
			param.FieldNames = append(param.FieldNames, pp.MappingKey)
		}
	}

	return nil
}

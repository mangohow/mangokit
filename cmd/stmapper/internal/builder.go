package internal

import (
	"fmt"
	"go/ast"
	"reflect"
	"slices"
)

type FuncLabel string

var (
	basicTypes = map[string]reflect.Kind{
		"bool":        reflect.Bool,
		"int":         reflect.Int,
		"int8":        reflect.Int8,
		"int16":       reflect.Int16,
		"int32":       reflect.Int32,
		"int64":       reflect.Int64,
		"uint":        reflect.Uint,
		"uint8":       reflect.Uint8,
		"uint16":      reflect.Uint16,
		"uint32":      reflect.Uint32,
		"uint64":      reflect.Uint64,
		"float32":     reflect.Float32,
		"float64":     reflect.Float64,
		"string":      reflect.String,
		"any":         reflect.Interface,
		"interface{}": reflect.Interface,
	}
)

func isBasicType(t string) bool {
	_, ok := basicTypes[t]
	return ok
}

func reflectKind(t string) reflect.Kind {
	return basicTypes[t]
}

const (
	BuildMapping     FuncLabel = "BuildMapping"
	BuildMappingFrom           = "BuildMappingFrom"
)

var (
	labels = map[FuncLabel]string{
		BuildMapping:     "stmapper",
		BuildMappingFrom: "stmapper",
	}
)

type funcInfo struct {
	absPkg string // 绝对包名
	name   FuncLabel
	input  []*fieldType
	output []*fieldType
}

func (f *funcInfo) Match(name string) (string, bool) {
	val, ok := labels[FuncLabel(name)]
	return val, ok
}

// Build 构建输入输出参数
func (f *funcInfo) Build(args []string, inputs, outputs []*fieldType, file, fnName string) (err error) {
	switch f.name {
	case BuildMapping: // 第一个参数为输入参数，第二个为输出参数
		err = f.buildMapping(args, inputs, outputs, file, fnName)
	case BuildMappingFrom: // 全部为输入参数
		err = f.buildMappingFrom(args, inputs, outputs, file, fnName)
	}

	return
}

func (f *funcInfo) buildMapping(args []string, inputs, outputs []*fieldType, file, fnName string) error {
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

func (f *funcInfo) buildMappingFrom(args []string, inputs, outputs []*fieldType, file, fnName string) error {
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

func (f *funcInfo) checkParam(name string, fts []*fieldType) int {
	return slices.IndexFunc(fts, func(f *fieldType) bool {
		return f.name == name
	})
}

type FuncDescBuilder struct {
	// 共享的，用于存放其他包的类型声明
	sharedTypeManager *TypeManager
	// 用于存放本包的类型声明
	typeManager *TypeManager
	mkf         MappingKeyFunc
	pkgs        [][2]string

	fli *funcInfo

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

func NewFuncDescBuilder(sharedTypeManager *TypeManager, typeManager *TypeManager, mkf MappingKeyFunc, pkgs [][2]string, fli *funcInfo) *FuncDescBuilder {
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

func (f *FuncDescBuilder) Build(fn string) (*Func, error) {
	switch f.fli.name {
	case BuildMapping:
		return f.buildMappingTwoInput(fn)
	case BuildMappingFrom:
		return f.buildMappingManyToMany(fn)
	}

	return nil, fmt.Errorf("invalid func label %s", f.fli.name)
}

func (f *FuncDescBuilder) buildMappingTwoInput(name string) (*Func, error) {
	fn := &Func{
		Name: name,
	}
	if err := f.buildMapping(fn, f.fli.input[0], true, false); err != nil {
		return nil, err
	}
	if err := f.buildMapping(fn, f.fli.output[0], false, false); err != nil {
		return nil, err
	}
	return fn, nil
}

func (f *FuncDescBuilder) buildMappingManyToMany(name string) (*Func, error) {
	fn := &Func{
		Name: name,
	}

	for _, input := range f.fli.input {
		if err := f.buildMapping(fn, input, true, false); err != nil {
			return nil, err
		}
	}

	for _, output := range f.fli.output {
		if err := f.buildMapping(fn, output, false, true); err != nil {
			return nil, err
		}
	}

	return fn, nil
}

func (f *FuncDescBuilder) buildMapping(fn *Func, ft *fieldType, isInput, isReturnParam bool) error {
	st, ok := ft.astType.Type.(*ast.StructType)
	if !ok {
		return fmt.Errorf("invalid struct type, expected struct")
	}
	pam := &Param{
		Name:          ft.name,
		TypeName:      ft.astType.Name.Name,
		Kind:          reflect.Struct,
		Package:       ft.pkg,
		AbsPackage:    ft.absPkg,
		IsPointer:     ft.star,
		IsReturnParam: isReturnParam,
		Fields:        make(map[string]*Param),
	}

	// 解析每个字段
	for _, field := range st.Fields.List {
		parm := &Param{Fields: make(map[string]*Param)}
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

		if parm.Package == "" && isBasicType(parm.TypeName) {
			parm.Kind = reflectKind(parm.TypeName)
		} else {
			// 如果不是基础类型，则需要继续解析
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
		}
	}

	if isInput {
		fn.Inputs = append(fn.Inputs, pam)
	} else {
		fn.Outputs = append(fn.Outputs, pam)
	}

	return nil
}

func (f *FuncDescBuilder) buildOtherType(param *Param) error {
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
		return fmt.Errorf("find type %s error, err=%v", param.TypeName, err)
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
			param.Kind = reflect.Interface
			break loop
		case *ast.Ident:
			param.PrimitiveType = &Param{TypeName: t.Name}
			if !isBasicType(t.Name) {
				return f.buildOtherType(param.PrimitiveType)
			}
			param.PrimitiveType.Kind = reflectKind(param.TypeName)
			break loop
		case *ast.StructType:
			param.Kind = reflect.Struct
			return f.buildStruct(t, param)
		case *ast.SelectorExpr:
			param.PrimitiveType = &Param{
				TypeName: t.Sel.Name,
				Package:  t.X.(*ast.Ident).Name,
			}
			if !isBasicType(t.Sel.Name) {
				return f.buildOtherType(param.PrimitiveType)
			}
			param.PrimitiveType.Kind = reflectKind(param.TypeName)
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

func (f *FuncDescBuilder) buildStruct(st *ast.StructType, param *Param) error {
	for _, fn := range f.typeFilters {
		if fn(param.Package, param.TypeName) {
			return nil
		}
	}
	if param.Fields == nil {
		param.Fields = make(map[string]*Param)
	}

	// 解析每个字段
	for _, field := range st.Fields.List {
		parm := &Param{}
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

		if parm.Package != "" || !isBasicType(parm.TypeName) {
			// 如果不是基础类型，则需要继续解析
			if err := f.buildOtherType(parm); err != nil {
				return err
			}
		} else {
			parm.Kind = reflectKind(parm.TypeName)
		}

		for i := 0; i < len(field.Names); i++ {
			pp := *parm
			pp.Name = field.Names[i].Name
			if field.Tag != nil {
				pp.Tag = field.Tag.Value
			}
			pp.MappingKey = f.mkf(pp.Name, pp.Tag)
			param.Fields[pp.MappingKey] = &pp
		}
	}

	return nil
}

package generator

import (
	"fmt"
	"github.com/mangohow/mangokit/cmd/stmapper/internal/types"
	"github.com/mangohow/mangokit/tools/strutil"
	"go/ast"
	"go/token"
	"strings"
)

type ConverterFunc func(left, right *types.Param, leftName, rightName []string) (res *ConvertResult, err error)

func getTypeConverter(left, right types.Kind) (ConverterFunc, error) {
	key := [2]types.TypeClass{types.ToTypeClass(left), types.ToTypeClass(right)}
	converter, ok := typeConverter[key]
	if !ok {
		return nil, fmt.Errorf("cannot convert %s to %s", left.String(), right.String())
	}

	return converter, nil
}

var (
	typeConverter = map[[2]types.TypeClass]ConverterFunc{
		// 数字类型转其他类型
		[2]types.TypeClass{types.TypeNumber, types.TypeNumber}:  NumberToNumberConvert,
		[2]types.TypeClass{types.TypeNumber, types.TypeString}:  NumberToStringConvert,
		[2]types.TypeClass{types.TypeNumber, types.TypeStdTime}: NumberToStdTimeConvert,

		// string类型转其他类型
		[2]types.TypeClass{types.TypeString, types.TypeNumber}:  StringToNumber,
		[2]types.TypeClass{types.TypeString, types.TypeBool}:    StringToBool,
		[2]types.TypeClass{types.TypeString, types.TypeStdTime}: StringToStdTimeConvert,

		// bool类型转其他类型
		[2]types.TypeClass{types.TypeBool, types.TypeString}: BoolToStringConvert,

		// std time类型转其他类型
		[2]types.TypeClass{types.TypeStdTime, types.TypeString}: StdTimeToStringConvert,
		[2]types.TypeClass{types.TypeStdTime, types.TypeNumber}: StdTimeToIntConvert,
	}
)

type ConvertResult struct {
	Pkg          []*ast.ImportSpec
	ConvertedAst []ast.Stmt
	Rhs          []ast.Expr
}

// NumberToNumberConvert 数字类型到数字类型的转换
// int64(b)
// a_b_xxx := int64(b)
// a = &a_b_xxx
func NumberToNumberConvert(left, right *types.Param, leftName, rightName []string) (res *ConvertResult, err error) {
	res = &ConvertResult{}
	callExpr := BuildCallExpr(ast.NewIdent(left.Kind.String()), []ast.Expr{BuildVarNameExpr(rightName, right.IsPointer)})

	if !left.IsPointer {
		res.Rhs = append(res.Rhs, callExpr)
		return res, nil
	}

	return obtainAddressConvert(left, right, callExpr, res), nil
}

// NumberToStringConvert 数字类型转string
// strconv.FormatInt(int64(b), 10)
// strconv.FormatUint(uint64(b), 10)
// strconv.FormatFloat(float64(b), 'g', -1, 64)
// a_b_xxx := strconv.FormatInt(int64(b), 10)
// a = &a_b_xxx
func NumberToStringConvert(left, right *types.Param, leftName, rightName []string) (res *ConvertResult, err error) {
	return numberOrBoolToStringConvert(left, right, leftName, rightName)
}
func numberOrBoolToStringConvert(left, right *types.Param, leftName, rightName []string) (res *ConvertResult, err error) {
	var (
		args []ast.Expr
		conv ast.Expr
	)

	switch {
	case types.IsInt(right.Kind):
		args = append(args, buildNumberTypeConvert(&types.Param{Kind: types.Int64}, right, rightName))
		args = append(args, BuildBasicLit(token.INT, "10"))
		conv = BuildCallExpr(BuildAstSelectorExpr([]string{"strconv", "FormatInt"}), args)
	case types.IsUint(right.Kind):
		args = append(args, buildNumberTypeConvert(&types.Param{Kind: types.Uint64}, right, rightName))
		args = append(args, BuildBasicLit(token.INT, "10"))
		conv = BuildCallExpr(BuildAstSelectorExpr([]string{"strconv", "FormatUint"}), args)
	case types.IsFloat(right.Kind):
		args = append(args, buildNumberTypeConvert(&types.Param{Kind: types.Float64}, right, rightName))
		args = append(args, BuildBasicLit(token.CHAR, "g"))
		args = append(args, BuildUnaryExpr("-", BuildBasicLit(token.INT, "1")))
		args = append(args, BuildBasicLit(token.INT, "64"))
		conv = BuildCallExpr(BuildAstSelectorExpr([]string{"strconv", "FormatFloat"}), args)
	case types.IsBool(right.Kind):
		args = append(args, BuildVarNameExpr(rightName, right.IsPointer))
		conv = BuildCallExpr(BuildAstSelectorExpr([]string{"strconv", "FormatBool"}), args)
	}
	res = &ConvertResult{
		Pkg: []*ast.ImportSpec{BuildImportSpec("strconv")},
	}

	if !left.IsPointer {
		res.Rhs = append(res.Rhs, conv)
		return res, nil
	}

	return obtainAddressConvert(left, right, conv, res), nil
}

func buildNumberTypeConvert(left, right *types.Param, rightName []string) ast.Expr {
	if right.Kind == left.Kind {
		return BuildVarNameExpr(rightName, right.IsPointer)
	}

	r, _ := NumberToNumberConvert(left, right, []string{}, rightName)
	return r.Rhs[0]
}

// 获取地址 a = &b
func obtainAddressConvert(left, right *types.Param, rightConv ast.Expr, res *ConvertResult) *ConvertResult {
	name := randVarName(left.Name, right.Name)
	c := BuildDefineStem([]ast.Expr{ast.NewIdent(name)}, []ast.Expr{rightConv})
	r := BuildUnaryExpr("&", ast.NewIdent(name))
	res.ConvertedAst = append(res.ConvertedAst, c)
	res.Rhs = append(res.Rhs, r)

	return res
}

func randVarName(left, right string) string {
	return strings.ToLower(left) + "_" + strings.ToLower(right) + "_" + strutil.RandLowerString(4)
}

// NumberToStdTimeConvert int or int64 -> time.Time
// time.Unix(b, 0)
// a_b_xxx = time.Unix(b, 0)
// a = &a_b_xxx
func NumberToStdTimeConvert(left, right *types.Param, leftName, rightName []string) (res *ConvertResult, err error) {
	if right.Kind != types.Int && right.Kind != types.Int64 {
		return nil, fmt.Errorf("cant convert %s to time.Time", right.Kind.String())
	}
	res = &ConvertResult{
		Pkg: []*ast.ImportSpec{BuildImportSpec("time")},
	}
	args := make([]ast.Expr, 0, 2)
	args = append(args, buildNumberTypeConvert(&types.Param{Kind: types.Int64}, right, rightName))
	args = append(args, BuildBasicLit(token.INT, "0"))
	conv := BuildCallExpr(BuildAstSelectorExpr([]string{"time", "Unix"}), args)
	if !left.IsPointer {
		res.Rhs = append(res.Rhs, conv)
		return res, nil
	}

	return obtainAddressConvert(left, right, conv, res), nil
}

// StringToNumber string -> int, uint, float
// a_b_xxx, _ := strconv.ParseInt(b, 10, 64)
// a_b_xxx, _ := strconv.ParseUint(b, 10, 64)
// a_b_xxx, _ := strconv.ParseFloat(b, 64)
// a = a_b_xxx
// or
// a = &a_b_xxx
func StringToNumber(left, right *types.Param, leftName, rightName []string) (res *ConvertResult, err error) {
	return stringToNumberOrBool(left, right, leftName, rightName)
}

func stringToNumberOrBool(left, right *types.Param, leftName, rightName []string) (res *ConvertResult, err error) {
	var (
		args []ast.Expr
		conv ast.Expr
	)
	args = append(args, BuildVarNameExpr(rightName, right.IsPointer))
	switch {
	case types.IsInt(left.Kind):
		args = append(args, BuildBasicLit(token.INT, "10"))
		args = append(args, BuildBasicLit(token.INT, "64"))
		conv = BuildCallExpr(BuildAstSelectorExpr([]string{"strconv", "ParseInt"}), args)
	case types.IsUint(left.Kind):
		args = append(args, BuildBasicLit(token.INT, "10"))
		args = append(args, BuildBasicLit(token.INT, "64"))
		conv = BuildCallExpr(BuildAstSelectorExpr([]string{"strconv", "ParseUint"}), args)
	case types.IsFloat(left.Kind):
		args = append(args, BuildBasicLit(token.INT, "64"))
		conv = BuildCallExpr(BuildAstSelectorExpr([]string{"strconv", "ParseFloat"}), args)
	case types.IsBool(left.Kind):
		conv = BuildCallExpr(BuildAstSelectorExpr([]string{"strconv", "ParseBool"}), args)
	}
	name := randVarName(left.Name, right.Name)
	assign := BuildDefineStem([]ast.Expr{
		ast.NewIdent(name),
		ast.NewIdent("_"),
	}, []ast.Expr{conv})

	res = &ConvertResult{
		Pkg:          []*ast.ImportSpec{BuildImportSpec("strconv")},
		ConvertedAst: []ast.Stmt{assign},
	}

	rh := ast.Expr(ast.NewIdent(name))
	if left.IsPointer {
		rh = BuildUnaryExpr("&", rh)
	}
	res.Rhs = append(res.Rhs, rh)

	return res, nil
}

// StringToBool string -> bool
// a_b_xxx, _ := strconv.ParseBool(b)
// a = a_b_xxx
// or
// a = &a_b_xxx
func StringToBool(left, right *types.Param, leftName, rightName []string) (res *ConvertResult, err error) {
	return stringToNumberOrBool(left, right, leftName, rightName)
}

// StringToStdTimeConvert string -> time.Time
// a_b_xxx, _ := time.Parse(time.DateTime, b)
// a = a_b_xxx
// or
// a = &a_b_xxx
func StringToStdTimeConvert(left, right *types.Param, leftName, rightName []string) (res *ConvertResult, err error) {
	var (
		args = []ast.Expr{BuildAstSelectorExpr([]string{"time", "DateTime"}), BuildVarNameExpr(rightName, right.IsPointer)}
		name = randVarName(left.Name, right.Name)
		conv = BuildCallExpr(BuildAstSelectorExpr([]string{"time", "Parse"}), args)
	)
	stmt := BuildDefineStem([]ast.Expr{ast.NewIdent(name), ast.NewIdent("_")}, []ast.Expr{conv})
	rh := ast.Expr(ast.NewIdent(name))
	if left.IsPointer {
		rh = BuildUnaryExpr("&", rh)
	}
	res = &ConvertResult{
		Pkg:          []*ast.ImportSpec{BuildImportSpec("time")},
		ConvertedAst: []ast.Stmt{stmt},
		Rhs:          []ast.Expr{rh},
	}

	return res, nil
}

// BoolToStringConvert bool -> string
// strconv.FormatBool(b)
// a_b_xxx = strconv.FormatBool(b)
// a = &a_b_xxx
func BoolToStringConvert(left, right *types.Param, leftName, rightName []string) (res *ConvertResult, err error) {
	return numberOrBoolToStringConvert(left, right, leftName, rightName)
}

// StdTimeToStringConvert time.Time -> string
// t.Format(time.DateTime)
// a_b_xxx := t.Format(time.DataTime)
// a = &a_b_xxx
func StdTimeToStringConvert(left, right *types.Param, leftName, rightName []string) (res *ConvertResult, err error) {
	expr := BuildCallExpr(BuildAstSelectorExpr(append(rightName, "Format")), []ast.Expr{BuildAstSelectorExpr([]string{"time", "DateTime"})})
	res = &ConvertResult{
		Pkg: []*ast.ImportSpec{BuildImportSpec("time")},
	}
	if !left.IsPointer {
		res.Rhs = append(res.Rhs, expr)
		return res, nil
	}

	return obtainAddressConvert(left, right, expr, res), nil
}

// StdTimeToIntConvert time.Time -> int, int64
// t.Unix()
// a_b_xxx := t.Unix()
// a = &a_b_xxx
func StdTimeToIntConvert(left, right *types.Param, leftName, rightName []string) (res *ConvertResult, err error) {
	expr := BuildCallExpr(BuildAstSelectorExpr(append(rightName, "Unix")), nil)
	res = &ConvertResult{}
	if !left.IsPointer {
		res.Rhs = append(res.Rhs, expr)
		return res, nil
	}

	return obtainAddressConvert(left, right, expr, res), nil
}

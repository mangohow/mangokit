package generator

import (
	"fmt"
	"go/ast"
	gotoken "go/token"
)

// BuildAstSelectorExpr 构建 a.b.c 表达式
func BuildAstSelectorExpr(names []string) *ast.SelectorExpr {
	if len(names) < 2 {
		return nil
	}

	se := &ast.SelectorExpr{
		X:   ast.NewIdent(names[0]),
		Sel: ast.NewIdent(names[1]),
	}
	i := 2
	for ; i < len(names); i++ {
		se = &ast.SelectorExpr{
			X:   se,
			Sel: ast.NewIdent(names[i]),
		}
	}

	return se
}

// BuildAssignStmt 构建赋值语句
func BuildAssignStmt(left [][]string, right []string) *ast.AssignStmt {
	lhs := make([]ast.Expr, 0, 1)
	rhs := make([]ast.Expr, 0, 1)
	for _, l := range left {
		if len(l) == 1 {
			lhs = append(lhs, ast.NewIdent(l[0]))
		} else {
			lhs = append(lhs, BuildAstSelectorExpr(l))
		}
	}

	if len(right) == 1 {
		rhs = append(rhs, ast.NewIdent(right[0]))
	} else {
		rhs = append(rhs, BuildAstSelectorExpr(right))
	}

	return &ast.AssignStmt{
		Lhs: lhs,
		Rhs: rhs,
		Tok: gotoken.ASSIGN,
	}
}

func BuildAssignStmt1(left []ast.Expr, right []ast.Expr) *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: left,
		Rhs: right,
		Tok: gotoken.ASSIGN,
	}
}

func BuildDefineStem(left []ast.Expr, right []ast.Expr) *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: left,
		Rhs: right,
		Tok: gotoken.DEFINE,
	}
}

// BuildUnaryExpr 构建 &a *b
func BuildUnaryExpr(token string, x ast.Expr) *ast.UnaryExpr {
	var op gotoken.Token
	switch token {
	case "&":
		op = gotoken.AND
	case "*":
		op = gotoken.MUL
	case "-":
		op = gotoken.SUB
	}
	return &ast.UnaryExpr{
		Op: op,
		X:  x,
	}
}

// BuildTypeAssertExpr 构建断言 b.(int) b.(*int)
func BuildTypeAssertExpr(x ast.Expr, typename string, isPointer bool) *ast.TypeAssertExpr {
	res := &ast.TypeAssertExpr{
		X: x,
	}
	t := ast.NewIdent(typename)
	if !isPointer {
		return res
	}

	res.Type = &ast.StarExpr{
		X: t,
	}

	return res
}

func BuildCallExpr(fn ast.Expr, args []ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:  fn,
		Args: args,
	}
}

func BuildBasicLit(kind gotoken.Token, val string) *ast.BasicLit {
	return &ast.BasicLit{
		Kind:  kind,
		Value: val,
	}
}

func BuildImportSpec(pkg string) *ast.ImportSpec {
	return &ast.ImportSpec{
		Path: &ast.BasicLit{
			ValuePos: 0,
			Kind:     gotoken.STRING,
			Value:    fmt.Sprintf("\"%s\"", pkg),
		},
	}
}

func BuildVarNameExpr(name []string, isPointer bool) ast.Expr {
	var expr ast.Expr
	if len(name) == 1 {
		expr = ast.NewIdent(name[0])
	} else {
		expr = BuildAstSelectorExpr(name)
	}

	if isPointer {
		expr = BuildUnaryExpr("*", expr)
	}

	return expr
}

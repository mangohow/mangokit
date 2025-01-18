package internal

import (
	"github.com/mangohow/mangokit/cmd/stmapper/internal/generator"
	"go/ast"
	"go/token"
	"testing"
)

func TestSelectorExprBuilder(t *testing.T) {
	res := generator.BuildAstSelectorExpr([]string{"u2", "U", "U", "Username"})
	ast.Print(token.NewFileSet(), res)
}

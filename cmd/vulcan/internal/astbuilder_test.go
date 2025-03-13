package internal

import (
	"go/ast"
	"go/token"
	"testing"

	"github.com/mangohow/mangokit/cmd/vulcan/internal/generator"
)

func TestSelectorExprBuilder(t *testing.T) {
	res := generator.BuildAstSelectorExpr([]string{"u2", "U", "U", "Username"})
	ast.Print(token.NewFileSet(), res)
}

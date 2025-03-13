package examples

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"testing"
)

var src = `
package main
import (
"fmt"
"strconv"
)

type UserProto struct {
	Id int
}

type User struct {
	Id int
}

func main() {
	u1 := UserProto{}
	u2 := User{}
	u2.Id = u1.Id
	u3 := User{
		Id: u1.Id,
	}
	n, _ := strconv.Atoi("123")
	fmt.Println(u1, u2, u3, n)
}


func test(a, b string) (c, d int) {

}
`

func TestAstParse(t *testing.T) {
	AstParse()
}

func TestGenerate(t *testing.T) {
	GenerateCode()
}

func TestPrintAst(t *testing.T) {
	PrintAst()
}

func GenerateCode() {
	specs := ParseAst()
	for _, spec := range specs {
		ast.Print(nil, spec)
	}

	fd := &ast.FuncDecl{
		Name: &ast.Ident{
			Name: "AtoB",
		},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							{
								Name: "u1",
							},
							{
								Name: "u2",
							},
						},
						Type: &ast.Ident{
							Name: specs[0].Name.Name,
						},
					},
					{
						Names: []*ast.Ident{
							{
								Name: "u2",
							},
						},
						Type: &ast.Ident{
							Name: specs[1].Name.Name,
						},
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Tok: token.ASSIGN,
					Lhs: []ast.Expr{
						&ast.SelectorExpr{
							X: &ast.Ident{
								Name: "u2",
							},
							Sel: &ast.Ident{
								Name: "Id",
							},
						},
					},
					Rhs: []ast.Expr{
						&ast.SelectorExpr{
							X: &ast.Ident{
								Name: "u1",
							},
							Sel: &ast.Ident{
								Name: "Id",
							},
						},
					},
				},
			},
		},
	}
	buffer := bytes.NewBuffer(nil)
	err := format.Node(buffer, token.NewFileSet(), fd)
	if err != nil {
		panic(err)
	}
	fmt.Println(buffer.String())
}

func PrintAst() {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		panic(err)
	}
	ast.Print(fset, f)
}

func ParseAst() []*ast.TypeSpec {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		panic(err)
	}

	specs := []*ast.TypeSpec{}
	ast.Inspect(f, func(n ast.Node) bool {
		if _, ok := n.(*ast.TypeSpec); !ok {
			return true
		}

		fd := n.(*ast.TypeSpec)
		specs = append(specs, fd)

		return true
	})

	return specs
}

func AstParse() {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "example.go", nil, parser.ParseComments)
	if err != nil {
		log.Fatal("parse file error: ", err)
	}
	file, err := os.OpenFile("example.ast.txt", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	if err := ast.Fprint(file, fset, f, nil); err != nil {
		log.Fatal(err)
	}
}

package parser

import (
	"reflect"
	"testing"

	"github.com/tombuente/lily/ast"
	"github.com/tombuente/lily/lexer"
)

type parserTest struct {
	name     string
	src      string
	expected *ast.Program
}

func TestParse(t *testing.T) {
	tests := []parserTest{
		{
			name: "identifier",
			src:  "a",
			expected: &ast.Program{
				Stmts: []ast.Stmt{
					&ast.ExprStmt{
						Expr: &ast.Ident{Value: "a"},
					},
				},
			},
		},
		{
			name: "addition",
			src:  "a + b",
			expected: &ast.Program{
				Stmts: []ast.Stmt{
					&ast.ExprStmt{
						Expr: &ast.BinaryOp{
							Op:    "+",
							Left:  &ast.Ident{Value: "a"},
							Right: &ast.Ident{Value: "b"},
						},
					},
				},
			},
		},
		{
			name: "function declaration and call",
			src:  "let add = fn(x, y) { x + y }; add(1, 2)",
			expected: &ast.Program{
				Stmts: []ast.Stmt{
					&ast.LetStmt{
						Ident: &ast.Ident{Value: "add"},
						Expr: &ast.Function{
							Params: []*ast.Ident{
								{Value: "x"},
								{Value: "y"},
							},
							Body: &ast.BlockStmt{
								Stmts: []ast.Stmt{
									&ast.ExprStmt{
										Expr: &ast.BinaryOp{
											Op:    "+",
											Left:  &ast.Ident{Value: "x"},
											Right: &ast.Ident{Value: "y"},
										},
									},
								},
							},
						},
					},
					&ast.ExprStmt{
						Expr: &ast.Call{
							Lhs:  &ast.Ident{Value: "add"},
							Args: []ast.Expr{&ast.Int{Value: 1}, &ast.Int{Value: 2}},
						},
					},
				},
			},
		},
		{
			name: "if",
			src:  "let x = if true { 1 } { 2 }",
			expected: &ast.Program{
				Stmts: []ast.Stmt{
					&ast.LetStmt{
						Ident: &ast.Ident{Value: "x"},
						Expr: &ast.If{
							Condition: &ast.Bool{Value: true},
							Consequence: &ast.BlockStmt{
								Stmts: []ast.Stmt{&ast.ExprStmt{Expr: &ast.Int{Value: 1}}},
							},
							Alternative: &ast.BlockStmt{
								Stmts: []ast.Stmt{&ast.ExprStmt{Expr: &ast.Int{Value: 2}}},
							},
						},
					},
				},
			},
		},
	}

	test(t, tests)
}

func test(t *testing.T, tests []parserTest) {
	t.Helper()
	for _, tt := range tests {
		name := tt.name
		if name == "" {
			name = tt.src
		}
		t.Run(name, func(t *testing.T) {
			t.Helper()
			program := parse(t, tt.src)
			if !reflect.DeepEqual(program, tt.expected) {
				expectedJSON, err := tt.expected.MarshalJSON()
				if err != nil {
					t.Fatalf("Failed to marshal expected: %v", err)
				}
				programJOSN, err := program.MarshalJSON()
				if err != nil {
					t.Fatalf("Failed to marshal program: %v", err)
				}
				t.Fatalf("want=%v, got %v", string(expectedJSON), string(programJOSN))
			}
		})
	}
}

func parse(t *testing.T, src string) *ast.Program {
	t.Helper()
	l := lexer.New(src)
	p := New(l)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parser has errors: %v", err)
	}
	return program
}

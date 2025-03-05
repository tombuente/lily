package parser

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/tombuente/lily/lexer"
)

// func TestLetStatements(t *testing.T) {
// 	tests := []struct {
// 		src           string
// 		expectedIdent string
// 		expectedValue any
// 	}{
// 		{"let x = 5", "x", 5},
// 	}

// 	for _, tt := range tests {
// 		l := lexer.New(tt.src)
// 		p := New(l)
// 		prog, err := p.Parse()
// 		if err != nil {
// 			t.Error(err)
// 			return
// 		}

// 		for i, test := range tests {
// 			statement := prog.Stmts[i]
// 			if !testLetStatement(t, statement, test.expectedIdent) {
// 				return
// 			}
// 		}
// 	}
// }

// func testLetStatement(t *testing.T, statement ast.Stmt, name string) bool {
// 	letStmt, ok := statement.(*ast.LetStmt)
// 	if !ok {
// 		t.Errorf("statement not *ast.LetStatement. got=%v", statement)
// 		return false
// 	}

// 	if letStmt.Ident.Value != name {
// 		t.Errorf("letStmt.Name.Value not '%v'. got=%v", name, letStmt.Ident.Value)
// 		return false
// 	}

// 	if letStmt.Ident.Token.Literal != name {
// 		t.Errorf("letStmt.Name not '%v'. got=%v", name, letStmt.Ident.Token.Literal)
// 		return false
// 	}

// 	return true
// }

func TestBadStmt(t *testing.T) {
	src := "let 0"

	l := lexer.New(src)
	p := New(l)
	prog, err := p.Parse()

	if len(prog.Stmts) != 0 {
		t.Errorf("program should have no statements, got %v", len(prog.Stmts))
	}
	if err == nil {
		t.Errorf("Parse should return error")
	}
}

func TestPrefixExprStmt(t *testing.T) {
	src := `1+1; 1+1`

	l := lexer.New(src)
	p := New(l)
	prog, err := p.Parse()
	if err != nil {
		t.Error(err)
	}

	jsonData, err := json.MarshalIndent(prog, "", "    ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}
	fmt.Println(string(jsonData))
}

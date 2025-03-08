package parser

import (
	"reflect"
	"testing"

	"github.com/tombuente/lily/ast"
	"github.com/tombuente/lily/lexer"
)

func TestParseIdentExpr(t *testing.T) {
	src := "ii"
	expected := &ast.IdentExpr{Value: "ii"}
	p := New(lexer.New(src))

	expr, err := p.parseIdentExpr()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, expr) {
		t.Fatalf("Expected %v, got %v", expected, expr)
	}
}

func TestParseIntExpr(t *testing.T) {
	src := "1"
	expected := &ast.IntExpr{Value: 1}
	p := New(lexer.New(src))

	expr, err := p.parseIntExpr()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, expr) {
		t.Fatalf("Expected %v, got %v", expected, expr)
	}
}

func TestParseLetStmt(t *testing.T) {

	src := "let x = 1"
	expected := &ast.LetStmt{Ident: &ast.IdentExpr{Value: "x"}, Expr: &ast.IntExpr{Value: 1}}
	p := New(lexer.New(src))

	stmt, err := p.parseLetStmt()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, stmt) {
		t.Fatalf("Expected %v, got %v", expected, stmt)
	}
}

func TestManual(t *testing.T) {

	src := "let add = fn(x, y) { return x + y }; add(1, 2)"
	p := New(lexer.New(src))

	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	jsonData, err := ast.MarshalIndent(prog, "", "    ")
	if err != nil {
		t.Fatalf("error marshaling JSON: %v", err)
		return
	}
	t.Log(string(jsonData))
}

// type testLexer struct {
// 	toks []token.Token
// 	pos  int
// }

// func newLexer(toks []token.Token) *testLexer {
// 	toks = append(toks, token.Token{Type: token.EOF, Literal: "0"})
// 	return &testLexer{toks: toks}
// }

// func (l *testLexer) Next() token.Token {
// 	tok := l.toks[l.pos]
// 	l.pos++
// 	return tok
// }

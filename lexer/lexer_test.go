package lexer

import (
	"fmt"
	"testing"
)

func TestManual(t *testing.T) {
	src := `"test" "test2"`

	l := New(src)
	for range 2 {
		actual := l.Next()
		fmt.Println(actual)
	}
}

// func TestArithmeticOperators(t *testing.T) {
// 	src := "+-*/<>==()"
// 	expected := []token.Token{
// 		{Type: token.Plus, Literal: "+"},
// 		{Type: token.Minus, Literal: "-"},
// 		{Type: token.Asterisk, Literal: "*"},
// 		{Type: token.Slash, Literal: "/"},
// 		{Type: token.Less, Literal: "<"},
// 		{Type: token.Greater, Literal: ">"},
// 		{Type: token.EQ, Literal: "=="},
// 		{Type: token.LParan, Literal: "("},
// 		{Type: token.RParan, Literal: ")"},
// 	}

// 	l := New(src)
// 	for i, expected := range expected {
// 		actual := l.Next()
// 		fmt.Println(actual)
// 		if actual != expected {
// 			t.Errorf("test[%d] - wrong token. expected=%+v, got=%+v", i, expected, actual)
// 		}
// 	}
// }

// func TestSkipWhitespace(t *testing.T) {
// 	src := "  =  + - * /  "
// 	expected := []token.Token{
// 		{Type: token.Assign, Literal: "="},
// 		{Type: token.Plus, Literal: "+"},
// 		{Type: token.Minus, Literal: "-"},
// 		{Type: token.Asterisk, Literal: "*"},
// 		{Type: token.Slash, Literal: "/"},
// 		{Type: token.EOF, Literal: ""},
// 	}

// 	l := New(src)
// 	for i, expected := range expected {
// 		actual := l.Next()
// 		if actual != expected {
// 			t.Errorf("test[%d] - wrong token. expected=%+v, got=%+v", i, expected, actual)
// 		}
// 	}
// }

// func TestLet(t *testing.T) {
// 	src := "let x = 5"
// 	expected := []token.Token{
// 		{Type: token.Let, Literal: "let"},
// 		{Type: token.Ident, Literal: "x"},
// 		{Type: token.Assign, Literal: "="},
// 		{Type: token.Int, Literal: "5"},
// 	}

// 	l := New(src)
// 	for i, expected := range expected {
// 		actual := l.Next()
// 		if actual != expected {
// 			t.Errorf("test[%d] - wrong token. expected=%+v, got=%+v", i, expected, actual)
// 		}
// 	}
// }

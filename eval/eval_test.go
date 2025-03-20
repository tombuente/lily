package eval

import (
	"errors"
	"reflect"
	"testing"

	"github.com/tombuente/lily/lexer"
	"github.com/tombuente/lily/parser"
)

type evalTest struct {
	name     string
	src      string
	expected object
}

type errorTest struct {
	name string
	src  string
}

func TestEval(t *testing.T) {
	tests := []evalTest{
		{src: "1", expected: &intObject{value: 1}},
		{src: "-1", expected: &intObject{value: -1}},
		{src: "1 + 1", expected: &intObject{value: 2}},
		{src: "1 - 1", expected: &intObject{value: 0}},
		{src: "3 * 3", expected: &intObject{value: 9}},
		{src: "9 / 3", expected: &intObject{value: 3}},
		{src: "2 + 3 * 4", expected: &intObject{value: 14}},
		{src: "(2 + 3) * 4", expected: &intObject{value: 20}},
		{src: "2 > 1", expected: trueInstance},
		{src: "1 > 1", expected: falseInstance},
		{src: "1 < 2", expected: trueInstance},
		{src: "1 < 1", expected: falseInstance},
		{src: "1 == 1", expected: trueInstance},
		{src: "1 == 2", expected: falseInstance},
		{src: "1 != 2", expected: trueInstance},
		{src: "1 != 1", expected: falseInstance},
		{src: "true", expected: trueInstance},
		{src: "false", expected: falseInstance},
		{src: "!true", expected: falseInstance},
		{src: "!false", expected: trueInstance},
		{src: "!!true", expected: trueInstance},
		{src: "true == true", expected: trueInstance},
		{src: "if (true) { 10 }", expected: &intObject{value: 10}},
		{src: "if (false) { 10 }", expected: nilInstance},
		{src: "if (false) { 10 } { 20 }", expected: &intObject{value: 20}},
		{src: "1; return 2; 3;", expected: &intObject{value: 2}},
		{
			name: "return first return expr",
			src: `
				if (10 > 1) {
					if (10 > 1) {
						return 10;
					}
					return 1;
				}`,
			expected: &intObject{value: 10},
		},
		{src: "let x = 5; x;", expected: &intObject{value: 5}},
		{src: "let add = fn(x, y) { x+y }; add(1, 2);", expected: &intObject{value: 3}},
		{src: "let add = fn(x, y) { x+y }; add(3, add(1, 2));", expected: &intObject{value: 6}},
		{
			name: "env capture",
			src: `
				let outer = 5;
				let funnyAdd = fn(x) {
					return outer + x;
				}
				funnyAdd(5);`,
			expected: &intObject{value: 10},
		},
		{
			name: "assignment",
			src:  "let x = 5; x = 10; x", expected: &intObject{value: 10},
		},
		{
			name: "mutate captured env",
			src: `
				let outer = 5;
				let mutate = fn() {
					outer = 10;
				};
				let funnyAdd = fn(x) {
					return outer + x;
				}
				mutate();
				funnyAdd(5);`,
			expected: &intObject{value: 15},
		},
		{
			src:      `let x = "tom"; x`,
			expected: &stringObject{value: "tom"},
		},
		{
			name:     "string concatenation",
			src:      `let x = "hello" + " " + "world"; x`,
			expected: &stringObject{value: "hello world"},
		},
	}

	test(t, tests)
}

func TestBuiltin(t *testing.T) {
	tests := []evalTest{
		{src: `len("123")`, expected: &intObject{value: 3}},
		{name: "override len", src: `let len = fn(x) { 1 }; len("123")`, expected: &intObject{value: 1}},
	}

	test(t, tests)
}

func TestTypeError(t *testing.T) {
	tests := []errorTest{
		{src: "-true"},
		{src: "!1"},
		{src: "1 > true; 1"},
		{name: "nested type error", src: "true == (1 > true); 1"},
	}

	testError[*typeError](t, tests)
}

func TestNameError(t *testing.T) {
	tests := []errorTest{
		{name: "test double declaration", src: "let x = 5; let x = 6;"},
		{name: "x undefined", src: "x = 5;"},
	}

	testError[*nameError](t, tests)
}

func test(t *testing.T, tests []evalTest) {
	t.Helper()
	for _, tt := range tests {
		name := tt.name
		if name == "" {
			name = tt.src
		}
		t.Run(name, func(t *testing.T) {
			t.Helper()
			res, err := evalHelper(t, tt.src)
			if err != nil {
				t.Fatalf("Failed with error: %v", err)
			}
			if !reflect.DeepEqual(res, tt.expected) {
				t.Fatalf("want=%v, got=%v", tt.expected, res)
			}
		})
	}
}

func testError[T error](t *testing.T, tests []errorTest) {
	t.Helper()
	for _, tt := range tests {
		name := tt.name
		if name == "" {
			name = tt.src
		}
		t.Run(name, func(t *testing.T) {
			t.Helper()
			_, err := evalHelper(t, tt.src)
			if err == nil {
				t.Fatalf("Expected error")
			}
			var asErr T
			if !errors.As(err, &asErr) {
				t.Fatalf("want=%v, got=%v", asErr, err)
			}
		})
	}
}

// evalHelper parses and evaluates the given source code, returning the result.
// It fails the test on parsing errors.
func evalHelper(t *testing.T, src string) (object, error) {
	t.Helper()
	l := lexer.New(src)
	p := parser.New(l)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}
	return Eval(prog)
}

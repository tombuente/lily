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
		{src: "1", expected: &intObject{Value: 1}},
		{src: "-1", expected: &intObject{Value: -1}},
		{src: "1 + 1", expected: &intObject{Value: 2}},
		{src: "1 - 1", expected: &intObject{Value: 0}},
		{src: "3 * 3", expected: &intObject{Value: 9}},
		{src: "9 / 3", expected: &intObject{Value: 3}},
		{src: "2 + 3 * 4", expected: &intObject{Value: 14}},
		{src: "(2 + 3) * 4", expected: &intObject{Value: 20}},
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
		{src: "if (true) { 10 }", expected: &intObject{Value: 10}},
		{src: "if (false) { 10 }", expected: nilInstance},
		{src: "if (false) { 10 } { 20 }", expected: &intObject{Value: 20}},
		{src: "1; return 2; 3;", expected: &intObject{Value: 2}},
		{
			name: "return first return expr",
			src: `
if (10 > 1) {
if (10 > 1) {
	return 10;
}
return 1;
}`,
			expected: &intObject{Value: 10},
		},
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

func test(t *testing.T, tests []evalTest) {
	t.Helper()
	for _, tt := range tests {
		name := tt.name
		if name == "" {
			name = tt.src
		}
		t.Run(name, func(t *testing.T) {
			t.Helper()
			res, err := eval(t, tt.src)
			if err != nil {
				t.Fatalf("eval failed with error: %v", err)
			}
			if !reflect.DeepEqual(res, tt.expected) {
				t.Fatalf("src=\"%v\", want=%v, got=%v", tt.src, tt.expected, res)
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
			_, err := eval(t, tt.src)
			if err == nil {
				t.Fatalf("eval did not return error")
			}
			var asErr T
			if !errors.As(err, &asErr) {
				t.Fatalf("src=%v, want=%v, got=%v", tt.src, asErr, err)
			}
		})
	}
}

// eval parses and evaluates the given source code, returning the result.
// It fails the test on parsing errors.
func eval(t *testing.T, src string) (object, error) {
	t.Helper()
	l := lexer.New(src)
	p := parser.New(l)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}
	return Eval(prog)
}

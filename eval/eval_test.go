package eval

import (
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

type evalErrorTest struct {
	name     string
	src      string
	expected errorType
}

func TestEval(t *testing.T) {
	tests := []evalTest{
		{src: "-true", expected: &intObject{Value: 1}},
		{src: "!1", expected: &intObject{Value: 1}},
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

	runEvalTest(t, tests)
}

func runEvalTest(t *testing.T, tests []evalTest) {
	t.Helper()
	for _, tt := range tests {
		name := tt.name
		if name == "" {
			name = tt.src
		}
		t.Run(name, func(t *testing.T) {
			t.Helper()
			res := eval(t, tt.src)
			if !reflect.DeepEqual(res, tt.expected) {
				t.Errorf("src=\"%v\", want=%v, got=%v", tt.src, tt.expected, res)
			}
		})
	}
}

func TestError(t *testing.T) {
	tests := []evalErrorTest{
		{src: "-true", expected: typeError},
		{src: "!1", expected: typeError},
	}

	for _, tt := range tests {
		name := tt.name
		if name == "" {
			name = tt.src
		}
		t.Run(name, func(t *testing.T) {
			res := eval(t, tt.src)
			errObj, ok := res.(*errorObject)
			if !ok {
				t.Error("Result is not *errorObject")
			}
			if errObj.errorType != tt.expected {
				t.Errorf("src=%v, want=%v, got=%v", tt.src, tt.expected, errObj.errorType)
			}
		})
	}
}

// eval parses and evaluates the given source code, returning the result.
// It fails the test on parsing errors.
func eval(t *testing.T, src string) object {
	t.Helper()
	l := lexer.New(src)
	p := parser.New(l)
	prog, err := p.Parse()
	if err != nil {
		t.Errorf("Failed to parse program: %v", err)
	}
	return Eval(prog)
}

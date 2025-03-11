package eval

import "fmt"

const (
	typeError errorType = "TypeError"
)

type object interface {
	DebugTypeInfo() string
}

type intObject struct {
	Value int64
}

type boolObject struct {
	Value bool
}

// Used for type assertion to return early
type returnObject struct {
	Value object
}

type nilObject struct{}

type errorType string

type errorObject struct {
	errorType errorType
	message   string
}

func (x intObject) DebugTypeInfo() string {
	return "int"
}

func (x boolObject) DebugTypeInfo() string {
	return "bool"
}

func (x returnObject) DebugTypeInfo() string {
	return "return"
}

func (x nilObject) DebugTypeInfo() string {
	return "nil"
}

func (x errorObject) DebugTypeInfo() string {
	return "error"
}

func (x errorObject) String() string {
	return fmt.Sprintf("%v: %v", x.errorType, x.message)
}

func newErrorObject(errorType errorType, format string, a ...any) *errorObject {
	return &errorObject{
		errorType: errorType,
		message:   fmt.Sprintf(format, a...),
	}
}

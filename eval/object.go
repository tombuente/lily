package eval

import "fmt"

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

type internalError struct {
	msg string
	err error
}

type typeError struct {
	msg string
}

func (x *intObject) DebugTypeInfo() string {
	return "int"
}

func (x *boolObject) DebugTypeInfo() string {
	return "bool"
}

func (x *returnObject) DebugTypeInfo() string {
	return "return"
}

func (x *nilObject) DebugTypeInfo() string {
	return "nil"
}

func (x *internalError) Error() string {
	if x.err != nil {
		return fmt.Sprintf("%v", x.msg)
	}
	return fmt.Sprintf("%v: %v", x.msg, x.err)
}

func (x *typeError) Error() string {
	return fmt.Sprintf("%v", x.msg)
}

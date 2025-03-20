package eval

import (
	"fmt"

	"github.com/tombuente/lily/ast"
)

type object interface {
	Info() string
}

type builtinFunc func(args ...object) (object, error)

type environment struct {
	local    map[string]object
	captured *environment
}

type intObject struct {
	value int64
}

type boolObject struct {
	value bool
}

type stringObject struct {
	value string
}

// Used for type assertion to return early
type returnObject struct {
	value object
}

type functionObject struct {
	params   []*ast.Ident
	body     *ast.BlockStmt
	captured *environment
}

type builtinFunctionObject struct {
	fn builtinFunc
}

type nilObject struct{}

type internalError struct {
	msg string
	err error
}

type typeError struct {
	msg string
}

type nameError struct {
	msg string
}

func NewEnvironment() *environment {
	return &environment{local: make(map[string]object)}
}

func (env *environment) get(key string) (object, bool) {
	obj, ok := env.local[key]
	if !ok && env.captured != nil {
		obj, ok = env.captured.get(key)
	}
	return obj, ok
}

func (env *environment) set(key string, obj object) {
	if env.captured != nil {
		if _, ok := env.captured.get(key); ok {
			env.captured.update(key, obj)
		}
	}
	env.local[key] = obj
}

func (env *environment) update(key string, obj object) error {
	if _, ok := env.local[key]; ok {
		env.set(key, obj)
		return nil
	}

	if env.captured != nil {
		if _, ok := env.captured.get(key); ok {
			env.captured.set(key, obj)
			return nil
		}
	}

	return &nameError{msg: fmt.Sprintf("'%v' is not defined", key)}
}

func (x *intObject) Info() string {
	return "int"
}

func (x *boolObject) Info() string {
	return "bool"
}

func (x *stringObject) Info() string {
	return "string"
}

func (x *returnObject) Info() string {
	return "return"
}

func (x *functionObject) Info() string {
	return "function"
}

func (x *builtinFunctionObject) Info() string {
	return "buildin function"
}

func (x *nilObject) Info() string {
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

func (x *nameError) Error() string {
	return fmt.Sprintf("%v", x.msg)
}

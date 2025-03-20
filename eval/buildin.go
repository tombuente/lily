package eval

import "fmt"

var builtin = map[string]*builtinFunctionObject{
	"len": {fn: lenBuildin},
}

func lenBuildin(args ...object) (object, error) {
	if len(args) != 1 {
		return nil, &internalError{msg: fmt.Sprintf("len accepts 1 argument, got=%v", len(args))}
	}

	switch arg := args[0].(type) {
	case *stringObject:
		return &intObject{value: int64(len(arg.value))}, nil
	}
	return nil, &internalError{msg: fmt.Sprintf("arg not supported for len, got=%v", args[0].Info())}
}

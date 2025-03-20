package eval

import (
	"fmt"

	"github.com/tombuente/lily/ast"
)

var (
	nilInstance   = &nilObject{}
	trueInstance  = &boolObject{value: true}
	falseInstance = &boolObject{value: false}
)

func Eval(node ast.Node) (object, error) {
	return eval(node, NewEnvironment())
}

func eval(node ast.Node, env *environment) (object, error) {
	switch node := node.(type) {
	case *ast.Int:
		return evalIntExpr(node)
	case *ast.Bool:
		return evalBoolExpr(node)
	case *ast.String:
		return evalString(node)
	case *ast.UnaryOp:
		return evalUnaryExpr(node, env)
	case *ast.If:
		return evalIfExpr(node, env)
	case *ast.BinaryOp:
		return evalBinaryExpr(node, env)
	case *ast.Ident:
		return evalIdentExpr(node, env)
	case *ast.Function:
		return evalFunctionExpr(node, env)
	case *ast.Call:
		return evalCallExpr(node, env)
	case *ast.Assignment:
		return evalAssignmentExpr(node, env)
	case *ast.ExprStmt:
		return evalExprStmt(node, env)
	case *ast.ReturnStmt:
		return evalReturnStmt(node, env)
	case *ast.LetStmt:
		return evalLetStmt(node, env)
	case *ast.BlockStmt:
		return evalBlockStmt(node, env)
	case *ast.Program:
		return evalProgram(node, env)
	}
	return nil, &internalError{msg: "node not supported"}
}

func evalIntExpr(expr *ast.Int) (object, error) {
	return &intObject{value: expr.Value}, nil
}

func evalBoolExpr(expr *ast.Bool) (object, error) {
	return boolInstance(expr.Value), nil
}

func evalString(node *ast.String) (object, error) {
	return &stringObject{value: node.Value}, nil
}

func evalUnaryExpr(expr *ast.UnaryOp, env *environment) (object, error) {
	obj, err := eval(expr.Rhs, env)
	if err != nil {
		return nil, err
	}

	switch expr.Op {
	case "-":
		return evalUnaryMinusExpr(obj)
	case "!":
		return evalUnaryBangExpr(obj)
	}

	return nil, &internalError{msg: fmt.Sprintf("operator '%v' not implemented for unary expression", expr.Op)}
}

func evalUnaryMinusExpr(obj object) (object, error) {
	intObj, ok := obj.(*intObject)
	if !ok {
		return nil, &typeError{msg: fmt.Sprintf("bad operand type for unary -: '%v'", obj.Info())}
	}

	return &intObject{value: -intObj.value}, nil
}

func evalUnaryBangExpr(obj object) (object, error) {
	switch obj {
	case trueInstance:
		return falseInstance, nil
	case falseInstance:
		return trueInstance, nil
	}

	return nil, &typeError{msg: fmt.Sprintf("bad operand type for unary !: '%v'", obj.Info())}
}

func evalIfExpr(expr *ast.If, env *environment) (object, error) {
	conditionRes, err := eval(expr.Condition, env)
	if err != nil {
		return nil, err
	}

	condition, ok := conditionRes.(*boolObject)
	if !ok {
		return nil, &typeError{msg: fmt.Sprintf("if condition must evaluate to bool: '%v'", conditionRes.Info())}
	}

	if condition.value {
		return eval(expr.Consequence, env)
	} else if expr.Alternative != nil {
		return eval(expr.Alternative, env)
	}
	return nilInstance, nil
}

func evalIdentExpr(node *ast.Ident, env *environment) (object, error) {
	obj, ok := env.get(node.Value)
	if ok {
		return obj, nil
	}

	if buildin, ok := builtin[node.Value]; ok {
		return buildin, nil
	}

	return nil, &nameError{msg: fmt.Sprintf("name '%v' not defined", node.Value)}
}

func evalFunctionExpr(node *ast.Function, env *environment) (object, error) {
	return &functionObject{
		params:   node.Params,
		body:     node.Body,
		captured: env,
	}, nil
}

func evalCallExpr(node *ast.Call, env *environment) (object, error) {
	fn, err := eval(node.Lhs, env)
	if err != nil {
		return nil, err
	}

	args, err := evalExpressions(node.Args, env)
	if err != nil {
		return nil, err
	}

	return applyFunction(fn, args)
}

func applyFunction(fn object, args []object) (object, error) {
	switch fn := fn.(type) {
	case *functionObject:
		localEnv := NewEnvironment()
		for i, param := range fn.params {
			localEnv.set(param.Value, args[i])
		}
		localEnv.captured = fn.captured

		obj, err := eval(fn.body, localEnv)
		if err != nil {
			return nil, err
		}

		if retObj, ok := obj.(*returnObject); ok {
			return retObj.value, nil
		}
		return obj, nil
	case *builtinFunctionObject:
		return fn.fn(args...)
	}
	return nil, &internalError{msg: "function cannot be applied"}
}

func evalBinaryExpr(expr *ast.BinaryOp, env *environment) (object, error) {
	left, err := eval(expr.Left, env)
	if err != nil {
		return nil, err
	}

	right, err := eval(expr.Right, env)
	if err != nil {
		return nil, err
	}

	leftInt, leftOk := left.(*intObject)
	rightInt, rightOk := right.(*intObject)
	if leftOk && rightOk {
		return evalBinaryIntExpr(expr.Op, leftInt, rightInt)
	}

	leftBool, leftOk := left.(*boolObject)
	rightBool, rightOk := right.(*boolObject)
	if leftOk && rightOk {
		return evalBinaryBoolExpr(expr.Op, leftBool, rightBool)
	}

	leftString, leftOk := left.(*stringObject)
	rightString, rightOk := right.(*stringObject)
	if leftOk && rightOk {
		return evalBinaryStringExpr(expr.Op, leftString, rightString)
	}

	return nil, &typeError{msg: fmt.Sprintf("unsupported operand type(s) for '%v': '%v' '%v'", expr.Op, left.Info(), right.Info())}
}

func evalBinaryIntExpr(op string, left, right *intObject) (object, error) {
	switch op {
	case "+":
		return &intObject{value: left.value + right.value}, nil
	case "-":
		return &intObject{value: left.value - right.value}, nil
	case "*":
		return &intObject{value: left.value * right.value}, nil
	case "/":
		return &intObject{value: left.value / right.value}, nil
	case "<":
		return boolInstance(left.value < right.value), nil
	case ">":
		return boolInstance(left.value > right.value), nil
	case "==":
		return boolInstance(left.value == right.value), nil
	case "!=":
		return boolInstance(left.value != right.value), nil
	}
	return nil, &typeError{msg: fmt.Sprintf("unsupported operand type(s) for '%v': '%v' '%v'", op, left.Info(), right.Info())}
}

func evalBinaryBoolExpr(op string, left, right *boolObject) (object, error) {
	switch op {
	case "==":
		return boolInstance(left.value == right.value), nil
	case "!=":
		return boolInstance(left.value != right.value), nil
	}
	return nil, &typeError{msg: fmt.Sprintf("unsupported operand type(s) for '%v': '%v' '%v'", op, left.Info(), right.Info())}
}

func evalBinaryStringExpr(op string, left *stringObject, right *stringObject) (object, error) {
	switch op {
	case "+":
		fmt.Println(left.value, right.value)
		return &stringObject{value: left.value + right.value}, nil
	}
	return nil, &typeError{msg: fmt.Sprintf("unsupported operand type(s) for '%v': '%v' '%v'", op, left.Info(), right.Info())}
}

func evalAssignmentExpr(node *ast.Assignment, env *environment) (object, error) {
	val, err := eval(node.Expr, env)
	if err != nil {
		return nil, fmt.Errorf("cannot eval rhs: %w", err)
	}
	if err := env.update(node.Ident.Value, val); err != nil {
		return nil, err
	}
	return nilInstance, nil
}

func evalExprStmt(stmt *ast.ExprStmt, env *environment) (object, error) {
	return eval(stmt.Expr, env)
}

func evalLetStmt(node *ast.LetStmt, env *environment) (object, error) {
	if _, ok := env.get(node.Ident.Value); ok {
		return nil, &nameError{msg: fmt.Sprintf("'%v' already defined", node.Ident.Value)}
	}

	val, err := eval(node.Expr, env)
	if err != nil {
		return nil, err
	}
	env.set(node.Ident.Value, val)
	return nilInstance, nil
}

func evalReturnStmt(stmt *ast.ReturnStmt, env *environment) (object, error) {
	obj, err := eval(stmt.Expr, env)
	if err != nil {
		return nil, err
	}
	return &returnObject{value: obj}, nil
}

func evalBlockStmt(blockStmt *ast.BlockStmt, env *environment) (object, error) {
	return evalStmts(blockStmt.Stmts, env, false)
}

func evalProgram(prog *ast.Program, env *environment) (object, error) {
	return evalStmts(prog.Stmts, env, true)
}

func evalStmts(stmts []ast.Stmt, env *environment, unwrap bool) (object, error) {
	var obj object
	var err error
	for _, statement := range stmts {
		obj, err = eval(statement, env)
		if err != nil {
			return nil, err
		}

		if retObj, ok := obj.(*returnObject); ok {
			if unwrap {
				return retObj.value, nil
			}
			return retObj, nil
		}
	}
	return obj, nil
}

func boolInstance(val bool) object {
	if val {
		return trueInstance
	}
	return falseInstance
}

func evalExpressions(exprs []ast.Expr, env *environment) ([]object, error) {
	var objs []object
	for _, e := range exprs {
		val, err := eval(e, env)
		if err != nil {
			return nil, err
		}
		objs = append(objs, val)
	}
	return objs, nil
}

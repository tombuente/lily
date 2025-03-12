package eval

import (
	"fmt"

	"github.com/tombuente/lily/ast"
)

var (
	nilInstance   = &nilObject{}
	trueInstance  = &boolObject{Value: true}
	falseInstance = &boolObject{Value: false}
)

func Eval(node any) (object, error) {
	switch node := node.(type) {
	case *ast.IntExpr:
		return evalIntExpr(node)
	case *ast.BoolExpr:
		return evalBoolExpr(node)
	case *ast.UnaryExpr:
		return evalUnaryExpr(node)
	case *ast.IfExpr:
		return evalIfExpr(node)
	case *ast.BinaryExpr:
		return evalBinaryExpr(node)
	case *ast.ExprStmt:
		return evalExprStmt(node)
	case *ast.ReturnStmt:
		return evalReturnStmt(node)
	case *ast.BlockStmt:
		return evalBlockStmt(node)
	case *ast.Program:
		return evalProgram(node)
	}
	return nil, &internalError{msg: "node not supported"}
}

func evalIntExpr(expr *ast.IntExpr) (object, error) {
	return &intObject{Value: expr.Value}, nil
}

func evalBoolExpr(expr *ast.BoolExpr) (object, error) {
	return boolInstance(expr.Value), nil
}

func evalUnaryExpr(expr *ast.UnaryExpr) (object, error) {
	obj, err := Eval(expr.Expr)
	if err != nil {
		return nil, err
	}

	switch expr.Op {
	case "-":
		return evalUnaryMinusExpr(obj)
	case "!":
		return evalUnaryBangExpr(obj)
	}

	// not reachable if parser works
	return nil, &internalError{msg: fmt.Sprintf("operator '%v' not implemented for unary expression", expr.Op)}
}

func evalUnaryMinusExpr(obj object) (object, error) {
	intObj, ok := obj.(*intObject)
	if !ok {
		return nil, &typeError{msg: fmt.Sprintf("bad operand type for unary -: '%v'", obj.DebugTypeInfo())}
	}

	return &intObject{Value: -intObj.Value}, nil
}

func evalUnaryBangExpr(obj object) (object, error) {
	switch obj {
	case trueInstance:
		return falseInstance, nil
	case falseInstance:
		return trueInstance, nil
	}

	return nil, &typeError{msg: fmt.Sprintf("bad operand type for unary !: '%v'", obj.DebugTypeInfo())}
}

func evalIfExpr(expr *ast.IfExpr) (object, error) {
	conditionRes, err := Eval(expr.Condition)
	if err != nil {
		return nil, err
	}

	condition, ok := conditionRes.(*boolObject)
	if !ok {
		return nil, &typeError{msg: fmt.Sprintf("if condition must evaluate to bool: '%v'", conditionRes.DebugTypeInfo())}
	}

	if condition.Value {
		return Eval(expr.Consequence)
	} else if expr.Alternative != nil {
		return Eval(expr.Alternative)
	}
	return nilInstance, nil
}

func evalBinaryExpr(expr *ast.BinaryExpr) (object, error) {
	left, err := Eval(expr.Left)
	if err != nil {
		return nil, err
	}

	right, err := Eval(expr.Right)
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

	return nil, &typeError{msg: fmt.Sprintf("unsupported operand type(s) for '%v': '%v' '%v'", expr.Op, left.DebugTypeInfo(), right.DebugTypeInfo())}
}

func evalBinaryIntExpr(op string, left, right *intObject) (object, error) {
	switch op {
	case "+":
		return &intObject{Value: left.Value + right.Value}, nil
	case "-":
		return &intObject{Value: left.Value - right.Value}, nil
	case "*":
		return &intObject{Value: left.Value * right.Value}, nil
	case "/":
		return &intObject{Value: left.Value / right.Value}, nil
	case "<":
		return boolInstance(left.Value < right.Value), nil
	case ">":
		return boolInstance(left.Value > right.Value), nil
	case "==":
		return boolInstance(left.Value == right.Value), nil
	case "!=":
		return boolInstance(left.Value != right.Value), nil
	}
	return nil, &typeError{msg: fmt.Sprintf("unsupported operand type(s) for '%v': '%v' '%v'", op, left.DebugTypeInfo(), right.DebugTypeInfo())}
}

func evalBinaryBoolExpr(op string, left, right *boolObject) (object, error) {
	switch op {
	case "==":
		return boolInstance(left.Value == right.Value), nil
	case "!=":
		return boolInstance(left.Value != right.Value), nil
	}
	return nil, &typeError{msg: fmt.Sprintf("unsupported operand type(s) for '%v': '%v' '%v'", op, left.DebugTypeInfo(), right.DebugTypeInfo())}
}

func evalExprStmt(stmt *ast.ExprStmt) (object, error) {
	return Eval(stmt.Expr)
}

func evalReturnStmt(stmt *ast.ReturnStmt) (object, error) {
	obj, err := Eval(stmt.Expr)
	if err != nil {
		return nil, err
	}

	return &returnObject{Value: obj}, nil
}

func evalBlockStmt(blockStmt *ast.BlockStmt) (object, error) {
	return evalStmts(blockStmt.Stmts, false)
}

func evalProgram(prog *ast.Program) (object, error) {
	return evalStmts(prog.Stmts, true)
}

func evalStmts(stmts []ast.Stmt, unwrap bool) (object, error) {
	var obj object
	var err error
	for _, statement := range stmts {
		obj, err = Eval(statement)
		if err != nil {
			return nil, err
		}

		if retObj, ok := obj.(*returnObject); ok {
			if unwrap {
				return retObj.Value, nil
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

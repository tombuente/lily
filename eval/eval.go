package eval

import (
	"github.com/tombuente/lily/ast"
)

var (
	nilInstance   = &nilObject{}
	trueInstance  = &boolObject{Value: true}
	falseInstance = &boolObject{Value: false}
)

func Eval(node any) object {
	switch node := node.(type) {
	case *ast.IntExpr:
		return &intObject{Value: node.Value}
	case *ast.BoolExpr:
		return evalBoolExpr(node)
	case *ast.UnaryExpr:
		return evalUnaryExpr(node)
	case *ast.IfExpr:
		return evalIfExpr(node)
	case *ast.BinaryExpr:
		return evalBinaryExpr(node)
	case *ast.ExprStmt:
		return Eval(node.Expr)
	case *ast.ReturnStmt:
		return evalReturnStmt(node)
	case *ast.BlockStmt:
		return evalBlockStmt(node)
	case *ast.Program:
		return evalProgram(node)
	}
	return nilInstance
}

func evalBoolExpr(expr *ast.BoolExpr) object {
	return boolInstance(expr.Value)
}

func evalUnaryExpr(expr *ast.UnaryExpr) object {
	obj := Eval(expr.Expr)

	switch expr.Op {
	case "-":
		return evalUnaryMinusExpr(obj)
	case "!":
		return evalUnaryBangExpr(obj)
	}

	return nilInstance
}

func evalUnaryMinusExpr(obj object) object {
	intObj, ok := obj.(*intObject)
	if !ok {
		return newErrorObject(typeError, "bad operand type for unary -: \"%v\"", obj.DebugTypeInfo())
	}

	return &intObject{Value: -intObj.Value}
}

func evalUnaryBangExpr(obj object) object {
	switch obj {
	case trueInstance:
		return falseInstance
	case falseInstance:
		return trueInstance
	}
	return newErrorObject(typeError, "bad operand type for unary !: \"%v\"", obj.DebugTypeInfo())
}

func evalIfExpr(expr *ast.IfExpr) object {
	res := Eval(expr.Condition)
	condition, ok := res.(*boolObject)
	if !ok {
		return newErrorObject(typeError, "if condition must evaluate to bool: \"%v\"", res.DebugTypeInfo())
	}

	if condition.Value {
		return Eval(expr.Consequence)
	} else if expr.Alternative != nil {
		return Eval(expr.Alternative)
	}
	return nilInstance
}

func evalBinaryExpr(expr *ast.BinaryExpr) object {
	left := Eval(expr.Left)
	right := Eval(expr.Right)

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

	return nilInstance
}

func evalBinaryIntExpr(op string, left, right *intObject) object {
	switch op {
	case "+":
		return &intObject{Value: left.Value + right.Value}
	case "-":
		return &intObject{Value: left.Value - right.Value}
	case "*":
		return &intObject{Value: left.Value * right.Value}
	case "/":
		return &intObject{Value: left.Value / right.Value}
	case "<":
		return boolInstance(left.Value < right.Value)
	case ">":
		return boolInstance(left.Value > right.Value)
	case "==":
		return boolInstance(left.Value == right.Value)
	case "!=":
		return boolInstance(left.Value != right.Value)
	}
	panic("evalBinaryIntExpr: op not implemented")
}

func evalBinaryBoolExpr(op string, left, right *boolObject) object {
	switch op {
	case "==":
		return boolInstance(left.Value == right.Value)
	case "!=":
		return boolInstance(left.Value != right.Value)
	}
	panic("evalBinaryBoolExpr: op not implemented")
}

func evalReturnStmt(stmt *ast.ReturnStmt) object {
	val := Eval(stmt.Expr)
	return &returnObject{Value: val}
}

func evalBlockStmt(blockStmt *ast.BlockStmt) object {
	return evalStmts(blockStmt.Stmts, false)
}

func evalProgram(prog *ast.Program) object {
	return evalStmts(prog.Stmts, true)
}

func evalStmts(stmts []ast.Stmt, unwrap bool) object {
	var res object
	for _, statement := range stmts {
		res = Eval(statement)

		if retObj, ok := res.(*returnObject); ok {
			if unwrap {
				return retObj.Value
			}
			return retObj
		}
	}
	return res
}

func boolInstance(val bool) object {
	if val {
		return trueInstance
	}
	return falseInstance
}

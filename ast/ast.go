package ast

import (
	"encoding/json"
	"reflect"
)

type Expr interface {
	expr()
}

type Stmt interface {
	stmt()
}

type Program struct {
	Stmts []Stmt `json:"statements"`
}

type IdentExpr struct {
	Value string `json:"value"`
}

type IntExpr struct {
	Value int64 `json:"value"`
}

type BoolExpr struct {
	Value bool `json:"value"`
}

type PrefixExpr struct {
	Op    string `json:"operator"`
	Value Expr   `json:"value"`
}

type InfixExpr struct {
	Op    string `json:"operator"`
	Left  Expr   `json:"left"`
	Right Expr   `json:"right"`
}

func (x *IdentExpr) expr()  {}
func (x *IntExpr) expr()    {}
func (x *BoolExpr) expr()   {}
func (x *PrefixExpr) expr() {}
func (x *InfixExpr) expr()  {}

func (x IdentExpr) MarshalJSON() ([]byte, error) {
	return addType(x, "identifier_expression")
}

func (x IntExpr) MarshalJSON() ([]byte, error) {
	return addType(x, "int_expression")
}

func (x BoolExpr) MarshalJSON() ([]byte, error) {
	return addType(x, "bool_expression")
}

func (x PrefixExpr) MarshalJSON() ([]byte, error) {
	return addType(x, "prefix_expression")
}

func (x InfixExpr) MarshalJSON() ([]byte, error) {
	return addType(x, "infix_expression")
}

type LetStmt struct {
	Ident *IdentExpr `json:"identifier"`
	Value Expr       `json:"value"`
}

type ReturnStmt struct {
	Value Expr `json:"value"`
}

type ExprStmt struct {
	Expr Expr `json:"expression"`
}

func (x *LetStmt) stmt()    {}
func (x *ReturnStmt) stmt() {}
func (x *ExprStmt) stmt()   {}

func (x LetStmt) MarshalJSON() ([]byte, error) {
	return addType(x, "let_statement")
}

func (x ReturnStmt) MarshalJSON() ([]byte, error) {
	return addType(x, "return_statement")
}

func (x ExprStmt) MarshalJSON() ([]byte, error) {
	return addType(x, "expression_statement")
}

func (x Program) MarshalJSON() ([]byte, error) {
	return addType(x, "program")
}

func addType(stmt any, typ string) ([]byte, error) {
	value := reflect.ValueOf(stmt)
	valueType := reflect.TypeOf(stmt)

	fields := []reflect.StructField{
		{
			Name: "Type",
			Type: reflect.TypeOf(""),
			Tag:  reflect.StructTag(`json:"type"`),
		},
	}
	for i := range valueType.NumField() {
		fields = append(fields, valueType.Field(i))
	}
	newType := reflect.StructOf(fields)

	newValue := reflect.New(newType).Elem()
	newValue.FieldByName("Type").SetString(typ)
	for i := range valueType.NumField() {
		newValue.Field(i + 1).Set(value.Field(i))
	}

	return json.Marshal(newValue.Interface())
}

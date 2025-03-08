package ast

import (
	"bytes"
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

type UnaryExpr struct {
	Op   string `json:"operator"`
	Expr Expr   `json:"value"`
}

type BinaryExpr struct {
	Op    string `json:"operator"`
	Left  Expr   `json:"left"`
	Right Expr   `json:"right"`
}

type IfExpr struct {
	Condition   Expr
	Consequence *BlockStmt
	Alternative *BlockStmt
}

type FnExpr struct {
	Params []*IdentExpr `json:"params"`
	Body   *BlockStmt   `json:"body"`
}

type CallExpr struct {
	Fn   Expr // [IdentExpr] or [FnExpr]
	Args []Expr
}

func (x *IdentExpr) expr()  {}
func (x *IntExpr) expr()    {}
func (x *BoolExpr) expr()   {}
func (x *UnaryExpr) expr()  {}
func (x *BinaryExpr) expr() {}
func (x *IfExpr) expr()     {}
func (x *FnExpr) expr()     {}
func (x *CallExpr) expr()   {}

func (x IdentExpr) MarshalJSON() ([]byte, error) {
	return addType(x, "identifier_expression")
}

func (x IntExpr) MarshalJSON() ([]byte, error) {
	return addType(x, "int_expression")
}

func (x BoolExpr) MarshalJSON() ([]byte, error) {
	return addType(x, "bool_expression")
}

func (x UnaryExpr) MarshalJSON() ([]byte, error) {
	return addType(x, "unary_expression")
}

func (x BinaryExpr) MarshalJSON() ([]byte, error) {
	return addType(x, "binary_expression")
}

func (x IfExpr) MarshalJSON() ([]byte, error) {
	return addType(x, "if_expression")
}

func (x FnExpr) MarshalJSON() ([]byte, error) {
	return addType(x, "function_expression")
}

func (x CallExpr) MarshalJSON() ([]byte, error) {
	return addType(x, "call_expression")
}

type LetStmt struct {
	Ident *IdentExpr `json:"identifier"`
	Expr  Expr       `json:"value"`
}

type ReturnStmt struct {
	Expr Expr `json:"value"`
}

type ExprStmt struct {
	Expr Expr `json:"expression"`
}

type BlockStmt struct {
	Stmts []Stmt `json:"statements"`
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

func (x BlockStmt) MarshalJSON() ([]byte, error) {
	return addType(x, "block_statement")
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

	return marshal(newValue.Interface())
}

func marshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}

	bytes := buf.Bytes()
	return bytes, nil
}

func MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent(prefix, indent)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}

	bytes := buf.Bytes()
	return bytes, nil
}

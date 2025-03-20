package ast

import (
	"bytes"
	"encoding/json"
	"reflect"
)

type Node interface {
	node()
}

type Expr interface {
	Node
	expr()
}

type Stmt interface {
	Node
	stmt()
}

type Program struct {
	Stmts []Stmt `json:"statements"`
}

func (x *Program) node() {}

type Ident struct {
	Value string `json:"value"`
}

type Int struct {
	Value int64 `json:"value"`
}

type Bool struct {
	Value bool `json:"value"`
}

type String struct {
	Value string `json:"value"`
}

type UnaryOp struct {
	Op  string `json:"operator"`
	Rhs Expr   `json:"value"`
}

type BinaryOp struct {
	Op    string `json:"operator"`
	Left  Expr   `json:"left"`
	Right Expr   `json:"right"`
}

type If struct {
	Condition   Expr
	Consequence *BlockStmt
	Alternative *BlockStmt
}

type Function struct {
	Params []*Ident   `json:"params"`
	Body   *BlockStmt `json:"body"`
}

type Call struct {
	Lhs  Expr // Ident or Function
	Args []Expr
}

type Assignment struct {
	Ident *Ident
	Expr  Expr
}

func (x *Ident) expr()      {}
func (x *Int) expr()        {}
func (x *Bool) expr()       {}
func (x *String) expr()     {}
func (x *UnaryOp) expr()    {}
func (x *BinaryOp) expr()   {}
func (x *If) expr()         {}
func (x *Function) expr()   {}
func (x *Call) expr()       {}
func (x *Assignment) expr() {}

func (x *Ident) node()      {}
func (x *Int) node()        {}
func (x *Bool) node()       {}
func (x *String) node()     {}
func (x *UnaryOp) node()    {}
func (x *BinaryOp) node()   {}
func (x *If) node()         {}
func (x *Function) node()   {}
func (x *Call) node()       {}
func (x *Assignment) node() {}

func (x Ident) MarshalJSON() ([]byte, error) {
	return addType(x, "identifier_expression")
}

func (x Int) MarshalJSON() ([]byte, error) {
	return addType(x, "int_expression")
}

func (x Bool) MarshalJSON() ([]byte, error) {
	return addType(x, "bool_expression")
}

func (x String) MarshalJSON() ([]byte, error) {
	return addType(x, "string_expression")
}

func (x UnaryOp) MarshalJSON() ([]byte, error) {
	return addType(x, "unary_expression")
}

func (x BinaryOp) MarshalJSON() ([]byte, error) {
	return addType(x, "binary_expression")
}

func (x If) MarshalJSON() ([]byte, error) {
	return addType(x, "if_expression")
}

func (x Function) MarshalJSON() ([]byte, error) {
	return addType(x, "function_expression")
}

func (x Call) MarshalJSON() ([]byte, error) {
	return addType(x, "call_expression")
}

func (x Assignment) MarshalJSON() ([]byte, error) {
	return addType(x, "assignment_expression")
}

type LetStmt struct {
	Ident *Ident `json:"identifier"`
	Expr  Expr   `json:"value"`
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
func (x *BlockStmt) stmt()  {}

func (x *LetStmt) node()    {}
func (x *ReturnStmt) node() {}
func (x *ExprStmt) node()   {}
func (x *BlockStmt) node()  {}

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

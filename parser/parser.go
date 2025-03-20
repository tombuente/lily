package parser

import (
	"fmt"
	"strconv"

	"github.com/tombuente/lily/ast"
	"github.com/tombuente/lily/token"
)

const (
	none int = iota // no precedence
	assign
	eq     // ==
	less   // < or >
	sum    // + or -
	mul    // * or /
	prefix // !
	call   // grouped expr or function call
)

var precedences = map[token.Type]int{
	token.Assign:   assign,
	token.EQ:       eq,
	token.NotEQ:    eq,
	token.Less:     less,
	token.Greater:  less,
	token.Plus:     sum,
	token.Minus:    sum,
	token.Asterisk: mul,
	token.Slash:    mul,
	token.LParan:   call,
}

type (
	prefixParseFn func() (ast.Expr, error)
	infixParseFn  func(left ast.Expr) (ast.Expr, error)
)

type Lexer interface {
	Next() token.Token
}

type Parser struct {
	l Lexer

	tok token.Token

	prefixParseFns map[token.Type]prefixParseFn
	infixParseFns  map[token.Type]infixParseFn
}

func New(lexer Lexer) *Parser {
	p := &Parser{l: lexer}

	p.prefixParseFns = make(map[token.Type]prefixParseFn)
	p.prefixParseFns[token.Ident] = p.parseIdent
	p.prefixParseFns[token.Int] = p.parseInt
	p.prefixParseFns[token.True] = p.parseBool
	p.prefixParseFns[token.False] = p.parseBool
	p.prefixParseFns[token.String] = p.parseString
	p.prefixParseFns[token.If] = p.parseIf
	p.prefixParseFns[token.LParan] = p.parseGroup
	p.prefixParseFns[token.Fn] = p.parseFunction
	p.prefixParseFns[token.Minus] = p.parseUnaryOp
	p.prefixParseFns[token.Bang] = p.parseUnaryOp

	p.infixParseFns = make(map[token.Type]infixParseFn)
	p.infixParseFns[token.Plus] = p.parseBinaryOp
	p.infixParseFns[token.Minus] = p.parseBinaryOp
	p.infixParseFns[token.Asterisk] = p.parseBinaryOp
	p.infixParseFns[token.Slash] = p.parseBinaryOp
	p.infixParseFns[token.EQ] = p.parseBinaryOp
	p.infixParseFns[token.NotEQ] = p.parseBinaryOp
	p.infixParseFns[token.Less] = p.parseBinaryOp
	p.infixParseFns[token.Greater] = p.parseBinaryOp
	p.infixParseFns[token.LParan] = p.parseCall
	p.infixParseFns[token.Assign] = p.parseAssingment

	p.next()

	return p
}

func (p *Parser) Parse() (*ast.Program, error) {
	stmts := []ast.Stmt{}
	for p.tok.Type != token.EOF {
		stmt, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
	}

	return &ast.Program{Stmts: stmts}, nil
}

func (p *Parser) parseExpr(prec int) (ast.Expr, error) {
	parsePrefix, ok := p.prefixParseFns[p.tok.Type]
	if !ok {
		return nil, fmt.Errorf("missing prefix parse function for '%v'", p.tok.Type)
	}
	lhs, err := parsePrefix()
	if err != nil {
		return nil, fmt.Errorf("failed to parse prefix: %w", err)
	}

	for prec < precedence(p.tok.Type) {
		parseInfix, ok := p.infixParseFns[p.tok.Type]
		if !ok {
			return nil, fmt.Errorf("missing infix parse function for '%v'", p.tok.Type)
		}

		lhs, err = parseInfix(lhs)
		if err != nil {
			return nil, fmt.Errorf("failed to parse infix: %w", err)
		}
	}

	return lhs, nil
}

func (p *Parser) parseIdent() (ast.Expr, error) {
	value := p.tok.Literal
	p.next()

	return &ast.Ident{
		Value: value,
	}, nil
}

func (p *Parser) parseInt() (ast.Expr, error) {
	value, err := strconv.ParseInt(p.tok.Literal, 0, 64)
	if err != nil {
		return nil, err
	}
	p.next()

	return &ast.Int{
		Value: value,
	}, nil
}

func (p *Parser) parseBool() (ast.Expr, error) {
	value := p.tok.Type == token.True
	p.next()

	return &ast.Bool{
		Value: value,
	}, nil
}

func (p *Parser) parseString() (ast.Expr, error) {
	value := p.tok.Literal
	p.next()

	return &ast.String{Value: value}, nil
}

// if <condition> { <consequence> } { <alternative> }
func (p *Parser) parseIf() (ast.Expr, error) {
	p.next()

	condition, err := p.parseExpr(none)
	if err != nil {
		return nil, fmt.Errorf("failed to parse if condition: %w", err)
	}

	consequence, err := p.parseBlockStmt()
	if err != nil {
		return nil, fmt.Errorf("failed to parse if consequence: %w", err)
	}

	var alternative *ast.BlockStmt
	if p.tok.Type == token.LBrace {
		alternative, err = p.parseBlockStmt()
		if err != nil {
			return nil, fmt.Errorf("failed to parse if alternative: %w", err)
		}
	}

	return &ast.If{
		Condition:   condition,
		Consequence: consequence,
		Alternative: alternative,
	}, nil
}

// (<ident> + <ident>)
func (p *Parser) parseGroup() (ast.Expr, error) {
	p.next()

	group, err := p.parseExpr(none)
	if err != nil {
		return nil, err
	}

	if err := p.expectNext(token.RParan); err != nil {
		return nil, fmt.Errorf("group must start with '%v': %w", token.RParan, err)
	}

	return group, nil
}

// fn(<ident>, <ident>) { <statement> }
func (p *Parser) parseFunction() (ast.Expr, error) {
	p.next() // consume fn

	params, err := p.parseFunctionParams()
	if err != nil {
		return nil, fmt.Errorf("failed to parse function parameter list: %w", err)
	}

	body, err := p.parseBlockStmt()
	if err != nil {
		return nil, fmt.Errorf("failed to parse function body: %w", err)
	}

	return &ast.Function{
		Params: params,
		Body:   body,
	}, nil
}

// (<ident>, <ident>)
func (p *Parser) parseFunctionParams() ([]*ast.Ident, error) {
	if err := p.expectNext(token.LParan); err != nil {
		return nil, fmt.Errorf("function parameters must start with '%v': %w", token.LParan, err)
	}

	idents := []*ast.Ident{}
	if p.tok.Type == token.RParan {
		p.next()
		return idents, nil
	}

	for p.tok.Type != token.RParan {
		ident := &ast.Ident{Value: p.tok.Literal}
		idents = append(idents, ident)
		p.next()

		// Consume the comma "," if present after the argument.
		// Comma will not be there if this was the last argument.
		if p.tok.Type == token.Comma {
			p.next()
		}
	}

	if err := p.expectNext(token.RParan); err != nil {
		return nil, fmt.Errorf("function parameters must end with '%v': %w", token.RParan, err)
	}

	return idents, nil
}

// -<ident>
// !<ident>
func (p *Parser) parseUnaryOp() (ast.Expr, error) {
	op := p.tok.Literal
	p.next()

	rhs, err := p.parseExpr(prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rhs: %w", err)
	}

	return &ast.UnaryOp{
		Op:  op,
		Rhs: rhs,
	}, nil
}

// <ident> + <ident>
// <ident> * <ident>
func (p *Parser) parseBinaryOp(left ast.Expr) (ast.Expr, error) {
	op := p.tok.Literal
	prec := precedence(p.tok.Type)
	p.next()

	rhs, err := p.parseExpr(prec)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rhs: %w", err)
	}

	return &ast.BinaryOp{
		Op:    op,
		Left:  left,
		Right: rhs,
	}, nil
}

// <lhs>(<ident>, <ident>, ..., <ident>)
// lhs is either [ast.Function] or [ast.Ident]
func (p *Parser) parseCall(lhs ast.Expr) (ast.Expr, error) {
	args, err := p.parseCallArgs()
	if err != nil {
		return nil, fmt.Errorf("failed to parse call args: %w", err)
	}

	return &ast.Call{
		Lhs:  lhs,
		Args: args,
	}, nil
}

func (p *Parser) parseCallArgs() ([]ast.Expr, error) {
	if err := p.expectNext(token.LParan); err != nil {
		return nil, fmt.Errorf("call args must start with '%v': %w", token.LParan, err)
	}

	args := []ast.Expr{}
	if p.tok.Type == token.RParan {
		p.next()
		return args, nil
	}

	for p.tok.Type != token.RParan {
		arg, err := p.parseExpr(none)
		if err != nil {
			return nil, fmt.Errorf("failed to parse arg of call: %w", err)
		}
		args = append(args, arg)

		// Consume the comma "," if present after the argument.
		// Comma will not be there if this was the last argument.
		if p.tok.Type == token.Comma {
			p.next()
		}
	}

	if err := p.expectNext(token.RParan); err != nil { // consume ")"
		return nil, fmt.Errorf("call args must end with '%v': %w", token.RParan, err)
	}

	return args, nil
}

func (p *Parser) parseAssingment(ident ast.Expr) (ast.Expr, error) {
	identExpr, ok := ident.(*ast.Ident)
	if !ok {
		return nil, fmt.Errorf("indet is not *ast.Ident")
	}

	if err := p.expectNext(token.Assign); err != nil {
		return nil, err
	}

	expr, err := p.parseExpr(none)
	if err != nil {
		return nil, err
	}

	return &ast.Assignment{
		Ident: identExpr,
		Expr:  expr,
	}, nil
}

func (p *Parser) parseStmt() (ast.Stmt, error) {
	switch p.tok.Type {
	case token.Let:
		return p.parseLetStmt()
	case token.Return:
		return p.parseReturnStmt()
	}
	return p.parseExprStmt()
}

// let <ident> = <expr>
func (p *Parser) parseLetStmt() (*ast.LetStmt, error) {
	p.next() // consume let

	if err := p.expect(token.Ident); err != nil {
		return nil, fmt.Errorf("expected an identifier: %w", err)
	}
	ident := &ast.Ident{Value: p.tok.Literal}
	p.next()

	if err := p.expectNext(token.Assign); err != nil {
		return nil, fmt.Errorf("expected assigment token: %w", err)
	}

	expr, err := p.parseExpr(none)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expression: %w", err)
	}

	if p.tok.Type == token.Semicolon {
		p.next()
	}

	return &ast.LetStmt{
		Ident: ident,
		Expr:  expr,
	}, nil
}

// return <expr>
func (p *Parser) parseReturnStmt() (*ast.ReturnStmt, error) {
	p.next() // consume "return"

	expr, err := p.parseExpr(none)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expression: %w", err)
	}

	if p.tok.Type == token.Semicolon {
		p.next()
	}

	return &ast.ReturnStmt{
		Expr: expr,
	}, nil
}

// <expr>
func (p *Parser) parseExprStmt() (*ast.ExprStmt, error) {
	expr, err := p.parseExpr(none)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expression: %w", err)
	}

	if p.tok.Type == token.Semicolon {
		p.next()
	}

	return &ast.ExprStmt{
		Expr: expr,
	}, nil
}

func (p *Parser) parseBlockStmt() (*ast.BlockStmt, error) {
	if err := p.expectNext(token.LBrace); err != nil {
		return nil, fmt.Errorf("block must start with '%v': %w", token.LBrace, err)
	}

	stmts := []ast.Stmt{}
	for p.tok.Type != token.RBrace && p.tok.Type != token.EOF {
		stmt, err := p.parseStmt()
		if err != nil {
			return nil, fmt.Errorf("failed to parse statement: %w", err)
		}
		stmts = append(stmts, stmt)
	}

	if err := p.expectNext(token.RBrace); err != nil {
		return nil, fmt.Errorf("block must stop with '%v': %w", token.RBrace, err)
	}

	if p.tok.Type == token.Semicolon {
		p.next()
	}

	return &ast.BlockStmt{
		Stmts: stmts,
	}, nil
}

func (p *Parser) next() {
	p.tok = p.l.Next()
}

// expect checks whether the current token matches the expected type.
// If not, it returns an error indicating the mismatch.
func (p *Parser) expect(typ token.Type) error {
	if p.tok.Type != typ {
		return fmt.Errorf("expected %v, got %v", typ, p.tok.Type)
	}
	return nil
}

// expectNext works like [expect] but also calls [Parser.next]
func (p *Parser) expectNext(typ token.Type) error {
	if err := p.expect(typ); err != nil {
		return err
	}
	p.next()
	return nil
}

func precedence(typ token.Type) int {
	if p, ok := precedences[typ]; ok {
		return p
	}
	return none
}

package parser

import (
	"fmt"
	"strconv"

	"github.com/tombuente/lily/ast"
	"github.com/tombuente/lily/token"
)

const (
	none   int = iota // no precedence
	eq                // ==
	less              // < or >
	sum               // + or -
	mul               // * or /
	prefix            // !
	call              // func call or grouped expr (1 + 2) * 3
)

var precedences = map[token.Type]int{
	token.Plus:     sum,
	token.Minus:    sum,
	token.Asterisk: mul,
	token.Slash:    mul,
	token.EQ:       eq,
	token.NotEQ:    eq,
	token.Less:     less,
	token.Greater:  less,
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
	p.prefixParseFns[token.Ident] = p.parseIdentExpr
	p.prefixParseFns[token.Int] = p.parseIntExpr
	p.prefixParseFns[token.True] = p.parseBoolExpr
	p.prefixParseFns[token.False] = p.parseBoolExpr
	p.prefixParseFns[token.If] = p.parseIfExpr
	p.prefixParseFns[token.LParan] = p.parseGroupedExpr
	p.prefixParseFns[token.Fn] = p.parseFnExpr
	p.prefixParseFns[token.Minus] = p.parseUnaryExpr
	p.prefixParseFns[token.Bang] = p.parseUnaryExpr

	p.infixParseFns = make(map[token.Type]infixParseFn)
	p.infixParseFns[token.Plus] = p.parseBinaryExpr
	p.infixParseFns[token.Minus] = p.parseBinaryExpr
	p.infixParseFns[token.Asterisk] = p.parseBinaryExpr
	p.infixParseFns[token.Slash] = p.parseBinaryExpr
	p.infixParseFns[token.EQ] = p.parseBinaryExpr
	p.infixParseFns[token.NotEQ] = p.parseBinaryExpr
	p.infixParseFns[token.Less] = p.parseBinaryExpr
	p.infixParseFns[token.Greater] = p.parseBinaryExpr
	p.infixParseFns[token.LParan] = p.parseCallExpr

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
		return nil, fmt.Errorf("missing prefix parse function for \"%v\"", p.tok.Type)
	}
	left, err := parsePrefix()
	if err != nil {
		return nil, fmt.Errorf("failed to parse prefix: %w", err)
	}

	// LBrace opens a new statement block
	for p.tok.Type != token.Semicolon && p.tok.Type != token.LBrace && prec < precedence(p.tok.Type) {
		parseInfix, ok := p.infixParseFns[p.tok.Type]
		if !ok {
			return nil, fmt.Errorf("missing infix parse function for \"%v\v", p.tok.Type)
		}

		left, err = parseInfix(left)
		if err != nil {
			return nil, fmt.Errorf("failed to parse infix: %w", err)
		}
	}

	return left, nil
}

func (p *Parser) parseIdentExpr() (ast.Expr, error) {
	if err := p.expect(token.Ident); err != nil {
		return nil, err
	}
	value := p.tok.Literal

	p.next()
	return &ast.IdentExpr{
		Value: value,
	}, nil
}

func (p *Parser) parseIntExpr() (ast.Expr, error) {
	if err := p.expect(token.Int); err != nil {
		return nil, err
	}
	value, err := strconv.ParseInt(p.tok.Literal, 0, 64)
	if err != nil {
		return nil, err
	}
	p.next()

	return &ast.IntExpr{
		Value: value,
	}, nil
}

func (p *Parser) parseBoolExpr() (ast.Expr, error) {
	// TODO: verify that p.tok is bool

	value := p.tok.Type == token.True
	p.next()

	return &ast.BoolExpr{
		Value: value,
	}, nil
}

// if <expr> { <expr> } { <expr> }
func (p *Parser) parseIfExpr() (ast.Expr, error) {
	if err := p.expectNext(token.If); err != nil { // consume "if"
		return nil, fmt.Errorf("if expression must start with \"%v\": %w", token.If, err)
	}

	condition, err := p.parseExpr(none)
	if err != nil {
		return nil, fmt.Errorf("failed to parse if expression condition expression: %w", err)
	}

	consequence, err := p.parseBlockStmt()
	if err != nil {
		return nil, fmt.Errorf("failed to parse if expression consequence: %w", err)
	}

	var alternative *ast.BlockStmt
	if p.tok.Type == token.LBrace {
		alternative, err = p.parseBlockStmt()
		if err != nil {
			return nil, fmt.Errorf("failed to parse if expression alternative: %w", err)
		}
	}

	return &ast.IfExpr{
		Condition:   condition,
		Consequence: consequence,
		Alternative: alternative,
	}, nil
}

// (<ident> + <ident>)
func (p *Parser) parseGroupedExpr() (ast.Expr, error) {
	if err := p.expectNext(token.LParan); err != nil { // consume "("
		return nil, fmt.Errorf("grouped expression must start with \"%v\v: %w", token.LParan, err)
	}

	expr, err := p.parseExpr(none)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expression: %w", err)
	}

	if err := p.expectNext(token.RParan); err != nil { // consume ")"
		return nil, fmt.Errorf("grouped expression must start with \"%v\v: %w", token.RParan, err)
	}

	return expr, nil
}

// fn(<ident>, <ident>) { <stmt> }
func (p *Parser) parseFnExpr() (ast.Expr, error) {
	if err := p.expectNext(token.Fn); err != nil {
		return nil, fmt.Errorf("function expression must start with \"%v\": %w", token.Fn, err)
	}

	params, err := p.parseFnExprParams()
	if err != nil {
		return nil, fmt.Errorf("failed to parse function expression parameter list: %w", err)
	}

	body, err := p.parseBlockStmt()
	if err != nil {
		return nil, fmt.Errorf("failed to parse function expression body statement: %w", err)
	}

	return &ast.FnExpr{
		Params: params,
		Body:   body,
	}, nil
}

// (<ident>, <ident>)
func (p *Parser) parseFnExprParams() ([]*ast.IdentExpr, error) {
	if err := p.expectNext(token.LParan); err != nil {
		return nil, fmt.Errorf("function expression parameters must start with \"%v\": %w", token.LParan, err)
	}

	idents := []*ast.IdentExpr{}

	if p.tok.Type == token.RParan {
		p.next()
		return idents, nil
	}

	ident := &ast.IdentExpr{Value: p.tok.Literal}
	idents = append(idents, ident)
	p.next()

	for p.tok.Type == token.Comma {
		p.next() // consume ","

		if err := p.expect(token.Ident); err != nil {
			return nil, err
		}

		ident := &ast.IdentExpr{Value: p.tok.Literal}
		idents = append(idents, ident)
		p.next()
	}

	if err := p.expectNext(token.RParan); err != nil {
		return nil, fmt.Errorf("function expression parameters must start with \"%v\": %w", token.RParan, err)
	}

	return idents, nil
}

// -<ident>
// !<ident>
func (p *Parser) parseUnaryExpr() (ast.Expr, error) {
	op := p.tok.Literal // TODO: verify that p.tok.Type is prefix
	p.next()

	expr, err := p.parseExpr(prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expression: %w", err)
	}

	return &ast.UnaryExpr{
		Op:   op,
		Expr: expr,
	}, nil
}

// <ident> + <ident>
// <ident> * <ident>
// ...
func (p *Parser) parseBinaryExpr(left ast.Expr) (ast.Expr, error) {
	op := p.tok.Literal

	prec := precedence(p.tok.Type)
	p.next()
	right, err := p.parseExpr(prec)
	if err != nil {
		return nil, fmt.Errorf("failed to parse right expression: %w", err)
	}

	return &ast.BinaryExpr{
		Op:    op,
		Left:  left,
		Right: right,
	}, nil
}

// <expr>(<expr>, <expr>, ..., <expr>)
func (p *Parser) parseCallExpr(fn ast.Expr) (ast.Expr, error) {
	args, err := p.parseCallExprArgs()
	if err != nil {
		return nil, fmt.Errorf("failed to parse call expression args: %w", err)
	}

	return &ast.CallExpr{
		Fn:   fn,
		Args: args,
	}, nil
}

func (p *Parser) parseCallExprArgs() ([]ast.Expr, error) {
	if err := p.expectNext(token.LParan); err != nil {
		return nil, fmt.Errorf("call expression args must start with \"%v\": %w", token.LParan, err)
	}

	args := []ast.Expr{}

	if p.tok.Type == token.RParan {
		p.next()
		return args, nil
	}

	arg, err := p.parseExpr(none)
	if err != nil {
		return nil, fmt.Errorf("failed to parse first arg of call expression: %w", err)
	}
	args = append(args, arg)

	for p.tok.Type == token.Comma {
		p.next() // consume ","

		arg, err := p.parseExpr(none)
		if err != nil {
			return nil, fmt.Errorf("failed to parse arg of call expression: %w", err)
		}
		args = append(args, arg)
	}

	if err := p.expectNext(token.RParan); err != nil {
		return nil, fmt.Errorf("call expression args must start with \"%v\": %w", token.RParan, err)
	}

	return args, nil
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
	if err := p.expectNext(token.Let); err != nil { // consume "let"
		return nil, err
	}

	if err := p.expect(token.Ident); err != nil {
		return nil, err
	}
	ident := &ast.IdentExpr{Value: p.tok.Literal}
	p.next()

	if err := p.expectNext(token.Assign); err != nil { // consume "="
		return nil, err
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
	if err := p.expectNext(token.Return); err != nil { // consume "return"
		return nil, err
	}

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
		return nil, fmt.Errorf("block statement must start with \"%v\": %w", token.LBrace, err)
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
		return nil, fmt.Errorf("block statement must stop with \"%v\": %w", token.RBrace, err)
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

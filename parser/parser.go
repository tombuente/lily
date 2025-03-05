package parser

import (
	"fmt"
	"strconv"

	"github.com/tombuente/lily/ast"
	"github.com/tombuente/lily/lexer"
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

type Parser struct {
	l       *lexer.Lexer
	currTok token.Token
	nextTok token.Token

	prefixParseFns map[token.Type]prefixParseFn
	infixParseFns  map[token.Type]infixParseFn
}

func New(lexer *lexer.Lexer) *Parser {
	p := &Parser{l: lexer}

	p.prefixParseFns = make(map[token.Type]prefixParseFn)
	p.prefixParseFns[token.Ident] = p.parseIdentExpr
	p.prefixParseFns[token.Int] = p.parseIntExpr
	p.prefixParseFns[token.Minus] = p.parsePrefixExpr
	p.prefixParseFns[token.True] = p.parseBoolExpr
	p.prefixParseFns[token.False] = p.parseBoolExpr
	p.prefixParseFns[token.ExclamationMark] = p.parsePrefixExpr
	p.prefixParseFns[token.LParan] = p.parseGroupedExpr

	p.infixParseFns = make(map[token.Type]infixParseFn)
	p.infixParseFns[token.Plus] = p.parseInfixExpr
	p.infixParseFns[token.Minus] = p.parseInfixExpr
	p.infixParseFns[token.Asterisk] = p.parseInfixExpr
	p.infixParseFns[token.Slash] = p.parseInfixExpr
	p.infixParseFns[token.EQ] = p.parseInfixExpr
	p.infixParseFns[token.NotEQ] = p.parseInfixExpr
	p.infixParseFns[token.Less] = p.parseInfixExpr
	p.infixParseFns[token.Greater] = p.parseInfixExpr

	p.next()
	p.next()

	return p
}

func (p *Parser) Parse() (ast.Program, error) {
	prog := ast.Program{
		Stmts: []ast.Stmt{},
	}

	for !p.currTokenIs(token.EOF) {
		stmt, err := p.parseStmt()
		if err != nil {
			return prog, err
		}
		prog.Stmts = append(prog.Stmts, stmt)

		// Consume semicolon after statement
		if p.nextTokenIs(token.Semicolon) {
			p.next()
		}

		p.next()
	}

	return prog, nil
}

func (p *Parser) parseExpr(precedence int) (ast.Expr, error) {
	parsePrefix, ok := p.prefixParseFns[p.currTok.Type]
	if !ok {
		return nil, fmt.Errorf("cannot find prefix parse function for %v", p.currTok)
	}
	left, err := parsePrefix()
	if err != nil {
		return nil, fmt.Errorf("failed to parse prefix: %w", err)
	}

	// Non-obvious break condition: The loop terminates when the next token is EOF,
	// because an EOF token is assigned a precedence of 0.
	for !p.nextTokenIs(token.Semicolon) && precedence < p.nextTokenPrecedence() {
		parseInfix, ok := p.infixParseFns[p.nextTok.Type]
		if !ok {
			return nil, fmt.Errorf("cannot find infix parse function for %v", p.nextTok)
		}

		p.next()
		left, err = parseInfix(left)
		if err != nil {
			return nil, fmt.Errorf("failed to parse infix: %w", err)
		}
	}

	return left, nil
}

func (p *Parser) parseIdentExpr() (ast.Expr, error) {
	return &ast.IdentExpr{
		Value: p.currTok.Literal,
	}, nil
}

func (p *Parser) parseIntExpr() (ast.Expr, error) {
	value, err := strconv.ParseInt(p.currTok.Literal, 0, 64)
	if err != nil {
		return nil, fmt.Errorf("failed not parse IntExpr: %w", err)
	}
	return &ast.IntExpr{
		Value: value,
	}, nil
}

func (p *Parser) parseBoolExpr() (ast.Expr, error) {
	return &ast.BoolExpr{
		Value: p.currTokenIs(token.True),
	}, nil
}

func (p *Parser) parsePrefixExpr() (ast.Expr, error) {
	expr := &ast.PrefixExpr{
		Op: p.currTok.Literal,
	}
	p.next()

	value, err := p.parseExpr(prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to parse prefix expr: %w", err)
	}
	expr.Value = value

	return expr, nil
}

func (p *Parser) parseInfixExpr(left ast.Expr) (ast.Expr, error) {
	op := p.currTok.Literal

	precedence := p.currTokenPrecedence()
	p.next()
	right, err := p.parseExpr(precedence)
	if err != nil {
		return nil, fmt.Errorf("failed to parse right expr: %w", err)
	}

	return &ast.InfixExpr{
		Op:    op,
		Left:  left,
		Right: right,
	}, nil
}

func (p *Parser) parseGroupedExpr() (ast.Expr, error) {
	p.next() // consume '('

	expr, err := p.parseExpr(none)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expr: %w", err)
	}

	if err = p.expect(token.RParan); err != nil {
		return nil, fmt.Errorf("expected closing paran: %w", err)
	}

	return expr, nil
}

func (p *Parser) parseStmt() (ast.Stmt, error) {
	switch p.currTok.Type {
	case token.Let:
		return p.parseLetStmt()
	}
	return p.parseExprStmt()
}

func (p *Parser) parseLetStmt() (*ast.LetStmt, error) {
	// let
	if err := p.expect(token.Ident); err != nil {
		return nil, err
	}
	ident := &ast.IdentExpr{Value: p.currTok.Literal}

	// =
	if err := p.expect(token.Assign); err != nil {
		return nil, err
	}
	p.next() // consume '='

	// expr
	value, err := p.parseExpr(none)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expr: %w", err)
	}

	return &ast.LetStmt{
		Ident: ident,
		Value: value,
	}, nil
}

// func (p *Parser) parseReturnStmt() (*ast.ReturnStmt, error) {
// 	// stmt := &ast.ReturnStmt{Token: p.currTok}

// 	// parse expr

// 	return nil, nil
// }

func (p *Parser) parseExprStmt() (*ast.ExprStmt, error) {
	expr, err := p.parseExpr(none)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ExprStmt: %w", err)
	}

	return &ast.ExprStmt{
		Expr: expr,
	}, nil
}

func (p *Parser) next() {
	p.currTok = p.nextTok
	p.nextTok = p.l.Next()
}

func (p *Parser) expect(t token.Type) error {
	if p.nextTokenIs(t) {
		p.next()
		return nil
	}
	return fmt.Errorf("parser expected %v, got %v", t, p.nextTok)
}

func (p *Parser) currTokenIs(t token.Type) bool {
	return p.currTok.Type == t
}

func (p *Parser) nextTokenIs(t token.Type) bool {
	return p.nextTok.Type == t
}

func (p *Parser) currTokenPrecedence() int {
	if p, ok := precedences[p.currTok.Type]; ok {
		return p
	}
	return none
}

func (p *Parser) nextTokenPrecedence() int {
	if p, ok := precedences[p.nextTok.Type]; ok {
		return p
	}
	return none
}

package lexer

import (
	"github.com/tombuente/lily/token"
)

type Lexer struct {
	src string

	currPos int
	nextPos int
	ch      byte // char at currPos
}

func New(src string) *Lexer {
	l := &Lexer{
		src: src,
	}
	return l
}

func (l *Lexer) Next() token.Token {
	l.next()
	l.skipWhitespace()

	switch l.ch {
	case '=':
		if l.nextChar() == '=' {
			ch := l.ch
			l.next()
			return token.Token{Type: token.EQ, Literal: string(ch) + string(l.ch)}
		}
		return token.Token{Type: token.Assign, Literal: string(l.ch)}
	case '!':
		if l.nextChar() == '=' {
			ch := l.ch
			l.next()
			return token.Token{Type: token.NotEQ, Literal: string(ch) + string(l.ch)}
		}
		return token.Token{Type: token.Bang, Literal: string(l.ch)}
	case '+':
		return token.Token{Type: token.Plus, Literal: string(l.ch)}
	case '-':
		return token.Token{Type: token.Minus, Literal: string(l.ch)}
	case '*':
		return token.Token{Type: token.Asterisk, Literal: string(l.ch)}
	case '/':
		return token.Token{Type: token.Slash, Literal: string(l.ch)}
	case '<':
		return token.Token{Type: token.Less, Literal: string(l.ch)}
	case '>':
		return token.Token{Type: token.Greater, Literal: string(l.ch)}
	case ';':
		return token.Token{Type: token.Semicolon, Literal: string(l.ch)}
	case '(':
		return token.Token{Type: token.LParan, Literal: string(l.ch)}
	case ')':
		return token.Token{Type: token.RParan, Literal: string(l.ch)}
	case '{':
		return token.Token{Type: token.LBrace, Literal: string(l.ch)}
	case '}':
		return token.Token{Type: token.RBrace, Literal: string(l.ch)}
	case ',':
		return token.Token{Type: token.Comma, Literal: string(l.ch)}
	case '"':
		return token.Token{Type: token.String, Literal: l.readString()}
	case 0:
		return token.Token{Type: token.EOF, Literal: "EOF"}
	}

	if isDigit(l.ch) {
		return token.Token{Type: token.Int, Literal: l.readNumber()}
	}
	if isLetter(l.ch) {
		literal := l.readLiteral()
		return token.Token{Type: token.LookupLiteral(literal), Literal: literal}
	}

	return token.Token{Type: token.Illegal, Literal: string(l.ch)}
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.next()
	}
}

func (l *Lexer) next() {
	if l.nextPos >= len(l.src) {
		l.ch = 0
	} else {
		l.ch = l.src[l.nextPos]
	}
	l.currPos = l.nextPos
	l.nextPos++
}

func (l *Lexer) nextChar() byte {
	if l.nextPos >= len(l.src) {
		return 0
	} else {
		return l.src[l.nextPos]
	}
}

func (l *Lexer) readNumber() string {
	pos := l.currPos
	for isDigit(l.nextChar()) {
		l.next()
	}
	return l.src[pos : l.currPos+1] // exclusive of end
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) readLiteral() string {
	pos := l.currPos
	for isLetter(l.nextChar()) {
		l.next()
	}
	return l.src[pos : l.currPos+1] // exclusive end
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func (l *Lexer) readString() string {
	l.next() // consume '"'

	pos := l.currPos
	for l.nextChar() != '"' {
		l.next()
	}
	literal := l.src[pos : l.currPos+1]

	l.next() // consume '"'
	return literal
}

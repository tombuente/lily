package token

const (
	Illegal Type = "illegal"

	Semicolon Type = "semicolon"
	Bang      Type = "exclamation_mark"
	Comma     Type = ","

	Assign   Type = "="
	Minus    Type = "-"
	Plus     Type = "+"
	Asterisk Type = "*"
	Slash    Type = "/"
	EQ       Type = "=="
	NotEQ    Type = "!="
	Less     Type = "<"
	Greater  Type = ">"

	Int    Type = "int"
	True   Type = "true"
	False  Type = "false"
	Ident  Type = "ident"
	String Type = "string"

	Let    Type = "let"
	Return Type = "return"
	If     Type = "if"
	Fn     Type = "fn"

	LParan Type = "("
	RParan Type = ")"
	LBrace Type = "{"
	RBrace Type = "}"

	EOF Type = "eof"
)

var keywords = map[string]Type{
	"let":    Let,
	"return": Return,
	"if":     If,
	"true":   True,
	"false":  False,
	"fn":     Fn,
}

type Type string

type Token struct {
	Type    Type   `json:"type"`
	Literal string `json:"literal"`
}

func LookupLiteral(literal string) Type {
	if tokenType, ok := keywords[literal]; ok {
		return tokenType
	}
	return Ident
}

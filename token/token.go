package token

const (
	Illegal Type = "illegal"

	Semicolon       Type = "semicolon"
	ExclamationMark Type = "exclamation_mark"

	Assign   Type = "assign"
	Minus    Type = "minus"
	Plus     Type = "plus"
	Asterisk Type = "asterix"
	Slash    Type = "slash"
	EQ       Type = "equal"
	NotEQ    Type = "not_equal"
	Less     Type = "less"
	Greater  Type = "greater"

	Int   Type = "int"
	True  Type = "true"
	False Type = "false"
	Ident Type = "ident"

	Let    Type = "let"
	Return Type = "return"

	LParan Type = "left_paranthese"
	RParan Type = "right_parantheses"

	EOF Type = "eof"
)

var keywords = map[string]Type{
	"let":    Let,
	"return": Return,
	"true":   True,
	"false":  False,
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

package token

import "fmt"

type TokenType int

const (
	ILLEGAL = iota
	EOF

	// Identifiers + literals
	IDENT // add, foobar, x, y, ...
	INT

	// Operators
	ASSIGN
	PLUS
	MINUS
	BANG
	ASTERISK
	SLASH

	LT
	GT

	EQ
	NOT_EQ

	// Delimiters
	COMMA
	SEMICOLON

	LPAREN
	RPAREN
	LBRACE
	RBRACE

	// Keywords
	FUNCTION
	LET
	TRUE
	FALSE
	IF
	ELSE
	RETURN
)

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"fn":     FUNCTION,
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

func (t TokenType) String() string {
	names := [...]string{
		ILLEGAL:   "ILLEGAL",
		EOF:       "EOF",
		IDENT:     "IDENT",
		INT:       "INT",
		ASSIGN:    "ASSIGN",
		PLUS:      "PLUS",
		MINUS:     "MINUS",
		BANG:      "BANG",
		ASTERISK:  "ASTERISK",
		SLASH:     "SLASH",
		LT:        "LT",
		GT:        "GT",
		EQ:        "EQ",
		NOT_EQ:    "NOT_EQ",
		COMMA:     "COMMA",
		SEMICOLON: "SEMICOLON",
		LPAREN:    "LPAREN",
		RPAREN:    "RPAREN",
		LBRACE:    "LBRACE",
		RBRACE:    "RBRACE",
		FUNCTION:  "FUNCTION",
		LET:       "LET",
		TRUE:      "TRUE",
		FALSE:     "FALSE",
		IF:        "IF",
		ELSE:      "ELSE",
		RETURN:    "RETURN",
	}
	if int(t) < len(names) {
		return names[t]
	}
	return fmt.Sprintf("TokenType(%d)", t)
}

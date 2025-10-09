package token

import "fmt"

type TokenType int

const (
	ILLEGAL = iota
	EOF

	// Identifiers + literals
	IDENT // add, foobar, x, y, ...
	INT
	FLOAT
	STRING
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
	COLON

	LPAREN
	RPAREN
	LBRACE
	RBRACE
	LBRACKET
	RBRACKET

	// Keywords
	FUNCTION
	MUT
	CONST
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
	"mut":    MUT,
	"const":  CONST,
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
		ILLEGAL:   "ILLEGAL", // nielegalny znak
		EOF:       "EOF",
		IDENT:     "IDENTIFIER", // identyfikator
		INT:       "0",          // liczba całkowita (symbolicznie)
		FLOAT:     "0.0",
		ASSIGN:    "=",
		PLUS:      "+",
		MINUS:     "-",
		BANG:      "!",
		ASTERISK:  "*",
		SLASH:     "/",
		LT:        "<",
		GT:        ">",
		EQ:        "==",
		NOT_EQ:    "!=",
		COMMA:     ",",
		SEMICOLON: ";",
		LPAREN:    "(",
		RPAREN:    ")",
		LBRACE:    "{",
		RBRACE:    "}",
		FUNCTION:  "fn",
		MUT:       "mut",
		CONST:     "const",
		TRUE:      "true",
		FALSE:     "false",
		IF:        "if",
		ELSE:      "else",
		RETURN:    "return",
		STRING:    `""""`,
		LBRACKET:  "[",
		RBRACKET:  "]",
		COLON:     ":",
	}
	if int(t) < len(names) {
		return names[t]
	}
	return fmt.Sprintf("TokenType(%d)", t)
}

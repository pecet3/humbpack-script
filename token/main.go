package token

import "fmt"

type TokenType int

const (
	ILLEGAL = iota
	EOF

	// Identifiers + literals
	IDENT // add, foobar, x, y, ...
	INT
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

	LPAREN
	RPAREN
	LBRACE
	RBRACE
	LBRACKET
	RBRACKET

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
		ILLEGAL:   "ILLEGAL", // nielegalny znak
		EOF:       "EOF",
		IDENT:     "id", // identyfikator
		INT:       "0",  // liczba całkowita (symbolicznie)
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
		LET:       "let",
		TRUE:      "true",
		FALSE:     "false",
		IF:        "if",
		ELSE:      "else",
		RETURN:    "return",
		STRING:    `""""`,
		LBRACKET:  "[",
		RBRACKET:  "]",
	}
	if int(t) < len(names) {
		return names[t]
	}
	return fmt.Sprintf("TokenType(%d)", t)
}

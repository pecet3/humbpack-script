package parser_test

import (
	"testing"

	"github.com/pecet3/aga-script/ast"
	"github.com/pecet3/aga-script/lexer"
	"github.com/pecet3/aga-script/parser"
)

func TestLetStatement(t *testing.T) {
	input := `
	let x = 5;
	let y= 10;
	let foo = 2137;
	`
	l := lexer.New(input)
	p := parser.New(l)

	t.Log(l)
	program := p.ParseProgram()
	t.Log(program)
	if program == nil {
		t.Fatalf("Program returned nil")
	}
	if len(program.Statements) != 3 {
		t.Fatalf("no 3 statements")
	}

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foo"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}
	}
}

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral not 'let'. got=%q", s.TokenLiteral())
		return false
	}
	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement. got=%T", s)
		return false
	}
	if letStmt.Name.Value != name {
		t.Errorf("letStmt.Name.Value not '%s'. got=%s", name, letStmt.Name.Value)
		return false
	}
	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("s.Name not '%s'. got=%s", name, letStmt.Name.Value)
		return false
	}
	return true
}

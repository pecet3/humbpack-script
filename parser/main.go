package parser

import (
	"fmt"
	"strconv"

	"github.com/pecet3/hmbk-script/ast"
	"github.com/pecet3/hmbk-script/lexer"
	"github.com/pecet3/hmbk-script/token"
)

// precedences
const (
	LOWEST      = iota
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	INDEX
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
	token.DOT:      INDEX,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token
	errors    []string

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func (p *Parser) registerPrefixParseFn(t token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[t] = fn
}
func (p *Parser) registerInfixParseFn(t token.TokenType, fn infixParseFn) {
	p.infixParseFns[t] = fn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},

		prefixParseFns: make(map[token.TokenType]prefixParseFn),
		infixParseFns:  make(map[token.TokenType]infixParseFn),
	}

	p.registerPrefixParseFn(token.IDENT, p.parseIdentifier)
	p.registerPrefixParseFn(token.INT, p.parseIntegerLiteral)
	p.registerPrefixParseFn(token.FLOAT, p.parseFloatLiteral)

	p.registerPrefixParseFn(token.BANG, p.parsePrefixExpression)
	p.registerPrefixParseFn(token.STRING, p.parseStringLiteral)
	p.registerPrefixParseFn(token.MINUS, p.parsePrefixExpression)
	p.registerPrefixParseFn(token.TRUE, p.parseBoolean)
	p.registerPrefixParseFn(token.FALSE, p.parseBoolean)
	p.registerPrefixParseFn(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefixParseFn(token.IF, p.parseIfExpression)
	p.registerPrefixParseFn(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefixParseFn(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefixParseFn(token.LBRACE, p.parseHashLiteral)

	p.registerInfixParseFn(token.PLUS, p.parseInfixExpression)
	p.registerInfixParseFn(token.MINUS, p.parseInfixExpression)
	p.registerInfixParseFn(token.SLASH, p.parseInfixExpression)
	p.registerInfixParseFn(token.ASTERISK, p.parseInfixExpression)
	p.registerInfixParseFn(token.EQ, p.parseInfixExpression)
	p.registerInfixParseFn(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfixParseFn(token.LT, p.parseInfixExpression)
	p.registerInfixParseFn(token.GT, p.parseInfixExpression)
	p.registerInfixParseFn(token.LPAREN, p.parseCallExpression)
	p.registerInfixParseFn(token.LBRACKET, p.parseIndexExpression)
	p.registerInfixParseFn(token.DOT, p.parseModuleExpression)
	// read two tokens to setup
	p.nextToken()
	p.nextToken()
	return p
}

// helpers

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p

	}
	return LOWEST
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	p.errors = append(p.errors, fmt.Sprintf(`EXPECETED: %s GOT: %s`, t.String(), p.peekToken.Type.String()))
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

// PARSER functions

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		if p.curTokenIs(token.MODULE) {
			mod := p.ParseModule()
			if mod != nil {
				program.Statements = append(program.Statements, mod)
			}
			p.nextToken()
			continue
		}
		stmt := p.parseStatement()
		program.Statements = append(program.Statements, stmt)

		p.nextToken()
	}
	return program
}

func (p *Parser) ParseModule() *ast.Module {
	module := &ast.Module{}
	module.Statements = []ast.Statement{}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	module.Name = p.curToken.Literal

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	p.nextToken()
	for !p.peekTokenIs(token.EOF) && !p.curTokenIs(token.RBRACE) {
		stmt := p.parseStatement()
		module.Statements = append(module.Statements, stmt)

		p.nextToken()
	}
	return module
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.MUT:
		stmt := p.parseMutStatement()
		return stmt
	case token.CONST:
		stmt := p.parseConstStatement()
		return stmt
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()
	stmt.ReturnValue = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}
func (p *Parser) parseMutStatement() *ast.MutStatement {
	stmt := &ast.MutStatement{Token: p.curToken}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseConstStatement() *ast.ConstStatement {
	stmt := &ast.ConstStatement{Token: p.curToken}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseAssignmentStatement(name ast.Expression) ast.Statement {
	stmt := &ast.AssignmentStatement{
		Token: p.curToken,
		Name:  name.(*ast.Identifier),
	}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() ast.Statement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if ident, ok := stmt.Expression.(*ast.Identifier); ok && p.peekTokenIs(token.ASSIGN) {
		return p.parseAssignmentStatement(ident)
	}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}
	return leftExp
}

// minor parser functions

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)

	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf(`cannot parse %q as integer`, p.curToken.Literal))
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)

	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf(`cannot parse %q as float`, p.curToken.Literal))
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	exp := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	exp.Right = p.parseExpression(PREFIX)
	return exp
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	exp := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	exp.Right = p.parseExpression(precedence)

	return exp
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	exp := &ast.IfExpression{
		Token: p.curToken,
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	exp.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	exp.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		if !p.expectPeek(token.LBRACE) {
			return nil
		}
		exp.Alternative = p.parseBlockStatement()
	}

	return exp
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}
	p.nextToken()
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		block.Statements = append(block.Statements, stmt)

		p.nextToken()
	}
	return block
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{
		Token: p.curToken,
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParams()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()
	return lit
}

func (p *Parser) parseFunctionParams() []*ast.Identifier {
	idents := []*ast.Identifier{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return idents
	}

	p.nextToken()

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	idents = append(idents, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()

		ident := &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		}

		idents = append(idents, ident)
	}
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return idents
}

func (p *Parser) parseCallExpression(fn ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: fn}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}
	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()

	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()

		list = append(list, p.parseExpression(LOWEST))
	}
	if !p.expectPeek(end) {
		return nil
	}
	return list
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()

	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}
func (p *Parser) parseModuleExpression(left ast.Expression) ast.Expression {
	exp := &ast.ModuleExpression{Token: p.curToken, Left: left}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	exp.Index = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	return exp
}

func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		key := p.parseExpression(LOWEST)

		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken()

		value := p.parseExpression(LOWEST)

		hash.Pairs[key] = value

		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}

	}
	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return hash

}

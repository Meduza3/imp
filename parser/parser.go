package parser

import (
	"fmt"

	"github.com/Meduza3/imp/ast"
	"github.com/Meduza3/imp/token"

	"github.com/Meduza3/imp/lexer"
)

type Parser struct {
	l *lexer.Lexer

	errors []string

	curToken  token.Token
	peekToken token.Token
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	fmt.Printf("in ParseProgram: %v\n", p.curToken)

	program := &ast.Program{}
	program.Commands = []ast.Command{}
	for p.curToken.Type != token.EOF {
		command := p.parseCommand()
		if command != nil {
			program.Commands = append(program.Commands, command)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseCommand() ast.Command {
	fmt.Printf("in parseCommand: %v\n", p.curToken)

	switch p.curToken.Type {
	case token.PIDENTIFIER:
		return p.parseAssignCommand()
	case token.IF:
		return p.parseIfCommand()
	default:
		return nil
	}
}

func (p *Parser) parseAssignCommand() *ast.AssignCommand {
	fmt.Printf("in parseAssignCommand: %v\n", p.curToken)

	identifier := p.parseIdentifier()
	if identifier == nil {
		return nil // Error handling - failed to parse identifier
	}

	if !p.expectPeek(token.ASSIGN) {
		p.peekError(token.ASSIGN)
		return nil // Error handling - missing ':='
	}
	assignToken := p.curToken
	p.nextToken()
	mathExpression := p.parseMathExpression()
	if mathExpression == nil {
		return nil // Error handling - failed to parse math expression
	}
	if !p.expectPeek(token.SEMICOLON) {
		p.peekError(token.SEMICOLON)
		return nil // Error handling - missing ':='
	}
	p.nextToken() // eat ';'

	return &ast.AssignCommand{
		Identifier:     *identifier,
		Token:          assignToken, // token.ASSIGN
		MathExpression: *mathExpression,
	}
}

func (p *Parser) parseIfCommand() *ast.IfCommand {
	fmt.Printf("in parseIfCommand: %v\n", p.curToken)
	var ifCmd ast.IfCommand
	ifCmd.Token = p.curToken
	p.nextToken()                           // Eat "IF"
	ifCmd.Condition = *p.parseCondition()   // Eat conditon
	p.nextToken()                           // Eat "THEN"
	ifCmd.ThenCommands = *p.parseCommands() // Eat commands
	if p.peekToken.Type == token.ELSE {
		p.nextToken()                           // Eat "ELSE"
		ifCmd.ElseCommands = *p.parseCommands() // Eat commands
	}
	p.nextToken() // Eat "ENDIF"
	return &ifCmd
}

func (p *Parser) parseCommands() *[]ast.Command {
	fmt.Printf("in parseCommands: %v\n", p.curToken)
	var commands = []ast.Command{}
	for p.curToken.Type != token.ELSE && p.curToken.Type != token.ENDIF && p.curToken.Type != token.EOF {
		command := p.parseCommand()
		if command == nil {
			break
		}
		commands = append(commands, command)
		p.nextToken()
	}
	return &commands
}

func (p *Parser) parseIdentifier() *ast.Identifier {
	fmt.Printf("in parseIdentifier: %v\n", p.curToken)
	identifier := &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
	if p.peekTokenIs(token.LBRACKET) { // LBRACKET = [
		p.nextToken() // Consume '['
		p.nextToken() // Move to the token inside the brackets

		identifier.Index = p.parseIndex() // Parse the index as an expression

		if !p.expectPeek(token.RBRACKET) { // RBRACKET = ]
			p.peekError(token.RBRACKET)
			return nil // Error handling can be added here
		}
	}
	return identifier
}

func (p *Parser) parseIndex() ast.Expression {
	fmt.Printf("in parseIndex: %v\n", p.curToken)

	var index ast.Expression
	switch p.curToken.Type {
	case token.NUM:
		index = &ast.NumberLiteral{Token: p.curToken, Value: p.curToken.Literal}
	case token.PIDENTIFIER:
		index = &ast.Pidentifier{Token: p.curToken, Value: p.curToken.Literal}
	}
	return index
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		return false
	}
}

func (p *Parser) parseMathExpression() *ast.MathExpression {
	fmt.Printf("in parseMathExpression: %v\n", p.curToken)
	left := p.parseValue()
	if left == nil {
		return nil // Error handling - failed to parse left-hand value
	}
	if !isOperator(p.peekToken.Type) {
		return &ast.MathExpression{
			Left:     left,
			Operator: token.Token{Type: token.ILLEGAL, Literal: ""},
			Right:    nil, // no right operand
		}
	}
	p.nextToken()
	operator := p.curToken
	p.nextToken()
	right := p.parseValue()
	if right == nil {
		return nil // Error handling - failed to parse right-hand value
	}
	return &ast.MathExpression{
		Left:     left,
		Operator: operator,
		Right:    right,
	}
}

func isOperator(tt token.TokenType) bool {
	switch tt {
	case token.PLUS, token.MINUS, token.DIVIDE, token.MULT:
		return true
	}
	return false
}

func isConditionOperator(tt token.TokenType) bool {
	switch tt {
	case token.LEQ, token.GEQ, token.LE, token.GE, token.EQUALS, token.NEQUALS:
		return true
	}
	return false
}

func (p *Parser) parseCondition() *ast.Condition {
	fmt.Printf("in parseCondition: %v\n", p.curToken)
	left := p.parseValue()
	if left == nil {
		return nil // Error handling - failed to parse left-hand value
	}
	p.nextToken()
	operator := p.curToken
	if !isConditionOperator(operator.Type) {
		p.peekError(token.EQUALS)
		return nil // Error handling - invalid operator
	}
	p.nextToken()
	right := p.parseValue()
	if right == nil {
		return nil // Error handling - failed to parse right-hand value
	}
	p.nextToken()
	return &ast.Condition{
		Left:     left,
		Operator: operator,
		Right:    right,
	}
}

func (p *Parser) parseValue() ast.Value {
	switch p.curToken.Type {
	case token.NUM:
		return &ast.NumberLiteral{
			Token: p.curToken,
			Value: p.curToken.Literal,
		}
	case token.PIDENTIFIER:
		return p.parseIdentifier()
	}
	fmt.Printf("in parseValue: %v\n", p.curToken)
	return nil
}

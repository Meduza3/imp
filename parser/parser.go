package parser

import (
	"github.com/Meduza3/imp/ast"
	"github.com/Meduza3/imp/token"

	"github.com/Meduza3/imp/lexer"
)

type Parser struct {
	l *lexer.Lexer

	curToken  token.Token
	peekToken token.Token
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
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
	switch p.curToken.Type {
	case token.PIDENTIFIER:
		return p.parseAssignCommand()
	default:
		return nil
	}
}

func (p *Parser) parseAssignCommand() *ast.AssignCommand {
	identifier := p.parseIdentifier()
	if identifier == nil {
		return nil // Error handling - failed to parse identifier
	}
	if !p.expectPeek(token.ASSIGN) {
		return nil // Error handling - missing ':='
	}
	token := p.curToken
	mathExpression := p.parseMathExpression()
	if mathExpression == nil {
		return nil // Error handling - failed to parse math expression
	}

	return &ast.AssignCommand{
		Identifier:     *identifier,
		Token:          token, // token.ASSIGN
		MathExpression: *mathExpression,
	}
}

func (p *Parser) parseIdentifier() *ast.Identifier {
	if !p.curTokenIs(token.PIDENTIFIER) {
		return nil // Error handling - expected identifier
	}
	identifier := &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
	if p.peekTokenIs(token.LBRACKET) { // LBRACKET = [
		p.nextToken() // Consume '['
		p.nextToken() // Move to the token inside the brackets

		identifier.Index = p.parseIndex() // Parse the index as an expression

		if !p.expectPeek(token.RBRACKET) { // RBRACKET = ]
			return nil // Error handling can be added here
		}
	}
	p.nextToken()
	return identifier
}

func (p *Parser) parseIndex() ast.Expression {
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
	left := p.parseValue()
	operator := p.curToken
	p.nextToken()
	right := p.parseValue()
	return &ast.MathExpression{
		Left:     left,
		Operator: operator,
		Right:    right,
	}
}

func (p *Parser) parseValue() ast.Value {
	switch p.curToken.Type {
	case token.NUM:
		p.nextToken()
		return &ast.NumberLiteral{
			Token: p.curToken,
			Value: p.curToken.Literal,
		}
	case token.PIDENTIFIER:
		return p.parseIdentifier()
	}
	return nil
}

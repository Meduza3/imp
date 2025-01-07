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

func (p *Parser) peekError(t token.TokenType) error {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
	return fmt.Errorf(msg)
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// Parse program is the first function that is being called when you start to parse the program
func (p *Parser) ParseProgram() *ast.Program {
	fmt.Printf("in ParseProgram: %v\n", p.curToken)
	//Currently the program is a list of commands
	program := &ast.Program{}
	program.Commands = []ast.Command{}
	for p.curToken.Type != token.EOF { // Iterate over the tokens until token.EOF
		command, err := p.parseCommand() // This uses the current token to figure out which command it is
		if err != nil {
			p.errors = append(p.errors, fmt.Sprintf("Failed to parseCommand: %v", err))
			p.nextToken()
			continue
		}
		program.Commands = append(program.Commands, command)
		p.nextToken() // Get to the next token at the end of the core loop
	}
	return program
}

// Uses the current token to identify which command it is
// Should return NIL when it failed to parse the command
func (p *Parser) parseCommand() (ast.Command, error) {
	fmt.Printf("in parseCommand: %v\n", p.curToken)

	switch p.curToken.Type {
	case token.PIDENTIFIER:
		return p.parseAssignCommand()
	case token.IF:
		return p.parseIfCommand()
	default:
		return nil, fmt.Errorf("failed to parseCommand, no matching command for token: %v", p.curToken.Type)
	}
}

func (p *Parser) parseAssignCommand() (*ast.AssignCommand, error) {
	fmt.Printf("in parseAssignCommand: %v\n", p.curToken)

	identifier, err := p.parseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("failed to parse assign command: %v", err) // Error handling - failed to parse identifier
	}

	if !p.expectPeek(token.ASSIGN) {
		err := p.peekError(token.ASSIGN)
		return nil, err
	}
	assignToken := p.curToken
	p.nextToken()
	mathExpression, err := p.parseMathExpression()
	if err != nil {
		return nil, fmt.Errorf("failed to parse expression: %v", err) // Error handling - failed to parse math expression
	}
	if !p.expectPeek(token.SEMICOLON) {
		err := p.peekError(token.SEMICOLON)
		return nil, err // Error handling - missing ':='
	}
	p.nextToken() // eat ';'

	return &ast.AssignCommand{
		Identifier:     *identifier,
		Token:          assignToken, // token.ASSIGN
		MathExpression: *mathExpression,
	}, nil
}

func (p *Parser) parseIfCommand() (*ast.IfCommand, error) {
	fmt.Printf("in parseIfCommand: %v\n", p.curToken)
	ifCmd := ast.IfCommand{Token: p.curToken}
	p.nextToken()                        // Eat "IF"
	condition, err := p.parseCondition() // Eat conditon
	if err != nil {
		return nil, fmt.Errorf("failed to parse condition: %v", err)
	}
	ifCmd.Condition = *condition
	if !p.curTokenIs(token.THEN) {
		return nil, fmt.Errorf("parseIfCommand: expected THEN, got %s", p.curToken)
	}
	p.nextToken()                           // skip THEN
	ifCmd.ThenCommands = *p.parseCommands() // Eat commands
	if p.peekToken.Type == token.ELSE {
		p.nextToken()                           // Eat "ELSE"
		ifCmd.ElseCommands = *p.parseCommands() // Eat commands
	}
	if !p.curTokenIs(token.ENDIF) {
		return nil, fmt.Errorf("parseIfCommand: expected ENDIF, got %s", p.curToken)
	}
	p.nextToken() // Eat "ENDIF"
	return &ifCmd, nil
}

func (p *Parser) parseCommands() *[]ast.Command {
	fmt.Printf("in parseCommands: %v\n", p.curToken)
	var commands = []ast.Command{}
	for p.curToken.Type != token.ELSE && p.curToken.Type != token.ENDIF && p.curToken.Type != token.EOF {
		command, err := p.parseCommand()
		if err != nil {
			p.errors = append(p.errors, fmt.Sprintf("failed to parsce command: %v", err))
			break
		}
		commands = append(commands, command)
		p.nextToken()
	}
	return &commands
}

func (p *Parser) parseIdentifier() (*ast.Identifier, error) {
	fmt.Printf("in parseIdentifier: %v\n", p.curToken)
	identifier := &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
	if p.peekTokenIs(token.LBRACKET) { // LBRACKET = [
		p.nextToken() // Consume '['
		p.nextToken() // Move to the token inside the brackets

		index, err := p.parseIndex() // Parse the index as an expression
		if err != nil {
			return nil, fmt.Errorf("in parseIdentifier(): failed to parse index: %v", err)
		}
		identifier.Index = index

		if !p.expectPeek(token.RBRACKET) { // RBRACKET = ]
			err = p.peekError(token.RBRACKET)
			return nil, err // Error handling can be added here
		}
	}
	return identifier, nil
}

func (p *Parser) parseIndex() (ast.Expression, error) {
	fmt.Printf("in parseIndex: %v\n", p.curToken)

	var index ast.Expression
	switch p.curToken.Type {
	case token.NUM:
		index = &ast.NumberLiteral{Token: p.curToken, Value: p.curToken.Literal}
	case token.PIDENTIFIER:
		index = &ast.Pidentifier{Token: p.curToken, Value: p.curToken.Literal}
	default:
		return nil, fmt.Errorf("failed to parse index: not valid token: %v", p.curToken)
	}
	return index, nil
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

func (p *Parser) parseMathExpression() (*ast.MathExpression, error) {
	fmt.Printf("in parseMathExpression: %v\n", p.curToken)
	left, err := p.parseValue()
	if err != nil {
		return nil, fmt.Errorf("in parseMathExpression: failed to parse left value: %v", err) // Error handling - failed to parse left-hand value
	}
	if !isOperator(p.peekToken.Type) {
		return &ast.MathExpression{
			Left:     left,
			Operator: token.Token{Type: token.ILLEGAL, Literal: ""},
			Right:    nil, // no right operand
		}, nil
	}
	p.nextToken()
	operator := p.curToken
	p.nextToken()
	right, err := p.parseValue()
	if err != nil {
		return nil, fmt.Errorf("in parseMathExpression: failed to parse right value: %v", err) // Error handling - failed to parse right-hand value
	}
	return &ast.MathExpression{
		Left:     left,
		Operator: operator,
		Right:    right,
	}, nil
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

func (p *Parser) parseCondition() (*ast.Condition, error) {
	fmt.Printf("in parseCondition: %v\n", p.curToken)
	left, err := p.parseValue()
	if err != nil {
		return nil, fmt.Errorf("in parseCondition: failed to parse left value: %v", err) // Error handling - failed to parse left-hand value
	}
	if !isConditionOperator(p.peekToken.Type) {
		return nil, fmt.Errorf("parseCondition: expected a comparison operator, got %s", p.peekToken.Type)
	}
	p.nextToken()
	operator := p.curToken
	p.nextToken()
	right, err := p.parseValue()
	if err != nil {
		return nil, fmt.Errorf("in parseCondition: failed to parse right value: %v", err) // Error handling - failed to parse right-hand value
	}
	return &ast.Condition{
		Left:     left,
		Operator: operator,
		Right:    right,
	}, nil
}

func (p *Parser) parseValue() (ast.Value, error) {
	fmt.Printf("in parseValue: %v\n", p.curToken)
	switch p.curToken.Type {
	case token.NUM:
		return &ast.NumberLiteral{
			Token: p.curToken,
			Value: p.curToken.Literal,
		}, nil
	case token.PIDENTIFIER:
		return p.parseIdentifier()
	}
	return nil, fmt.Errorf("failed to parse value for token: %v", p.curToken)
}

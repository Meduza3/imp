package parser

import (
	"fmt"
	"log"
	"time"

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

func (p *Parser) addError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	p.errors = append(p.errors, msg)
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
	fmt.Printf("curToken = %v, peekToken = %v\n", p.curToken, p.peekToken)
}

// Parse program is the first function that is being called when you start to parse the program
func (p *Parser) ParseProgram() *ast.Program {
	fmt.Printf("in ParseProgram: %v\n", p.curToken)
	//Currently the main is a list of commands
	token := token.Token{Literal: "PROGRAM_ALL", Type: token.PROGRAM_ALL}
	procedures := p.parseProcedures()
	//time.Sleep(10 * time.Second)
	main, err := p.parseMain()
	if err != nil {
		log.Fatalf("failed to parse main: %v", err)
	}
	program := &ast.Program{token, *procedures, *main}
	return program
}

func (p *Parser) parseMain() (*ast.Main, error) {
	main := ast.Main{}
	if !p.curTokenIs(token.PROGRAM) {
		return nil, fmt.Errorf("line %d: expected PROGRAM got %s", p.curToken.Line, p.curToken.Type)
	}
	main.Token = p.curToken
	p.nextToken() // curToken = IS
	if !p.curTokenIs(token.IS) {
		return nil, fmt.Errorf("line %d: expected IS got %s", p.curToken.Line, p.curToken.Type)
	}
	p.nextToken() // curToken = BEGIN
	if !p.curTokenIs(token.BEGIN) {
		decl := p.parseDeclarations()
		main.Declarations = *decl
	}
	p.nextToken() // eat 'BEGIN'
	commands := p.parseCommandsUntil(token.END)
	main.Commands = *commands
	return &main, nil
}

func (p *Parser) parsePidentifier() ast.Pidentifier {
	pid := ast.Pidentifier{
		Value: p.curToken.Literal,
		Token: p.curToken,
	}
	p.nextToken()
	return pid
}

func (p *Parser) parseDeclarations() *[]ast.Declaration {
	var decl = []ast.Declaration{}
	for !p.curTokenIs(token.BEGIN) {
		if !p.peekTokenIs(token.LBRACKET) {
			pid := p.parsePidentifier()
			decl = append(decl, ast.Declaration{IsTable: false, Pidentifier: pid})
		} else {
			pid := p.parsePidentifier() // [ = curToken
			p.nextToken()               // num = curToken
			from := ast.NumberLiteral{
				Token: p.curToken,
				Value: p.curToken.Literal,
			}
			p.nextToken() // : = curToken
			p.nextToken() // num = curtoken
			to := ast.NumberLiteral{
				Token: p.curToken,
				Value: p.curToken.Literal,
			}
			p.nextToken() // ] = curtoken
			decl = append(decl, ast.Declaration{IsTable: true, Pidentifier: pid, From: from, To: to})
		}
		p.nextToken() // eat the ','
		time.Sleep(300 * time.Millisecond)
	}
	return &decl
}

// Uses the current token to identify which command it is
// Should return NIL when it failed to parse the command
func (p *Parser) parseCommand() (ast.Command, error) {
	fmt.Printf("in parseCommand: %v\n", p.curToken)

	switch p.curToken.Type {
	case token.PIDENTIFIER:
		if p.peekTokenIs(token.LPAREN) {
			return p.parseProcCallCommand()
		}
		return p.parseAssignCommand() // or function call!
	case token.IF:
		return p.parseIfCommand()
	case token.WHILE:
		return p.parseWhileCommand()
	case token.REPEAT:
		return p.parseRepeatCommand()
	case token.FOR:
		return p.parseForCommand()
	case token.READ:
		return p.parseReadCommand()
	case token.WRITE:
		return p.parseWriteCommand()
	default:
		return nil, fmt.Errorf("failed to parseCommand, no matching command for token: %v", p.curToken.Type)
	}
}

func (p *Parser) parseProcCallCommand() (*ast.ProcCallCommand, error) {
	procCallToken := p.curToken
	name := p.parsePidentifier()
	if !p.curTokenIs(token.LPAREN) {
		return nil, fmt.Errorf("line %d expected '(' got %s", p.curToken.Line, p.curToken.Type)
	}
	p.nextToken()
	args, err := p.parseArgs()
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments in proccall: %v", err)
	}
	if !p.curTokenIs(token.RPAREN) {
		return nil, fmt.Errorf("line %d expected ')' got %s", p.curToken.Line, p.curToken.Type)
	}
	p.nextToken()
	if !p.curTokenIs(token.SEMICOLON) {
		return nil, fmt.Errorf("line %d expected ';' got %s", p.curToken.Line, p.curToken.Type)
	}
	return &ast.ProcCallCommand{
		Token: procCallToken,
		Name:  name,
		Args:  *args,
	}, nil
}

func (p *Parser) parseArgs() (*[]ast.Pidentifier, error) {
	args := []ast.Pidentifier{}

	if !p.curTokenIs(token.PIDENTIFIER) {
		return nil, fmt.Errorf("failed parsing proccall line %d: expected pidentifier in args, got %s", p.curToken.Line, p.curToken.Type)
	}

	pid := p.parsePidentifier()
	args = append(args, pid)

	for p.curTokenIs(token.COMMA) {
		p.nextToken() // eat ','
		if !p.curTokenIs(token.PIDENTIFIER) {
			return nil, fmt.Errorf("failed parsing proccall line %d: expected pidentifier in args, got %s", p.curToken.Line, p.curToken.Type)
		}

		pid = p.parsePidentifier()
		args = append(args, pid)
	}

	return &args, nil
}

func (p *Parser) parseWhileCommand() (*ast.WhileCommand, error) {
	whileComm := &ast.WhileCommand{}
	whileToken := p.curToken
	p.nextToken()
	condition, err := p.parseCondition()
	if err != nil {
		fmt.Printf("failed to parse condition: %v", err)
	}
	if !p.curTokenIs(token.DO) {
		return nil, fmt.Errorf("failed to parse while command, expected DO got %v", p.curToken.Type)
	}
	p.nextToken()
	commands := p.parseCommandsUntil(token.ENDWHILE)
	whileComm.Token = whileToken
	whileComm.Condition = *condition
	whileComm.Commands = *commands
	return whileComm, nil
}

func (p *Parser) parseRepeatCommand() (*ast.RepeatCommand, error) {
	repComm := &ast.RepeatCommand{}
	repToken := p.curToken
	p.nextToken()
	commands := p.parseCommandsUntil(token.UNTIL)
	p.nextToken()
	condition, err := p.parseCondition()
	if err != nil {
		return nil, fmt.Errorf("failed to parse condition: %v", err)
	}
	if !p.curTokenIs(token.SEMICOLON) {
		return nil, fmt.Errorf("failed to parse for line %d: expected ';' got %s", p.curToken.Line, p.curToken.Type)
	}
	p.nextToken()
	repComm.Token = repToken
	repComm.Commands = *commands
	repComm.Condition = *condition
	return repComm, nil
}

func (p *Parser) parseForCommand() (*ast.ForCommand, error) {
	forComm := &ast.ForCommand{}
	forToken := p.curToken
	p.nextToken()
	pid := p.parsePidentifier()
	if !p.expectPeek(token.FROM) {
		return nil, fmt.Errorf("failed to parse for line %d: expected FROM got %s", p.curToken.Line, p.curToken.Type)
	}
	p.nextToken()
	valFrom, err := p.parseValue()
	if err != nil {
		return nil, fmt.Errorf("failed to parse for: failed to parse value at line %d: %v", p.curToken.Line, err)
	}
	if p.curToken.Type == token.TO {
		forComm.IsDownTo = false
	} else if p.curToken.Type == token.FROM {
		forComm.IsDownTo = true
	} else {
		return nil, fmt.Errorf("failed to parse for line %d: expected DOWNTO or TO got %s", p.curToken.Line, p.curToken.Type)
	}
	p.nextToken()
	valTo, err := p.parseValue()
	if err != nil {
		return nil, fmt.Errorf("failed to parse for: failed to parse value at line %d: %v", p.curToken.Line, err)
	}
	if !p.expectPeek(token.DO) {
		return nil, fmt.Errorf("failed to parse for line %d: expected DO got %s", p.curToken.Line, p.curToken.Type)
	}
	commands := p.parseCommandsUntil(token.ENDFOR)
	p.nextToken()
	forComm.Token = forToken
	forComm.Iterator = pid
	forComm.From = valFrom
	forComm.To = valTo
	forComm.Commands = *commands
	return forComm, nil
}

func (p *Parser) parseReadCommand() (*ast.ReadCommand, error) {
	fmt.Printf("in parseWriteCommand: %v\n", p.curToken)
	tok := p.curToken
	p.nextToken() // Skip "READ". p.curToken now holds the value to read
	value, err := p.parseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("failed to parse read command: failed to parse identifier: %v", err)
	}
	p.nextToken() // skip ';'
	return &ast.ReadCommand{
		Token:      tok,
		Identifier: *value,
	}, nil
}

func (p *Parser) parseWriteCommand() (*ast.WriteCommand, error) {
	fmt.Printf("in parseWriteCommand: %v\n", p.curToken)
	tok := p.curToken
	p.nextToken() // Skip "WRITE". p.curToken now holds the value to write
	value, err := p.parseValue()
	if err != nil {
		return nil, fmt.Errorf("failed to parse write command: failed to parse value: %v", err)
	}
	p.nextToken() // skip ';'
	return &ast.WriteCommand{
		Token: tok,
		Value: value,
	}, nil
}

func (p *Parser) parseAssignCommand() (*ast.AssignCommand, error) {
	fmt.Printf("in parseAssignCommand: %v\n", p.curToken)

	identifier, err := p.parseIdentifier()
	if err != nil {
		p.addError("failed to parse assign command: %v", err)
		return nil, err
	}

	if !p.curTokenIs(token.ASSIGN) {
		errMsg := "expected ASSIGN token ':=' but got %v"
		p.addError(errMsg, p.curToken.Type)
		return nil, fmt.Errorf(errMsg, p.curToken.Type)
	}
	assignToken := p.curToken
	p.nextToken() // consume ':='

	mathExpression, err := p.parseMathExpression()
	if err != nil {
		p.addError("failed to parse assign command: expression error: %v", err)
		return nil, err
	}

	if !p.curTokenIs(token.SEMICOLON) {
		errMsg := "expected semicolon ';' but got %v"
		p.addError(errMsg, p.curToken.Type)
		return nil, fmt.Errorf(errMsg, p.curToken.Type)
	}
	p.nextToken() // consume ';'

	return &ast.AssignCommand{
		Identifier:     *identifier,
		Token:          assignToken,
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
		return nil, fmt.Errorf("parseIfCommand: expected THEN, got %v", p.curToken)
	}
	p.nextToken()                                                       // skip THEN
	ifCmd.ThenCommands = *p.parseCommandsUntil(token.ELSE, token.ENDIF) // Eat commands
	if p.curToken.Type == token.ELSE {
		p.nextToken()                                           // Eat "ELSE"
		ifCmd.ElseCommands = *p.parseCommandsUntil(token.ENDIF) // Eat commands
	}
	if !p.curTokenIs(token.ENDIF) {
		return nil, fmt.Errorf("parseIfCommand: expected ENDIF, got %v", p.curToken)
	}
	p.nextToken() // Eat "ENDIF"
	return &ifCmd, nil
}

func (p *Parser) parseCommands() *[]ast.Command {
	fmt.Printf("in parseCommands: %v\n", p.curToken)
	var commands = []ast.Command{}
	for p.curToken.Type != token.EOF {
		command, err := p.parseCommand()
		if err != nil {
			p.errors = append(p.errors, fmt.Sprintf("failed to parse command: %v", err))
			break
		}
		commands = append(commands, command)
	}
	return &commands
}

func (p *Parser) parseProcedures() *[]ast.Procedure {
	procedures := []ast.Procedure{}
	// parse commands until we hit one of the stopTokens (ELSE, ENDIF) or EOF
	for p.curToken.Type != token.PROGRAM {
		procedure, err := p.parseProcedure()
		if err != nil {
			p.errors = append(p.errors, fmt.Sprintf("failed to parse procedure: %v", err))
			continue
		}
		procedures = append(procedures, *procedure)
	}
	return &procedures
}

func (p *Parser) parseProcedure() (*ast.Procedure, error) {
	proc := ast.Procedure{}
	if !p.curTokenIs(token.PROCEDURE) {
		return nil, fmt.Errorf("line %d: expected PROCEDURE got %s", p.curToken.Line, p.curToken.Type)
	}
	proc.Token = p.curToken
	p.nextToken()
	procHead, err := p.parseProcHead()
	if err != nil {
		return nil, fmt.Errorf("failed to parse prochead: %v", err)
	}
	if !p.curTokenIs(token.IS) {
		return nil, fmt.Errorf("line %d: expected IS got %s", p.curToken.Line, p.curToken.Type)
	}
	p.nextToken()
	if !p.curTokenIs(token.BEGIN) {
		declarations := p.parseDeclarations()
		proc.Declarations = *declarations
	}
	p.nextToken() // eat 'BEGIN'
	commands := p.parseCommandsUntil(token.END)
	p.nextToken()
	proc.Commands = *commands
	proc.ProcHead = *procHead
	return &proc, nil
}

func (p *Parser) parseProcHead() (*ast.ProcHead, error) {
	procHead := ast.ProcHead{}
	procHead.Token = p.curToken
	name := p.parsePidentifier()
	if !p.curTokenIs(token.LPAREN) {
		return nil, fmt.Errorf("line %d expected '(' got %s", p.curToken.Line, p.curToken.Type)
	}
	p.nextToken()
	argsDecl, err := p.parseArgsDecl()
	if err != nil {
		return nil, fmt.Errorf("failed to parse args: %s", err)
	}
	if !p.curTokenIs(token.RPAREN) {
		return nil, fmt.Errorf("line %d expected ')' got %s", p.curToken.Line, p.curToken.Type)
	}
	p.nextToken()
	procHead.ArgsDecl = *argsDecl
	procHead.Name = name
	return &procHead, nil
}

func (p *Parser) parseArgsDecl() (*[]ast.ArgDecl, error) {
	args := []ast.ArgDecl{}

	if !p.curTokenIs(token.PIDENTIFIER) {
		return nil, fmt.Errorf("failed parsing argsdecl line %d: expected pidentifier in args, got %s", p.curToken.Line, p.curToken.Type)
	}

	arg, err := p.parseArgDecl()
	if err != nil {
		return nil, fmt.Errorf("failed parsing argsdecl: %v", err)
	}
	args = append(args, *arg)
	for p.curTokenIs(token.COMMA) {
		p.nextToken() // eat ','
		if !p.curTokenIs(token.PIDENTIFIER) {
			return nil, fmt.Errorf("failed parsing argsdecl line %d: expected pidentifier in args, got %s", p.curToken.Line, p.curToken.Type)
		}
		arg, err := p.parseArgDecl()
		if err != nil {
			return nil, fmt.Errorf("failed parsing argsdecl: %v", err)
		}
		args = append(args, *arg)
	}
	return &args, nil
}

func (p *Parser) parseArgDecl() (*ast.ArgDecl, error) {
	var arg ast.ArgDecl
	if p.curTokenIs(token.T) {
		arg.IsTable = true
		p.nextToken()
	}
	name := p.parsePidentifier()
	arg.Name = name
	arg.Token = p.curToken
	p.nextToken()
	return &arg, nil
}

func (p *Parser) parseCommandsUntil(stopTokens ...token.TokenType) *[]ast.Command {
	commands := []ast.Command{}
	// parse commands until we hit one of the stopTokens (ELSE, ENDIF) or EOF
	for !p.inSet(p.curToken.Type, stopTokens) && p.curToken.Type != token.EOF {
		command, err := p.parseCommand()
		if err != nil {
			p.errors = append(p.errors, fmt.Sprintf("failed to parse command: %v", err))
			// maybe break or continue
		}
		commands = append(commands, command)
		time.Sleep(300 * time.Millisecond)
	}
	return &commands
}
func (p *Parser) inSet(tt token.TokenType, set []token.TokenType) bool {
	for _, t := range set {
		if tt == t {
			return true
		}
	}
	return false
}
func (p *Parser) parseIdentifier() (*ast.Identifier, error) {
	fmt.Printf("in parseIdentifier: %v\n", p.curToken)
	identifier := &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
	p.nextToken()                     // Consume the identifier token
	if p.curTokenIs(token.LBRACKET) { // LBRACKET = [
		p.nextToken() // Consume [ & Move to the token inside the brackets

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
	if !isOperator(p.curToken.Type) {
		return &ast.MathExpression{
			Left:     left,
			Operator: token.Token{Type: token.ILLEGAL, Literal: ""},
			Right:    nil, // no right operand
		}, nil
	}
	operator := p.curToken
	p.nextToken() // eat operator
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
	if !isConditionOperator(p.curToken.Type) {
		return nil, fmt.Errorf("parseCondition: expected a comparison operator, got %s", p.peekToken.Type)
	}
	operator := p.curToken
	// fmt.Printf("%v - THIS IS THE OPERATOR I GOT\n\n\n", operator)
	p.nextToken() // eat operator
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
		val := &ast.NumberLiteral{
			Token: p.curToken,
			Value: p.curToken.Literal,
		}
		p.nextToken()
		return val, nil
	case token.PIDENTIFIER:
		return p.parseIdentifier()
	}
	return nil, fmt.Errorf("failed to parse value for token: %v", p.curToken)
}

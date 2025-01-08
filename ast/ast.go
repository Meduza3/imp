package ast

import (
	"fmt"

	"github.com/Meduza3/imp/token"
)

// Node represents an AST node.
type Node interface {
	TokenLiteral() string
	String() string
}

// Command represents a statement in the language.
type Command interface {
	Node
	commandNode()
}

// Expression represents an expression in the language.
type Expression interface {
	Node
	expressionNode()
}

// Holds the entire file
type Program struct {
	Declarations []Declaration
	Commands     []Command // List of commands
}

func (p *Program) String() string {
	var string string
	string += "PROGRAM IS \n"
	for _, decl := range p.Declarations {
		string += decl.String()
	}
	string += "\nBEGIN\n"
	for _, comm := range p.Commands {
		string += comm.String()
	}
	string += "\nEND"
	return string
}

type Declaration struct {
	IsTable     bool
	Pidentifier Pidentifier
	From        NumberLiteral
	To          NumberLiteral
}

func (d *Declaration) String() string {
	if d.IsTable {
		return fmt.Sprintf("%s[%v:%v]", d.Pidentifier.String(), d.From, d.To)
	} else {
		return d.Pidentifier.String()
	}
}

// represtents "expression" in BNF
type MathExpression struct {
	Left     Value
	Operator token.Token
	Right    Value
}

func (me *MathExpression) expressionNode()      {}
func (me *MathExpression) TokenLiteral() string { return me.Operator.Literal }
func (me *MathExpression) String() string {
	return me.Left.String() + me.Operator.Literal + me.Right.String()
}

type Condition struct {
	Left     Value
	Operator token.Token
	Right    Value
}

func (c *Condition) expressionNode()      {}
func (c *Condition) TokenLiteral() string { return c.Operator.Literal }
func (c *Condition) String() string {
	return c.Left.String() + c.Operator.Literal + c.Right.String()
}

type Value interface {
	Expression
	valueNode()
}

type NumberLiteral struct {
	Token token.Token //token.NUM
	Value string      //
}

func (nl *NumberLiteral) expressionNode()      {}
func (nl *NumberLiteral) valueNode()           {}
func (nl *NumberLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NumberLiteral) String() string       { return fmt.Sprintf("%v", nl.Value) }

type ProcCallCommand struct {
	Token token.Token
	Name  Pidentifier
	Args  []Pidentifier
}

func (pcc *ProcCallCommand) commandNode()         {}
func (pcc *ProcCallCommand) TokenLiteral() string { return pcc.Token.Literal }
func (pcc *ProcCallCommand) String() string {
	var string string
	string += pcc.Name.String()
	string += "("
	for i, arg := range pcc.Args {
		if i > 0 {
			string += ", "
		}
		string += arg.String()
	}
	string += ")"
	string += ";"
	return string
}

type AssignCommand struct {
	Identifier     Identifier  // Where will the expression be assigned to?
	Token          token.Token // token.ASSIGN
	MathExpression MathExpression
}

func (ac *AssignCommand) commandNode()         {}
func (ac *AssignCommand) TokenLiteral() string { return ac.Token.Literal }
func (ac *AssignCommand) String() string {
	// Single-value expression
	if ac.MathExpression.Right == nil {
		return fmt.Sprintf("%s := %s; ",
			ac.Identifier.String(),
			ac.MathExpression.Left.String(),
		)
	}

	// Otherwise, typical operator-based expression
	return fmt.Sprintf("%s := %s %s %s; ",
		ac.Identifier.String(),
		ac.MathExpression.Left.String(),
		ac.MathExpression.Operator.Literal,
		ac.MathExpression.Right.String(),
	)
}

type WhileCommand struct {
	Token     token.Token //WHILE
	Condition Condition
	Commands  []Command
}

func (wc *WhileCommand) commandNode()         {}
func (wc *WhileCommand) TokenLiteral() string { return wc.Token.Literal }
func (wc *WhileCommand) String() string {
	var string string
	string += "WHILE "
	string += wc.Condition.String()
	string += " DO "
	string = appendCommands(string, wc.Commands)
	string += " UNTIL "
	string += ";"
	return string
}

type RepeatCommand struct {
	Token     token.Token //REPEAT
	Commands  []Command
	Condition Condition
}

func (rc *RepeatCommand) commandNode()         {}
func (rc *RepeatCommand) TokenLiteral() string { return rc.Token.Literal }
func (rc *RepeatCommand) String() string {
	var string string
	string += "REPEAT "
	string = appendCommands(string, rc.Commands)
	string += " UNTIL "
	string += rc.Condition.String()
	string += ";"
	return string
}

type ForCommand struct {
	Token    token.Token // FOR
	Iterator Pidentifier
	IsDownTo bool // 0: To, 1: Downto
	From     Value
	To       Value
	Commands []Command
}

func (fc *ForCommand) commandNode()         {}
func (fc *ForCommand) TokenLiteral() string { return fc.Token.Literal }
func (fc *ForCommand) String() string {
	var string string
	string += "FOR "
	string += fc.Iterator.Value
	string += " FROM "
	string += fc.From.String()
	if fc.IsDownTo {
		string += " DOWNTO "
	} else {
		string += " TO "
	}
	string = appendCommands(string, fc.Commands)
	string += " ENDFOR"
	return string
}

func appendCommands(string string, commands []Command) string {
	for _, comm := range commands {
		string += comm.String()
	}
	return string
}

type ReadCommand struct {
	Token      token.Token //READ
	Identifier Identifier
}

func (rc *ReadCommand) commandNode()         {}
func (rc *ReadCommand) TokenLiteral() string { return rc.Token.Literal }
func (rc *ReadCommand) String() string {
	return "WRITE " + rc.Identifier.String()
}

type WriteCommand struct {
	Token token.Token //WRITE
	Value Value
}

func (wc *WriteCommand) commandNode()         {}
func (wc *WriteCommand) TokenLiteral() string { return wc.Token.Literal }
func (wc *WriteCommand) String() string {
	if wc.Value != nil {
		return fmt.Sprintf("WRITE %v", wc.Value)
	}
	return "WRITE "
}

type IfCommand struct {
	Token        token.Token //IF
	Condition    Condition
	ThenCommands []Command
	ElseCommands []Command
}

func (ic *IfCommand) commandNode()         {}
func (ic *IfCommand) TokenLiteral() string { return ic.Token.Literal }
func (ic *IfCommand) String() string {
	result := "IF " + ic.Condition.String() + " THEN\n"

	// Then commands
	for _, command := range ic.ThenCommands {
		result += "  " + command.String() + "\n"
	}

	// Else commands
	if len(ic.ElseCommands) > 0 {
		result += "ELSE\n"
		for _, command := range ic.ElseCommands {
			result += "  " + command.String() + "\n"
		}
	}

	result += "ENDIF\n"
	return result
}

type Identifier struct {
	Token token.Token // token.IDENT
	Value string
	Index Expression
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) valueNode()           {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string {
	if i.Index != nil {
		return i.Value + "[" + i.Index.String() + "]"
	}
	return i.Value
}

type Pidentifier struct {
	Token token.Token
	Value string
}

func (pi *Pidentifier) expressionNode()      {}
func (pi *Pidentifier) valueNode()           {}
func (pi *Pidentifier) TokenLiteral() string { return pi.Token.Literal }
func (pi *Pidentifier) String() string       { return pi.Value }

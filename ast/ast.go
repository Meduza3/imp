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

type Program struct {
	Commands []Command
}

func (p *Program) String() string {
	var string string
	for _, comm := range p.Commands {
		if comm != nil {
			string += comm.String()
		}
	}
	return string
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
func (nl *NumberLiteral) String() string       { return fmt.Sprintf("%d", nl.Value) }

type AssignCommand struct {
	Identifier     Identifier  // Where will the expression be assigned to?
	Token          token.Token // token.ASSIGN
	MathExpression MathExpression
}

func (ac *AssignCommand) commandNode()         {}
func (ac *AssignCommand) TokenLiteral() string { return ac.Token.Literal }
func (ac *AssignCommand) String() string {
	// x := a + 2
	return ac.Identifier.String() + " := " + ac.MathExpression.String()
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

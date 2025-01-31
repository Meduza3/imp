package ast

import (
	"fmt"
	"strconv"
	"strings"

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
	Token      token.Token
	Procedures []*Procedure
	Main       *Main
}

func (p *Program) TokenLiteral() string { return p.Token.Literal }
func (p *Program) String() string {
	var string string
	for _, proc := range p.Procedures {
		string += proc.String()
	}
	string += p.Main.String()
	return string
}

type Procedure struct {
	Token        token.Token // PROCEDURE
	ProcHead     ProcHead
	Declarations []Declaration
	Commands     []Command
}

func (p *Procedure) TokenLiteral() string { return p.Token.Literal }
func (p *Procedure) String() string {
	var string string
	string += "PROCEDURE "
	string += p.ProcHead.String()
	string += " IS "
	for _, decl := range p.Declarations {
		string += decl.String() + ", "
	}
	string += " BEGIN "
	for _, comm := range p.Commands {
		string += comm.String()
	}
	string += " END"
	return string
}

type ProcHead struct {
	Token    token.Token
	Name     Pidentifier
	ArgsDecl []ArgDecl
}

func (ph *ProcHead) TokenLiteral() string { return ph.Token.Literal }
func (ph *ProcHead) String() string {
	var string string
	string += ph.Name.String()
	string += "("
	for i, arg := range ph.ArgsDecl {
		if i > 0 {
			string += ", "
		}
		string += arg.String()
	}
	string += ")"
	return string
}

type ArgDecl struct {
	Token   token.Token
	IsTable bool
	Name    Pidentifier
}

func (ad *ArgDecl) TokenLiteral() string { return ad.Token.Literal }
func (ad *ArgDecl) String() string {
	if ad.IsTable {
		return "T " + ad.Name.String()
	} else {
		return ad.Name.String()
	}
}

// Holds the Main
type Main struct {
	Token        token.Token
	Declarations []Declaration
	Commands     []Command // List of commands
}

func (m *Main) TokenLiteral() string { return m.Token.Literal }
func (m *Main) String() string {
	var s strings.Builder

	// Append the program header
	s.WriteString("PROGRAM IS\n")

	// Collect all declaration strings
	declStrings := make([]string, 0, len(m.Declarations))
	for _, decl := range m.Declarations {
		declStrings = append(declStrings, decl.String())
	}

	// Join declarations with ", " and append
	s.WriteString(strings.Join(declStrings, ", "))
	s.WriteString("\nBEGIN\n")

	// Append all command strings
	for _, comm := range m.Commands {
		s.WriteString(comm.String())
	}

	// Append the program footer
	s.WriteString("\nEND")

	return s.String()
}

type Declaration struct {
	IsTable     bool
	Pidentifier Pidentifier
	From        NumberLiteral
	To          NumberLiteral
}

func (d *Declaration) String() string {
	if d.IsTable {
		return fmt.Sprintf("%s[%v:%v]", d.Pidentifier.String(), d.From.String(), d.To.String())
	} else {
		return d.Pidentifier.String()
	}
}
func (d *Declaration) TokenLiteral() string {
	return d.Pidentifier.Token.Literal
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
	if me.Right == nil {
		// single operand, no operator
		return me.Left.String()
	}
	// operator-based expression
	return fmt.Sprintf("%s %s %s", me.Left.String(), me.Operator.Literal, me.Right.String())
}

type Condition struct {
	Left     Value
	Operator token.Token
	Right    Value
}

func (c *Condition) expressionNode()      {}
func (c *Condition) TokenLiteral() string { return c.Operator.Literal }
func (c *Condition) String() string {
	return c.Left.String() + " " + c.Operator.Literal + " " + c.Right.String()
}

type Value interface {
	Expression
	valueNode()
	String() string
}

type NumberLiteral struct {
	Token token.Token //token.NUM
	Value string      //
}

func (nl *NumberLiteral) expressionNode()      {}
func (nl *NumberLiteral) valueNode()           {}
func (nl *NumberLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NumberLiteral) String() string {
	int, err := strconv.Atoi(nl.TokenLiteral())
	if err != nil {
		return "oops"
	}
	return fmt.Sprintf("%d", int)
}

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
	return "READ " + rc.Identifier.String() + ";"
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
	Token   token.Token // token.IDENT
	Value   string
	IsTable bool
	Index   string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) valueNode()           {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string {
	if i.IsTable {
		return fmt.Sprintf("%s[%s]", i.Value, i.Index)
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

package ast

import (
	"bytes"
	"strings"

	"github.com/Meduza3/imp/token"
)

// Node represents an AST node.
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement represents a statement in the language.
type Statement interface {
	Node
	statementNode()
}

// Expression represents an expression in the language.
type Expression interface {
	Node
	expressionNode()
}

// ProgramAll is the root node of the AST, representing the entire program.
type ProgramAll struct {
	Procedures []*Procedure
	Main       *Main
}

func (p *ProgramAll) TokenLiteral() string {
	if len(p.Procedures) > 0 {
		return p.Procedures[0].TokenLiteral()
	}
	if p.Main != nil {
		return p.Main.TokenLiteral()
	}
	return ""
}

func (p *ProgramAll) String() string {
	var out bytes.Buffer

	for _, proc := range p.Procedures {
		out.WriteString(proc.String())
	}
	if p.Main != nil {
		out.WriteString(p.Main.String())
	}

	return out.String()
}

// Procedure represents a procedure declaration.
type Procedure struct {
	Token        *token.Token // The 'PROCEDURE' token
	Head         *ProcHead
	Declarations []Declaration
	Commands     []Statement
}

func (p *Procedure) statementNode() {}
func (p *Procedure) TokenLiteral() string {
	return p.Token.Literal
}

func (p *Procedure) String() string {
	var out bytes.Buffer

	out.WriteString("PROCEDURE ")
	out.WriteString(p.Head.String())
	out.WriteString(" IS ")
	if len(p.Declarations) > 0 {
		out.WriteString("Declarations: ")
		for _, decl := range p.Declarations {
			out.WriteString(decl.String())
			out.WriteString(", ")
		}
	}
	out.WriteString("BEGIN ")
	for _, cmd := range p.Commands {
		out.WriteString(cmd.String())
	}
	out.WriteString(" END")

	return out.String()
}

// Main represents the main program block.
type Main struct {
	Token        *token.Token // The 'PROGRAM' token
	Declarations []Declaration
	Commands     []Statement
}

func (m *Main) statementNode() {}
func (m *Main) TokenLiteral() string {
	return m.Token.Literal
}

func (m *Main) String() string {
	var out bytes.Buffer

	out.WriteString("PROGRAM IS ")
	if len(m.Declarations) > 0 {
		out.WriteString("Declarations: ")
		for _, decl := range m.Declarations {
			out.WriteString(decl.String())
			out.WriteString(", ")
		}
	}
	out.WriteString("BEGIN ")
	for _, cmd := range m.Commands {
		out.WriteString(cmd.String())
	}
	out.WriteString(" END")

	return out.String()
}

// Declaration represents a variable or array declaration.
type Declaration interface {
	Node
	declarationNode()
}

// VarDeclaration represents a variable declaration.
type VarDeclaration struct {
	Token *token.Token // The identifier token
	Name  *Identifier
}

func (vd *VarDeclaration) declarationNode() {}
func (vd *VarDeclaration) TokenLiteral() string {
	return vd.Token.Literal
}

func (vd *VarDeclaration) String() string {
	return vd.Name.String()
}

// ArrayDeclaration represents an array declaration.
type ArrayDeclaration struct {
	Token *token.Token // The identifier token
	Name  *Identifier
	From  Expression
	To    Expression
}

func (ad *ArrayDeclaration) declarationNode() {}
func (ad *ArrayDeclaration) TokenLiteral() string {
	return ad.Token.Literal
}

func (ad *ArrayDeclaration) String() string {
	var out bytes.Buffer

	out.WriteString(ad.Name.String())
	out.WriteString("[")
	out.WriteString(ad.From.String())
	out.WriteString(":")
	out.WriteString(ad.To.String())
	out.WriteString("]")

	return out.String()
}

// ProcHead represents the head of a procedure with its arguments.
type ProcHead struct {
	Name     *Identifier
	ArgsDecl []*ArgDeclaration
}

func (ph *ProcHead) TokenLiteral() string {
	return ph.Name.TokenLiteral()
}

func (ph *ProcHead) String() string {
	var out bytes.Buffer

	out.WriteString(ph.Name.String())
	out.WriteString("(")
	args := make([]string, len(ph.ArgsDecl))
	for i, arg := range ph.ArgsDecl {
		args[i] = arg.String()
	}
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

// ArgDeclaration represents an argument declaration in a procedure.
type ArgDeclaration struct {
	Token *token.Token // The identifier token
	Name  *Identifier
	Type  string // Optional type information
}

func (ad *ArgDeclaration) TokenLiteral() string {
	return ad.Token.Literal
}

func (ad *ArgDeclaration) String() string {
	if ad.Type != "" {
		return ad.Type + " " + ad.Name.String()
	}
	return ad.Name.String()
}

// Commands represents a sequence of commands/statements.
type Commands struct {
	Statements []Statement
}

func (c *Commands) statementNode() {}
func (c *Commands) TokenLiteral() string {
	if len(c.Statements) > 0 {
		return c.Statements[0].TokenLiteral()
	}
	return ""
}

func (c *Commands) String() string {
	var out bytes.Buffer

	for _, stmt := range c.Statements {
		out.WriteString(stmt.String())
	}

	return out.String()
}

// AssignmentStatement represents an assignment operation.
type AssignmentStatement struct {
	Token *token.Token // The ':=' token
	Left  Expression   // Identifier or ArrayAccess
	Right Expression
}

func (as *AssignmentStatement) statementNode() {}
func (as *AssignmentStatement) TokenLiteral() string {
	return as.Token.Literal
}

func (as *AssignmentStatement) String() string {
	var out bytes.Buffer

	out.WriteString(as.Left.String())
	out.WriteString(" := ")
	out.WriteString(as.Right.String())
	out.WriteString(";")

	return out.String()
}

// IfStatement represents an if-else statement.
type IfStatement struct {
	Token       *token.Token // The 'IF' token
	Condition   Expression
	Consequence []Statement
	Alternative []Statement
}

func (is *IfStatement) statementNode() {}
func (is *IfStatement) TokenLiteral() string {
	return is.Token.Literal
}

func (is *IfStatement) String() string {
	var out bytes.Buffer

	out.WriteString("IF ")
	out.WriteString(is.Condition.String())
	out.WriteString(" THEN ")
	for _, stmt := range is.Consequence {
		out.WriteString(stmt.String())
	}
	if len(is.Alternative) > 0 {
		out.WriteString(" ELSE ")
		for _, stmt := range is.Alternative {
			out.WriteString(stmt.String())
		}
	}
	out.WriteString(" ENDIF")

	return out.String()
}

// WhileStatement represents a while loop.
type WhileStatement struct {
	Token     *token.Token // The 'WHILE' token
	Condition Expression
	Body      []Statement
}

func (ws *WhileStatement) statementNode() {}
func (ws *WhileStatement) TokenLiteral() string {
	return ws.Token.Literal
}

func (ws *WhileStatement) String() string {
	var out bytes.Buffer

	out.WriteString("WHILE ")
	out.WriteString(ws.Condition.String())
	out.WriteString(" DO ")
	for _, stmt := range ws.Body {
		out.WriteString(stmt.String())
	}
	out.WriteString(" ENDWHILE")

	return out.String()
}

// RepeatStatement represents a repeat-until loop.
type RepeatStatement struct {
	Token     *token.Token // The 'REPEAT' token
	Body      []Statement
	Condition Expression
}

func (rs *RepeatStatement) statementNode() {}
func (rs *RepeatStatement) TokenLiteral() string {
	return rs.Token.Literal
}

func (rs *RepeatStatement) String() string {
	var out bytes.Buffer

	out.WriteString("REPEAT ")
	for _, stmt := range rs.Body {
		out.WriteString(stmt.String())
	}
	out.WriteString(" UNTIL ")
	out.WriteString(rs.Condition.String())
	out.WriteString(";")

	return out.String()
}

// ForStatement represents a for loop.
type ForStatement struct {
	Token    *token.Token // The 'FOR' token
	Iterator *Identifier
	From     Expression
	To       Expression
	DownTo   bool
	Body     []Statement
}

func (fs *ForStatement) statementNode() {}
func (fs *ForStatement) TokenLiteral() string {
	return fs.Token.Literal
}

func (fs *ForStatement) String() string {
	var out bytes.Buffer

	out.WriteString("FOR ")
	out.WriteString(fs.Iterator.String())
	out.WriteString(" FROM ")
	out.WriteString(fs.From.String())
	if fs.DownTo {
		out.WriteString(" DOWNTO ")
	} else {
		out.WriteString(" TO ")
	}
	out.WriteString(fs.To.String())
	out.WriteString(" DO ")
	for _, stmt := range fs.Body {
		out.WriteString(stmt.String())
	}
	out.WriteString(" ENDFOR")

	return out.String()
}

// ProcedureCallStatement represents a procedure call.
type ProcedureCallStatement struct {
	Token     *token.Token // The procedure name token
	Name      *Identifier
	Arguments []Expression
}

func (pcs *ProcedureCallStatement) statementNode() {}
func (pcs *ProcedureCallStatement) TokenLiteral() string {
	return pcs.Token.Literal
}

func (pcs *ProcedureCallStatement) String() string {
	var out bytes.Buffer

	out.WriteString(pcs.Name.String())
	out.WriteString("(")
	args := make([]string, len(pcs.Arguments))
	for i, arg := range pcs.Arguments {
		args[i] = arg.String()
	}
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(");")

	return out.String()
}

// ReadStatement represents a read operation.
type ReadStatement struct {
	Token      *token.Token // The 'READ' token
	Identifier Expression   // Identifier or ArrayAccess
}

func (rs *ReadStatement) statementNode() {}
func (rs *ReadStatement) TokenLiteral() string {
	return rs.Token.Literal
}

func (rs *ReadStatement) String() string {
	var out bytes.Buffer

	out.WriteString("READ ")
	out.WriteString(rs.Identifier.String())
	out.WriteString(";")

	return out.String()
}

// WriteStatement represents a write operation.
type WriteStatement struct {
	Token *token.Token // The 'WRITE' token
	Value Expression
}

func (ws *WriteStatement) statementNode() {}
func (ws *WriteStatement) TokenLiteral() string {
	return ws.Token.Literal
}

func (ws *WriteStatement) String() string {
	var out bytes.Buffer

	out.WriteString("WRITE ")
	out.WriteString(ws.Value.String())
	out.WriteString(";")

	return out.String()
}

// Expression implementations

// Identifier represents an identifier or variable name.
type Identifier struct {
	Token *token.Token // The identifier token
	Value string
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}

func (i *Identifier) String() string {
	return i.Value
}

// IntegerLiteral represents an integer value.
type IntegerLiteral struct {
	Token *token.Token // The integer token
	Value int64
}

func (il *IntegerLiteral) expressionNode() {}
func (il *IntegerLiteral) TokenLiteral() string {
	return il.Token.Literal
}

func (il *IntegerLiteral) String() string {
	return il.Token.Literal
}

// ArrayAccess represents access to an array element.
type ArrayAccess struct {
	Token     *token.Token // The '[' token
	ArrayName *Identifier
	Index     Expression
}

func (aa *ArrayAccess) expressionNode() {}
func (aa *ArrayAccess) TokenLiteral() string {
	return aa.Token.Literal
}

func (aa *ArrayAccess) String() string {
	var out bytes.Buffer

	out.WriteString(aa.ArrayName.String())
	out.WriteString("[")
	out.WriteString(aa.Index.String())
	out.WriteString("]")

	return out.String()
}

// BinaryExpression represents a binary operation (e.g., +, -, *, /, %, =, !=, >, <, >=, <=).
type BinaryExpression struct {
	Token    *token.Token // The operator token
	Left     Expression
	Operator string
	Right    Expression
}

func (be *BinaryExpression) expressionNode() {}
func (be *BinaryExpression) TokenLiteral() string {
	return be.Token.Literal
}

func (be *BinaryExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(be.Left.String())
	out.WriteString(" ")
	out.WriteString(be.Operator)
	out.WriteString(" ")
	out.WriteString(be.Right.String())
	out.WriteString(")")

	return out.String()
}

// The following are helper functions and types to represent values and conditions based on the grammar.

// Value represents a numerical value or an identifier.
type Value struct {
	Token *token.Token // The token representing the value
	Expr  Expression   // IntegerLiteral or Identifier or ArrayAccess
}

func (v *Value) expressionNode() {}
func (v *Value) TokenLiteral() string {
	return v.Token.Literal
}

func (v *Value) String() string {
	return v.Expr.String()
}

// Condition represents a conditional expression.
type Condition struct {
	Left     Expression
	Operator string
	Right    Expression
}

func (c *Condition) expressionNode() {}
func (c *Condition) TokenLiteral() string {
	return c.Operator
}

func (c *Condition) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(c.Left.String())
	out.WriteString(" ")
	out.WriteString(c.Operator)
	out.WriteString(" ")
	out.WriteString(c.Right.String())
	out.WriteString(")")

	return out.String()
}

// ProcCall represents a procedure call within an expression.
type ProcCall struct {
	Token     token.Token // The procedure name token
	Name      *Identifier
	Arguments []Expression
}

func (pc *ProcCall) expressionNode() {}
func (pc *ProcCall) TokenLiteral() string {
	return pc.Token.Literal
}

func (pc *ProcCall) String() string {
	var out bytes.Buffer

	out.WriteString(pc.Name.String())
	out.WriteString("(")
	args := make([]string, len(pc.Arguments))
	for i, arg := range pc.Arguments {
		args[i] = arg.String()
	}
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

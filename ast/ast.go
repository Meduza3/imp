package ast

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

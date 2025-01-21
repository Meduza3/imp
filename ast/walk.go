package ast

import "fmt"

func Walk(node Node, visit func(Node)) {
	if node == nil {
		return
	}

	// Perform an action for the current node
	visit(node)

	// Descend into children based on concrete type:
	switch n := node.(type) {

	case *Program:
		// Walk procedures
		for _, proc := range n.Procedures {
			Walk(proc, visit)
		}
		// Walk main
		Walk(n.Main, visit)

	case *Procedure:
		// Walk procedure head (proc_head)
		Walk(&n.ProcHead, visit)

		// Walk declarations
		for _, decl := range n.Declarations {
			Walk(&decl, visit)
		}

		// Walk commands
		for _, cmd := range n.Commands {
			Walk(cmd, visit)
		}

	case *ProcHead:
		// Walk each argument declaration
		for i := range n.ArgsDecl {
			Walk(&n.ArgsDecl[i], visit)
		}

	case *ArgDecl:
		// Contains no child nodes besides the name (Pidentifier).
		// If you wanted to visit the name, do:
		Walk(&n.Name, visit)

	case *Main:
		// Walk declarations
		for _, decl := range n.Declarations {
			Walk(&decl, visit)
		}

		// Walk commands
		for _, cmd := range n.Commands {
			Walk(cmd, visit)
		}

	case *Declaration:
		// Pidentifier and the NumberLiterals are children.
		Walk(&n.Pidentifier, visit)
		if n.IsTable {
			Walk(&n.From, visit)
			Walk(&n.To, visit)
		}

	// --- Commands ---

	case *AssignCommand:
		Walk(&n.Identifier, visit)
		// The MathExpression includes two Value children (Left and Right).
		// We can also walk the MathExpression if we treat it as a separate node.
		Walk(&n.MathExpression, visit)

	case *ProcCallCommand:
		// The Pidentifier is n.Name, plus the arguments
		Walk(&n.Name, visit)
		for i := range n.Args {
			Walk(&n.Args[i], visit)
		}

	case *WhileCommand:
		Walk(&n.Condition, visit)
		for _, cmd := range n.Commands {
			Walk(cmd, visit)
		}

	case *RepeatCommand:
		for _, cmd := range n.Commands {
			Walk(cmd, visit)
		}
		Walk(&n.Condition, visit)

	case *ForCommand:
		Walk(&n.Iterator, visit)
		Walk(n.From, visit)
		Walk(n.To, visit)
		for _, cmd := range n.Commands {
			Walk(cmd, visit)
		}

	case *ReadCommand:
		Walk(&n.Identifier, visit)

	case *WriteCommand:
		Walk(n.Value, visit)

	case *IfCommand:
		Walk(&n.Condition, visit)
		for _, cmd := range n.ThenCommands {
			Walk(cmd, visit)
		}
		for _, cmd := range n.ElseCommands {
			Walk(cmd, visit)
		}

	// --- Expressions ---

	case *MathExpression:
		if n.Left != nil {
			Walk(n.Left, visit)
		}
		// `n.Operator` is just a token, not a node, so we skip it
		if n.Right != nil {
			Walk(n.Right, visit)
		}

	case *Condition:
		if n.Left != nil {
			Walk(n.Left, visit)
		}
		// `n.Operator` is just a token, not a node
		if n.Right != nil {
			Walk(n.Right, visit)
		}

	// --- Values ---

	case *NumberLiteral:
		// Just a leaf node, no children to walk

	case *Identifier:
		// Possibly a leaf, unless you consider the index a child node.
		// If the index is also an identifier or expression, walk it here.

	case *Pidentifier:
		// Leaf node

	default:
		// Unknown or unhandled node
		fmt.Printf("Walk: unhandled node type %T\n", n)
	}
}

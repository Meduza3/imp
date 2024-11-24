package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/Meduza3/imp/lexer"
	"github.com/Meduza3/imp/parser"
)

const PROMPT = ">:] imp>"

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	for {
		fmt.Print(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			break
		}
		line := scanner.Text()

		// Pass the line to the parser
		lexer := lexer.New(line)
		p := parser.YyNewParser()
		if p.Parse(lexer) != 0 {
			fmt.Fprintln(out, "Error: Invalid syntax")
		} else {
			fmt.Fprintln(out, "Parsed successfully")
		}
	}
}

func StartFile(filePath string, out io.Writer) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(out, "Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	// Read the file content
	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Fprintf(out, "Error reading file: %v\n", err)
		return
	}

	// Parse the content
	lex := lexer.New(string(content))
	p := parser.YyNewParser()

	// Use a struct to capture the parse result (e.g., AST)
	var ast interface{}
	if p.Parse(lex) != 0 {
		fmt.Fprintln(out, "Error: Invalid syntax")
	} else {
		// Assuming the parser can produce an AST or some structured output
		ast = lex.GetAST()
		fmt.Fprintln(out, "Parsed successfully")
		VisualizeAST(ast, out)
	}
}

type Node interface {
	String() string
	Children() []Node
}

func VisualizeAST(node interface{}, out io.Writer) {
	if root, ok := node.(Node); ok {
		printNode(root, "", true, out)
	} else {
		fmt.Fprintln(out, "No valid AST to visualize")
	}
}

func printNode(node Node, prefix string, isTail bool, out io.Writer) {
	if node == nil {
		return
	}

	// Print current node
	fmt.Fprintf(out, "%s%s%s\n", prefix, "└── ", node.String())

	// Print children
	children := node.Children()
	for i, child := range children {
		newPrefix := prefix
		if isTail {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}
		printNode(child, newPrefix, i == len(children)-1, out)
	}
}

type ASTNode struct {
	Type     string
	Value    string
	Children []*ASTNode
}

func (n *ASTNode) String() string {
	return fmt.Sprintf("%s: %s", n.Type, n.Value)
}

func (n *ASTNode) Children() []Node {
	var nodes []Node
	for _, child := range n.Children {
		nodes = append(nodes, child)
	}
	return nodes
}

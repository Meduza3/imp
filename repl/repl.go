package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/Meduza3/imp/ast"
	"github.com/Meduza3/imp/lexer"
	"github.com/Meduza3/imp/parser"
	"github.com/Meduza3/imp/tac"
	"github.com/Meduza3/imp/token"
	"github.com/Meduza3/imp/translator"
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

		lexer := lexer.New(line)
		p := parser.New(lexer)
		program := p.ParseProgram()
		ast.NewPrinter().Print(program)
	}
}

func StartParsing(in io.Reader, out io.Writer) {
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
		p := parser.New(lexer)
		program := p.ParseProgram()
		fmt.Println("parsed program")
		io.WriteString(out, program.String())
		io.WriteString(out, fmt.Sprintf("%#+v", program))
		io.WriteString(out, "\n")
		for _, err := range p.Errors() {
			io.WriteString(out, err)
			fmt.Println()
		}
	}
}

func StartParsingFile(filepath string, out io.Writer) {
	// Open the file
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Fprintf(out, "Error opening file %s: %v\n", filepath, err)
		return
	}
	defer file.Close()

	// Read the entire file content
	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Fprintf(out, "Error reading file %s: %v\n", filepath, err)
		return
	}

	// Pass the entire content to the parser
	l := lexer.New(string(content))
	l2 := lexer.New(string(content))
	for tok := l2.NextToken(); tok.Type != token.EOF; tok = l2.NextToken() {
		fmt.Println(tok)
	}
	p := parser.New(l)
	program := p.ParseProgram()

	// Write the program's string representation to output
	// io.WriteString(out, program.String())
	// io.WriteString(out, "\n")

	// printer := ast.NewPrinter()
	// // formattedCode := printer.Print(program)

	// // // io.WriteString(out, "Pretty-Printed Code:\n")
	// // // io.WriteString(out, formattedCode)
	// // // for _, err := range p.Errors() {
	// // // 	io.WriteString(out, err)
	// // // 	fmt.Println()
	// // // }
	ast.Walk(program, func(n ast.Node) {
		fmt.Printf("Visiting node of type %T: %v\n", n, n)
	})
}

func StartGeneratingFile(filepath string, out io.Writer) {
	// Open the file
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Fprintf(out, "Error opening file %s: %v\n", filepath, err)
		return
	}
	defer file.Close()

	// Read the entire file content
	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Fprintf(out, "Error reading file %s: %v\n", filepath, err)
		return
	}

	// Pass the entire content to the parser
	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	// Write the program's string representation to output
	// io.WriteString(out, program.String())
	// io.WriteString(out, "\n")

	// printer := ast.NewPrinter()
	// // formattedCode := printer.Print(program)

	// // // io.WriteString(out, "Pretty-Printed Code:\n")
	// // // io.WriteString(out, formattedCode)
	// // // for _, err := range p.Errors() {
	// // // 	io.WriteString(out, err)
	// // // 	fmt.Println()
	// // // }
	g := tac.NewGenerator()
	g.Generate(program)
	g.Instructions = tac.MergeLabelOnlyInstructions(g.Instructions)
	symbolTable := g.GetSymbolTable()
	fmt.Println("==SYMBOL TABLE==")
	for key, value := range symbolTable.Table {
		fmt.Printf("PROCEDURE %s\n", key)
		for keyy, valuee := range value {
			fmt.Printf("name=%q symbol=%v\n", keyy, valuee)
		}
	}
	blocks := tac.SplitIntoBasicBlocks(g.Instructions)
	blocks = tac.BuildFlowGraph(blocks)
	// for _, block := range blocks {
	// 	fmt.Printf("Block %d connections:\n", block.ID)
	// 	fmt.Printf("  Predecessors: %v\n", tac.GetBlockIDs(block.Predecessors))
	// 	fmt.Printf("  Successors: %v\n", tac.GetBlockIDs(block.Successors))
	// }
	line := 0
	for _, block := range blocks {
		// fmt.Println()
		// fmt.Printf("Block %d\n", block.ID)
		for _, instr := range block.Instructions {
			fmt.Printf("%03d: %s\n", line, instr.String())
			line++
		}
	}
	translator := translator.New(*g.SymbolTable)
	for _, table := range translator.St.Table {
		for key, value := range table {
			fmt.Printf("%s: %v\n", key, value)
		}
	}
	fmt.Println("TRANSLATED: ")
	translator.Translate(g.Instructions)
	line = 0
	for _, instr := range translator.Output {
		fmt.Printf("%03d: %s\n", line, instr.String())
		line++
	}
	for _, err := range translator.Errors() {
		fmt.Println(err)
	}
}

func StartFile(filepath string, out io.Writer) {
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Fprintf(out, "Error opening file %s: %v\n", filepath, err)
		return
	}
	defer file.Close()

	// Read the entire file content
	content, err := io.ReadAll(file)
	fmt.Println("# reading file")
	if err != nil {
		fmt.Fprintf(out, "Error reading file %s: %v\n", filepath, err)
		return
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	fmt.Print("# parsing program...		")
	program := p.ParseProgram()
	fmt.Println("parsed. ")
	for _, err := range p.Errors() {
		fmt.Printf("# parse Error: %s\n", err)
	}
	g := tac.NewGenerator()
	fmt.Printf("# generating TAC...		")
	g.Generate(program)
	fmt.Println("generated. ")
	blocks := tac.SplitIntoBasicBlocks(g.Instructions)
	blocks = tac.BuildFlowGraph(blocks)
	// for _, block := range blocks {
	// 	fmt.Printf("Block %d connections:\n", block.ID)
	// 	fmt.Printf("  Predecessors: %v\n", tac.GetBlockIDs(block.Predecessors))
	// 	fmt.Printf("  Successors: %v\n", tac.GetBlockIDs(block.Successors))
	// }
	// for _, block := range blocks {
	// 	// fmt.Println()
	// 	// fmt.Printf("Block %d\n", block.ID)
	// 	for _, instr := range block.Instructions {
	// 		fmt.Printf("# %s\n", instr.String())
	// 	}
	// }
	g.Instructions = tac.MergeLabelOnlyInstructions(g.Instructions)

	blocks = tac.SplitIntoBasicBlocks(g.Instructions)
	blocks = tac.BuildFlowGraph(blocks)
	translator := translator.New(*g.SymbolTable)
	fmt.Println("# Translating TAC...		")
	translator.Translate(g.Instructions)
	for _, instr := range translator.Output {
		fmt.Printf("%s\n", instr.String())
	}
	for _, err := range translator.Errors() {
		fmt.Printf("# ERROR: %s\n", err)
	}
	translator.St.Display(out, "# ")
}

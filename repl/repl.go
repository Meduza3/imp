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
	// symbolTable := g.GetSymbolTable()
	// for key, value := range symbolTable.ProcedureTables {
	// 	fmt.Printf("PROCEDURE %s\n", key)
	// 	for keyy, valuee := range value {
	// 		fmt.Printf("name=%q symbol=%v\n", keyy, valuee)
	// 	}
	// }
	// fmt.Printf("MAIN\n")
	// for key, value := range symbolTable.MainTable {
	// 	fmt.Printf("name=%q symbol=%v\n", key, value)
	// }

	for _, instr := range g.Instructions {
		fmt.Println(instr.String())
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
	_, err = io.ReadAll(file)
	if err != nil {
		fmt.Fprintf(out, "Error reading file %s: %v\n", filepath, err)
		return
	}
}

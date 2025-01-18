package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/Meduza3/imp/ast"
	"github.com/Meduza3/imp/ir"
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

		lexer := lexer.New(line)
		p := parser.New(lexer)
		program := p.ParseProgram()
		compiler := ir.NewCodeGenerator()
		compiler.GenerateProgram(*program)
		// if err != nil {
		// 	io.WriteString(out, "compiler error: "+err.Error())
		// }
		bytecode := compiler.GetInstructions()
		for _, ins := range bytecode {
			io.WriteString(out, ins.String()+"\n")
		}
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
	printer := ast.NewPrinter()
	formattedCode := printer.Print(program)

	io.WriteString(out, "Pretty-Printed Code:\n")
	io.WriteString(out, formattedCode)
	for _, err := range p.Errors() {
		io.WriteString(out, err)
		fmt.Println()
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
	if err != nil {
		fmt.Fprintf(out, "Error reading file %s: %v\n", filepath, err)
		return
	}

	// Pass the entire content to the parser
	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()
	compiler := ir.NewCodeGenerator()

	compiler.GenerateProgram(*program)
	// if err != nil {
	// 	io.WriteString(out, "compiler error: "+err.Error())
	// }
	bytecode := compiler.GetInstructions()
	for _, ins := range bytecode {
		io.WriteString(out, ins.String()+"\n")
	}
	for _, symbol := range compiler.GetSymbolTable() {
		fmt.Printf("%+v\n", symbol)
	}
	cfgs, mainBlocks := ir.BuildAllProceduresCFG(bytecode)
	// Display them
	ir.PrintAllProceduresCFG(cfgs, mainBlocks)

}

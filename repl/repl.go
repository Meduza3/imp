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

	if p.Parse(lex) != 0 {
		fmt.Fprintln(out, "Error: Invalid syntax")
	} else {
		// Assuming the parser can produce an AST or some structured output
		fmt.Fprintln(out, "Parsed successfully")
	}
}

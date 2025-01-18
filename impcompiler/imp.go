package imp

import (
	"fmt"
	"io"

	"github.com/Meduza3/imp/ir"
	"github.com/Meduza3/imp/lexer"
	"github.com/Meduza3/imp/parser"
	"github.com/Meduza3/imp/symboltable"
)

type IMP struct {
	SymbolTable   symboltable.SymbolTable
	lexer         lexer.Lexer
	parser        parser.Parser
	codeGenerator ir.CodeGenerator
}

func NewIMP(input io.Reader) *IMP {
	inputBytes, err := io.ReadAll(input)
	if err != nil {
		fmt.Println("Error reading input:", err)
		return nil
	}

	inputString := string(inputBytes)
	symbolTable := symboltable.New()
	lexer := lexer.New(inputString, symbolTable)
	parser := parser.New(lexer, symbolTable)
	codeGenerator := ir.NewCodeGenerator(parser, symbolTable)
	program := parser.ParseProgram()
	codeGenerator.GenerateProgram(*program)
}

package compiler

import (
	"fmt"
	"strconv"

	"github.com/Meduza3/imp/ast"
	"github.com/Meduza3/imp/code"
)

type Compiler struct {
	instructions []code.Instruction
}

func New() *Compiler {
	return &Compiler{
		instructions: []code.Instruction{},
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, p := range node.Procedures {
			err := c.Compile(&p)
			if err != nil {
				return err
			}
		}
		err := c.Compile(&node.Main)
		if err != nil {
			return err
		}
	case *ast.Main:
		// for _, decl := range node.Declarations {
		// 	err := c.compileDeclaration(&decl)
		// 	if err != nil {
		// 		return err
		// 	}
		// }
		for _, comm := range node.Commands {
			err := c.Compile(comm)
			if err != nil {
				return err
			}
		}
	case *ast.WriteCommand:
		int, _ := strconv.Atoi(node.Value.String()) //later add support for identifier values
		c.emit(code.PUT, int64(int))
	}
	return nil
}

func (c *Compiler) emit(op code.Opcode, operands ...int64) (int, error) {
	ins, err := code.Make(op, operands...)
	if err != nil {
		return 0, fmt.Errorf("failed to make instruction for %v(%v): %v", op, operands, err)
	}
	pos := c.addInstruction(*ins)
	return pos, nil
}

func (c *Compiler) addInstruction(ins code.Instruction) int {
	posNewInstructon := len(c.instructions)
	c.instructions = append(c.instructions, ins)
	return posNewInstructon
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
	}
}

type Bytecode struct {
	Instructions []code.Instruction
}

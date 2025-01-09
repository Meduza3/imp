package compiler

import (
	"fmt"
	"strconv"

	"github.com/Meduza3/imp/ast"
	"github.com/Meduza3/imp/code"
	"github.com/Meduza3/imp/token"
)

const TEMP int64 = 420

type Compiler struct {
	instructions []code.Instruction
	addresses    map[string]int
}

func New() *Compiler {
	return &Compiler{
		instructions: []code.Instruction{},
		addresses:    map[string]int{},
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
		c.emit(code.HALT)
	case *ast.Main:
		for _, decl := range node.Declarations {
			err := c.compileDeclaration(&decl)
			if err != nil {
				return err
			}
		}
		for _, comm := range node.Commands {
			err := c.Compile(comm)
			if err != nil {
				return err
			}
		}
	case *ast.RepeatCommand:
		return c.compileRepeatCommand(node)
	case *ast.AssignCommand:
		err := c.Compile(&node.MathExpression)
		if err != nil {
			return err
		}
		addr, err := c.getAddr(node.Identifier.String())
		if err != nil {
			return err
		}
		c.emit(code.STORE, int64(addr))
	case *ast.MathExpression:
		switch node.Operator.Type {
		case token.PLUS:
			err := c.compileAddition(*node)
			if err != nil {
				return err
			}
		case token.MINUS:
			err := c.compileSubtraction(*node)
			if err != nil {
				return err
			}
		case token.MULT:
			err := c.compileMultiplication(*node)
			if err != nil {
				return err
			}
		case token.DIVIDE:
			err := c.compileDivision(*node)
			if err != nil {
				return err
			}
		case token.MODULO:
			err := c.compileModulo(*node)
			if err != nil {
				return err
			}
		case token.ILLEGAL:
			// TODO: single value expression
		}
	case *ast.WriteCommand:
		int, _ := strconv.Atoi(node.Value.String()) //later add support for identifier values
		c.emit(code.PUT, int64(int))
	case *ast.ReadCommand:
		int, _ := strconv.Atoi(node.Identifier.String()) //later add support for identifier values
		c.emit(code.GET, int64(int))
	}
	return nil
}

func (c *Compiler) compileRepeatCommand(rc *ast.RepeatCommand) error {
	startAddr := len(c.instructions)
	for _, cmd := range rc.Commands {
		err := c.Compile(cmd)
		if err != nil {
			return err
		}
	}

	operator := rc.Condition.Operator.Type

	conditionCompiler := func(jumpBack int) error {
		switch operator {
		case token.GR: // >
			// Jeśli p0 > 0, przejdź do następnej instrukcji (kontynuuj)
			// W przeciwnym razie, skocz do początku pętli
			proceedAddr := 1
			c.emit(code.JPOS, int64(proceedAddr)) // Jeśli p0 > 0, przejdź dalej
			c.emit(code.JUMP, int64(jumpBack))    // W przeciwnym razie, skocz do startAddr

		case token.GEQ: // >=
			// Jeśli p0 > 0 lub p0 == 0, przejdź do następnej instrukcji (kontynuuj)
			// W przeciwnym razie, skocz do początku pętli
			proceedAddr := 2
			c.emit(code.JPOS, int64(proceedAddr))  // Jeśli p0 > 0, przejdź dalej
			c.emit(code.JZERO, int64(proceedAddr)) // Jeśli p0 == 0, przejdź dalej
			c.emit(code.JUMP, int64(jumpBack))     // W przeciwnym razie, skocz do startAddr

		case token.EQUALS: // ==
			// Jeśli p0 == 0, przejdź do następnej instrukcji (kontynuuj)
			// W przeciwnym razie, skocz do początku pętli
			proceedAddr := 1
			c.emit(code.JZERO, int64(proceedAddr)) // Jeśli p0 == 0, przejdź dalej
			c.emit(code.JUMP, int64(jumpBack))     // W przeciwnym razie, skocz do startAddr

		case token.LEQ: // <=
			// Jeśli p0 < 0 lub p0 == 0, przejdź do następnej instrukcji (kontynuuj)
			// W przeciwnym razie, skocz do początku pętli
			proceedAddr := 2
			c.emit(code.JNEG, int64(proceedAddr))  // Jeśli p0 < 0, przejdź dalej
			c.emit(code.JZERO, int64(proceedAddr)) // Jeśli p0 == 0, przejdź dalej
			c.emit(code.JUMP, int64(jumpBack))     // W przeciwnym razie, skocz do startAddr

		case token.LE: // <
			// Jeśli p0 < 0, przejdź do następnej instrukcji (kontynuuj)
			// W przeciwnym razie, skocz do początku pętli
			proceedAddr := 1
			c.emit(code.JNEG, int64(proceedAddr)) // Jeśli p0 < 0, przejdź dalej
			c.emit(code.JUMP, int64(jumpBack))    // W przeciwnym razie, skocz do startAddr

		case token.NEQUALS:
			proceedAddr := 2
			c.emit(code.JNEG, int64(proceedAddr))
			c.emit(code.JPOS, int64(proceedAddr))
			c.emit(code.JZERO, int64(jumpBack))
		default:
			return fmt.Errorf("nieobsługiwany operator warunkowy: %v", operator)
		}
		return nil
	}

	leftVal, err := strconv.Atoi(rc.Condition.Left.String())
	if err != nil {
		addr, err := c.getAddr(rc.Condition.Left.String())
		if err != nil {
			return fmt.Errorf("failed to get address of leftVal: %v", err)
		}
		c.emit(code.LOAD, int64(addr))
	} else {
		c.emit(code.SET, int64(leftVal))
	}

	rightVal, err := strconv.Atoi(rc.Condition.Right.String())
	if err != nil {
		addr, err := c.getAddr(rc.Condition.Right.String())
		if err != nil {
			return fmt.Errorf("failed to get address of rightVal: %v", err)
		}
		c.emit(code.SUB, int64(addr))
		jumpBack := -(len(c.instructions) - startAddr)
		conditionCompiler(jumpBack)
	} else {
		c.emit(code.STORE, TEMP)
		c.emit(code.SET, int64(rightVal))
		c.emit(code.SUB, TEMP)
		jumpBack := -(len(c.instructions) - startAddr)
		conditionCompiler(jumpBack)
	}

	return nil
}

func (c *Compiler) compileDeclaration(declaration *ast.Declaration) error {
	if !declaration.IsTable {
		c.addresses[declaration.Pidentifier.Value] = len(c.addresses)
	} else {
		from, _ := strconv.Atoi(declaration.From.Value)
		to, _ := strconv.Atoi(declaration.To.Value)
		for i := from; i <= to; i++ {
			c.addresses[fmt.Sprintf("%s[%d]", declaration.Pidentifier.Value, i)] = len(c.addresses)
		}
	}
	return nil
}

func (c *Compiler) compileMultiplication(expression ast.MathExpression) error {
	panic("unimplemented")
}

func (c *Compiler) compileDivision(expression ast.MathExpression) error {
	panic("unimplemented")
}

func (c *Compiler) compileModulo(expression ast.MathExpression) error {
	panic("unimplemented")
}

func (c *Compiler) compileAddition(node ast.MathExpression) error {
	leftVal, err := strconv.Atoi(node.Left.String())
	if err != nil {
		addr, err := c.getAddr(node.Left.String())
		if err != nil {
			return fmt.Errorf("failed to get address of leftVal: %v", err)
		}
		c.emit(code.LOAD, int64(addr))
	} else {
		c.emit(code.SET, int64(leftVal))
	}

	rightVal, err := strconv.Atoi(node.Right.String())
	if err != nil {
		addr, err := c.getAddr(node.Right.String())
		if err != nil {
			return fmt.Errorf("failed to get address of rightVal: %v", err)
		}
		c.emit(code.ADD, int64(addr))
	} else {
		c.emit(code.STORE, TEMP)
		c.emit(code.SET, int64(rightVal))
		c.emit(code.ADD, TEMP)
	}
	return nil
}

func (c *Compiler) compileSubtraction(node ast.MathExpression) error {
	rightVal, err := strconv.Atoi(node.Left.String())
	if err != nil {
		addr, err := c.getAddr(node.Left.String())
		if err != nil {
			return fmt.Errorf("failed to get address of leftVal: %v", err)
		}
		c.emit(code.LOAD, int64(addr))
	} else {
		c.emit(code.SET, int64(rightVal))
	}

	leftVal, err := strconv.Atoi(node.Right.String())
	if err != nil {
		addr, err := c.getAddr(node.Right.String())
		if err != nil {
			return fmt.Errorf("failed to get address of rightVal: %v", err)
		}
		c.emit(code.SUB, int64(addr))
	} else {
		c.emit(code.STORE, TEMP)
		c.emit(code.SET, int64(leftVal))
		c.emit(code.SUB, TEMP)
	}
	return nil
}

func (c *Compiler) getAddr(identifier string) (int, error) {
	addr, ok := c.addresses[identifier]
	if !ok {
		return 0, fmt.Errorf("use of undeclared identifier: %s", identifier)
	}
	return addr, nil
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

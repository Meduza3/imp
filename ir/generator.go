package ir

import (
	"fmt"
	"io"

	"github.com/Meduza3/imp/ast"
	"github.com/Meduza3/imp/symboltable"
	"github.com/Meduza3/imp/token"
)

type CodeGenerator struct {
	symbolTable      symboltable.SymbolTable
	nextMemoryOffset int
	program          ast.Program
	instructions     []IRInstruction
	tempCount        int
	labelCount       int
}

func (cg *CodeGenerator) walkAST(w io.Writer) {
	for _, procedure := range cg.program.Procedures {
		cg.walkProcedure(procedure, w)
	}
}

func (cg *CodeGenerator) walkCommand(command ast.Command, w io.Writer) {
	switch node := command.(type) {
	case *ast.AssignCommand:
		// First check if identifier is of type pid, pid[num] or pid[pid]
		identifier := node.Identifier
		entry := cg.symbolTable[identifier.Value]
		switch entry.Type {
		case "ARRAY":
		case "IDENTIFIER":

		}
	}
}

func (cg *CodeGenerator) walkProcedure(procedure ast.Procedure, w io.Writer) {
	// procHead := procedure.ProcHead
}

// NewCodeGenerator returns a fresh code generator
func NewCodeGenerator(program ast.Program, symbolTable symboltable.SymbolTable) *CodeGenerator {
	return &CodeGenerator{
		program:     program,
		symbolTable: symbolTable,
	}
}

// GetInstructions returns the generated IR
func (cg *CodeGenerator) GetInstructions() []IRInstruction {
	return cg.instructions
}

// emit is a helper to append a new IR instruction
func (cg *CodeGenerator) emit(instr IRInstruction) {
	cg.instructions = append(cg.instructions, instr)
}

// newTemp returns a new temporary variable name, e.g. "t1", "t2", ...
func (cg *CodeGenerator) newTemp() string {
	cg.tempCount++
	return fmt.Sprintf("t%d", cg.tempCount)
}

// newLabel returns a new label name, e.g. "L1", "L2", ...
func (cg *CodeGenerator) newLabel() string {
	cg.labelCount++
	return fmt.Sprintf("L%d", cg.labelCount)
}

func (cg *CodeGenerator) GenerateProgram(prog ast.Program) {
	for _, proc := range prog.Procedures {
		cg.GenerateProcedure(&proc)
	}
	cg.GenerateMain(prog.Main)
}

func (cg *CodeGenerator) GenerateMain(main ast.Main) {
	for _, decl := range main.Declarations {
		cg.AddToSymbolTable(decl)
	}

	cg.emit(IRInstruction{Op: OpLabel, Label: "main"})
	// Generate IR for commands in the main procedure
	for _, command := range main.Commands {
		cg.GenerateCommand(command)
	}

	// Emit halt instruction at the end of the main procedure
	cg.emit(IRInstruction{
		Op: OpHalt,
	})
}

func (c *CodeGenerator) GenerateIfCommand(ifc ast.IfCommand) {
	cond := ifc.Condition
	left := c.GenerateValue(cond.Left)
	right := c.GenerateValue(cond.Right)
	var op Opcode
	switch cond.Operator.Type {
	case token.EQUALS:
		op = OpIfEQ
	case token.NEQUALS:
		op = OpIfNEQ
	case token.GR:
		op = OpIfGR
	case token.LE:
		op = OpIfLE
	case token.GEQ:
		op = OpIfGEQ
	case token.LEQ:
		op = OpIfLEQ
	default:
		// handle unknown operator
	}

	elseLabel := c.newLabel()
	endLabel := c.newLabel()

	c.emit(IRInstruction{
		Op:    invertCompare(op), // e.g. if we see OpIfEQ, then we do OpIfNE => goto else
		Arg1:  left,
		Arg2:  right,
		Label: elseLabel,
	})

	for _, comm := range ifc.ThenCommands {
		c.GenerateCommand(comm)
	}

	c.emit(IRInstruction{
		Op:    OpGoto,
		Label: endLabel,
	})

	c.emit(IRInstruction{
		Op:    OpLabel,
		Label: elseLabel,
	})

	for _, comm := range ifc.ElseCommands {
		c.GenerateCommand(comm)
	}

	c.emit(IRInstruction{
		Op:    OpLabel,
		Label: endLabel,
	})

}

func (c *CodeGenerator) GenerateCommand(cmd ast.Command) {
	switch comm := cmd.(type) {
	case *ast.IfCommand:
		c.GenerateIfCommand(*comm)
	case *ast.ProcCallCommand:
		c.GenerateProcCallCommand(*comm)
	case *ast.ForCommand:
		c.GenerateForCommand(comm)
	case *ast.WriteCommand:
		c.GenerateWriteCommand(comm)
	case *ast.ReadCommand:
		c.GenerateReadCommand(comm)
	case *ast.AssignCommand:
		c.GenerateAssignCommand(comm)
	case *ast.RepeatCommand:
		c.GenerateRepeatCommand(comm)
	case *ast.WhileCommand:
		c.GenerateWhileCommand(comm)
	}
}

func (cg *CodeGenerator) GenerateProcCallCommand(node ast.ProcCallCommand) {
	for _, arg := range node.Args {
		argName := arg.Value
		cg.emit(IRInstruction{
			Op:   OpParam,
			Arg1: argName,
		})
	}

	cg.emit(IRInstruction{
		Op:   OpCall,
		Arg1: node.Name.Value, // e.g. “foo”
	})
}

func (cg *CodeGenerator) GenerateForCommand(fc *ast.ForCommand) {
	loopVar := fc.Iterator.Value

	fromVal := cg.GenerateValue(fc.From)
	toVal := cg.GenerateValue(fc.To)

	cg.emit(IRInstruction{
		Op:   OpAssign,
		Arg1: fromVal,
		Dest: loopVar,
	})

	loopLabel := cg.newLabel()
	endLabel := cg.newLabel()

	cg.emit(IRInstruction{
		Op:    OpLabel,
		Label: loopLabel,
	})

	if !fc.IsDownTo {
		cg.emit(IRInstruction{
			Op:    OpIfGR,
			Arg1:  loopVar,
			Arg2:  toVal,
			Label: endLabel,
		})

		for _, cmd := range fc.Commands {
			cg.GenerateCommand(cmd)
		}

		cg.emit(IRInstruction{
			Op:   OpAdd,
			Arg1: loopVar,
			Arg2: "1",
			Dest: loopVar,
		})

		cg.emit(IRInstruction{
			Op:    OpGoto,
			Label: loopLabel,
		})
	} else {
		cg.emit(IRInstruction{
			Op:    OpIfLE,
			Arg1:  loopVar,
			Arg2:  toVal,
			Label: endLabel,
		})

		for _, cmd := range fc.Commands {
			cg.GenerateCommand(cmd)
		}

		cg.emit(IRInstruction{
			Op:   OpSub,
			Arg1: loopVar,
			Arg2: "1",
			Dest: loopVar,
		})

		cg.emit(IRInstruction{
			Op:    OpGoto,
			Label: loopLabel,
		})
	}
	cg.emit(IRInstruction{
		Op:    OpLabel,
		Label: endLabel,
	})
}

func (cg *CodeGenerator) GenerateReadCommand(rc *ast.ReadCommand) {
	// Suppose rc.Identifier.String() is the variable name, e.g. "x" or "tab[5]".
	// For a basic approach, we can just read directly into that name:
	varName := rc.Identifier.String()

	cg.emit(IRInstruction{
		Op:   OpRead,
		Arg1: varName,
	})
}

func (cg *CodeGenerator) GenerateWriteCommand(wc *ast.WriteCommand) {
	// Evaluate the expression => yields a string name (like "t1" or "42")
	val := cg.GenerateValue(wc.Value)

	cg.emit(IRInstruction{
		Op:   OpWrite,
		Arg1: val,
	})
}

func (cg *CodeGenerator) GenerateAssignCommand(ac *ast.AssignCommand) {
	// Evaluate the expression on the right side => yields a string
	// For a single value expression, Right can be nil, so handle that case carefully.
	var exprVal string
	if ac.MathExpression.Right == nil {
		// Single-value expression (like "x := 5;")
		exprVal = cg.GenerateValue(ac.MathExpression.Left)
	} else {
		// Operator-based expression: "x := left op right"
		// We can do the generation in cg.GenerateValue(...) if you handle MathExpression inside that
		exprVal = cg.GenerateMathExpression(&ac.MathExpression)
	}

	// The left-hand side is an Identifier, e.g. "x" or "arr[i]"
	dest := ac.Identifier.String()

	// Emit IR: dest = exprVal
	cg.emit(IRInstruction{
		Op:   OpAssign,
		Arg1: exprVal,
		Dest: dest,
	})
}

func (cg *CodeGenerator) GenerateProcedure(p *ast.Procedure) {
	procLabel := p.ProcHead.Name.Value
	cg.emit(IRInstruction{
		Op:    OpLabel,
		Label: procLabel,
	})

	for _, decl := range p.Declarations {
		cg.AddToSymbolTable(decl)
	}

	for _, comm := range p.Commands {
		cg.GenerateCommand(comm)
	}

	// For a procedure with no explicit return, you might do:
	cg.emit(IRInstruction{
		Op: OpRet,
	})
}

func (cg *CodeGenerator) GenerateMathExpression(node *ast.MathExpression) string {
	// Evaluate left operand => might be a literal, identifier, or nested expression.
	left := cg.GenerateValue(node.Left)

	// Evaluate right operand
	right := cg.GenerateValue(node.Right)

	// Allocate a new temporary variable to hold the result
	resultTemp := cg.newTemp()

	// Pick the IR opcode based on the operator
	var op Opcode
	switch node.Operator.Type {
	case token.PLUS:
		op = OpAdd
	case token.MINUS:
		op = OpSub
	case token.MULT:
		op = OpMul
	case token.DIVIDE:
		op = OpDiv
	case token.MODULO:
		op = OpMod
	default:
		panic(fmt.Sprintf("Unknown operator %s", node.Operator.Type))
	}

	// Emit the IR instruction: resultTemp = left <op> right
	cg.emit(IRInstruction{
		Op:   op,
		Arg1: left,
		Arg2: right,
		Dest: resultTemp,
	})

	// Return the name of the temporary that now holds the result
	return resultTemp
}

func (cg *CodeGenerator) GenerateValue(node ast.Value) string {
	return node.String()
}

func (cg *CodeGenerator) GenerateWhileCommand(wc *ast.WhileCommand) {
	// Create labels
	startLabel := cg.newLabel()
	endLabel := cg.newLabel()

	// 1) Emit startLabel:
	cg.emit(IRInstruction{
		Op:    OpLabel,
		Label: startLabel,
	})

	// 2) Generate condition code
	left := cg.GenerateValue(wc.Condition.Left)
	right := cg.GenerateValue(wc.Condition.Right)

	// Convert the AST operator to IR opcode
	var op Opcode
	switch wc.Condition.Operator.Type {
	case token.EQUALS:
		op = OpIfEQ
	case token.NEQUALS:
		op = OpIfNEQ
	case token.GR:
		op = OpIfGR
	case token.LE:
		op = OpIfLE
	case token.GEQ:
		op = OpIfGEQ
	case token.LEQ:
		op = OpIfLEQ
	default:
		panic(fmt.Sprintf("unknown condition operator: %s", wc.Condition.Operator.Type))
	}

	// We want "if condition == false => goto endLabel".
	// So invert the condition. E.g. if (i < 10) is op=OpIfLE, then invert => OpIfGEQ
	invOp := invertCompare(op)

	cg.emit(IRInstruction{
		Op:    invOp,
		Arg1:  left,
		Arg2:  right,
		Label: endLabel,
	})

	// 3) Generate the loop body
	for _, cmd := range wc.Commands {
		cg.GenerateCommand(cmd)
	}

	// 4) Go back to start
	cg.emit(IRInstruction{
		Op:    OpGoto,
		Label: startLabel,
	})

	// 5) endLabel:
	cg.emit(IRInstruction{
		Op:    OpLabel,
		Label: endLabel,
	})
}

func (cg *CodeGenerator) GenerateRepeatCommand(rc *ast.RepeatCommand) {
	// 1) Create a start label
	startLabel := cg.newLabel()

	cg.emit(IRInstruction{
		Op:    OpLabel,
		Label: startLabel,
	})

	// 2) Generate the body
	for _, cmd := range rc.Commands {
		cg.GenerateCommand(cmd)
	}

	// 3) Evaluate the condition
	left := cg.GenerateValue(rc.Condition.Left)
	right := cg.GenerateValue(rc.Condition.Right)

	// Convert operator to IR opcode
	var op Opcode
	switch rc.Condition.Operator.Type {
	case token.EQUALS:
		op = OpIfEQ
	case token.NEQUALS:
		op = OpIfNEQ
	case token.GR:
		op = OpIfGR
	case token.LE:
		op = OpIfLE
	case token.GEQ:
		op = OpIfGEQ
	case token.LEQ:
		op = OpIfLEQ
	default:
		panic(fmt.Sprintf("unknown condition operator: %s", rc.Condition.Operator.Type))
	}

	// "until (cond)" means "while cond == false do repeat".
	// So, if condition is false => jump to start.
	// That’s effectively "if NOT(cond) goto Lstart".
	invOp := invertCompare(op)

	cg.emit(IRInstruction{
		Op:    invOp,
		Arg1:  left,
		Arg2:  right,
		Label: startLabel,
	})
}

func invertCompare(op Opcode) Opcode {
	switch op {
	case OpIfEQ:
		return OpIfNEQ
	case OpIfNEQ:
		return OpIfEQ
	case OpIfGR:
		return OpIfLEQ
	case OpIfLEQ:
		return OpIfGR
	case OpIfGEQ:
		return OpIfLE
	case OpIfLE:
		return OpIfGEQ
	}
	return op
}

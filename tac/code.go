package tac

import (
	"fmt"
	"strings"

	"github.com/Meduza3/imp/symboltable"
)

type Op string

const (
	OpAssign Op = "="

	OpAdd Op = "+"
	OpSub Op = "-"
	OpMul Op = "*"
	OpDiv Op = "/"
	OpMod Op = "%"

	OpLoadIndirect Op = "LOADI" // New opcode for loading from address

	// New: unconditional jump
	OpGoto Op = "goto"

	// New: conditional branches
	OpIfEQ Op = "if=="
	OpIfNE Op = "if!="
	OpIfLT Op = "if<"
	OpIfLE Op = "if<="
	OpIfGT Op = "if>"
	OpIfGE Op = "if>="

	OpRead  Op = "read"
	OpWrite Op = "write"

	OpParam Op = "param"
	OpCall  Op = "call"

	OpRet       Op = "ret"
	OpArrayLoad Op = "arrayLoad"
	OpHalt      Op = "halt"
)

type Instruction struct {
	Op          Op
	JumpTo      string
	Destination *symboltable.Symbol
	Arg1        *symboltable.Symbol
	Arg1Index   string
	Arg2        *symboltable.Symbol
	Arg2Index   string
	Labels      []string
}

func (ins Instruction) String() string {
	var parts []string
	if len(ins.Labels) != 0 {
		for _, label := range ins.Labels {
			parts = append(parts, fmt.Sprintf("%s: ", label))
		}
	}
	switch ins.Op {
	case OpAssign:
		// For a direct assignment, we only need Destination = Arg1
		// Example: "x = t1"
		if ins.Arg1Index != "" {
			parts = append(parts, fmt.Sprintf("%s[%s] = ", ins.Arg1.Name, ins.Arg1Index))
		} else {
			parts = append(parts, fmt.Sprintf("%s = ", ins.Arg1.Name))
		}
		if ins.Arg2Index != "" {
			parts = append(parts, fmt.Sprintf("%s[%s]", ins.Arg2.Name, ins.Arg2Index))
		} else {
			parts = append(parts, fmt.Sprintf("%s", ins.Arg2.Name))
		}

	case OpAdd, OpSub, OpMul, OpDiv, OpMod:
		// For arithmetic, we use three-address style: Destination = Arg1 op Arg2
		// Example: "t1 = x + y"
		parts = append(parts, fmt.Sprintf("%s = %s %s %s", ins.Destination.Name, ins.Arg1.Name, ins.Op, ins.Arg2.Name))

	case OpGoto:
		parts = append(parts, fmt.Sprintf("%s %s", ins.Op, ins.JumpTo))

	// conditional jumps
	case OpIfEQ, OpIfNE, OpIfLT, OpIfLE, OpIfGT, OpIfGE:
		parts = append(parts, fmt.Sprintf("%s %s, %s goto %s", ins.Op, ins.Arg1.Name, ins.Arg2.Name, ins.JumpTo))

	case OpCall:
		parts = append(parts, fmt.Sprintf("%s %s", ins.Op, ins.Arg1.Name))
	case OpRead, OpWrite, OpParam:
		if ins.Arg1Index != "" {
			parts = append(parts, fmt.Sprintf("%s[%s] %s", ins.Op, ins.Arg1Index, ins.Arg1.Name))
		} else {
			parts = append(parts, fmt.Sprintf("%s %s", ins.Op, ins.Arg1.Name))
		}

	case OpHalt, OpRet:
		parts = append(parts, string(ins.Op))
	default:
		// Handle any unrecognized ops (or extend this switch to cover other cases)
		parts = append(parts, fmt.Sprintf("Unknown instruction (Op=%q, Dest=%v, Arg1=%v, Arg2=%v)", ins.Op, ins.Destination, ins.Arg1, ins.Arg2))
	}
	return strings.Join(parts, " ")
}

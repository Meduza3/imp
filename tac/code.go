package tac

import "fmt"

type Op string

const (
	OpAssign Op = "="

	OpAdd Op = "+"
	OpSub Op = "-"
	OpMul Op = "*"
	OpDiv Op = "/"
	OpMod Op = "%"

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
)

type Instruction struct {
	Op          Op
	Destination string
	Arg1        string
	Arg2        string
	Label       string
}

func (ins Instruction) String() string {
	if ins.Label != "" {
		return fmt.Sprintf("%s:", ins.Label)
	}

	switch ins.Op {
	case OpAssign:
		// For a direct assignment, we only need Destination = Arg1
		// Example: "x = t1"
		return fmt.Sprintf("%s = %s", ins.Destination, ins.Arg1)

	case OpAdd, OpSub, OpMul, OpDiv, OpMod:
		// For arithmetic, we use three-address style: Destination = Arg1 op Arg2
		// Example: "t1 = x + y"
		return fmt.Sprintf("%s = %s %s %s", ins.Destination, ins.Arg1, string(ins.Op), ins.Arg2)

	case OpGoto:
		return fmt.Sprintf("%s %s", ins.Op, ins.Destination)

	// conditional jumps
	case OpIfEQ, OpIfNE, OpIfLT, OpIfLE, OpIfGT, OpIfGE:
		return fmt.Sprintf("%s %s, %s goto %s", ins.Op, ins.Arg1, ins.Arg2, ins.Destination)

	case OpRead, OpWrite, OpParam, OpCall:
		return fmt.Sprintf("%s %s", ins.Op, ins.Arg1)
	default:
		// Handle any unrecognized ops (or extend this switch to cover other cases)
		return fmt.Sprintf("Unknown instruction (Op=%q, Dest=%s, Arg1=%s, Arg2=%s)", ins.Op, ins.Destination, ins.Arg1, ins.Arg2)
	}
}

package ir

import "fmt"

type Opcode int

const (
	OpLabel Opcode = iota
	OpGoto
	OpIfEQ
	OpIfNEQ
	OpIfGEQ
	OpIfGR
	OpIfLE
	OpIfLEQ
	OpParam
	OpCall
	OpRet
	OpHalt

	OpAssign
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod

	OpRead
	OpWrite
)

type IRInstruction struct {
	Op    Opcode
	Arg1  string
	Arg2  string
	Dest  string
	Label string
}

func (i IRInstruction) String() string {
	switch i.Op {
	case OpLabel:
		return fmt.Sprintf("%s:", i.Label)
	case OpGoto:
		return fmt.Sprintf("goto %s", i.Label)
	case OpIfEQ:
		return fmt.Sprintf("if %s == %s goto %s", i.Arg1, i.Arg2, i.Label)
	case OpIfNEQ:
		return fmt.Sprintf("if %s != %s goto %s", i.Arg1, i.Arg2, i.Label)
	case OpIfGEQ:
		return fmt.Sprintf("if %s >= %s goto %s", i.Arg1, i.Arg2, i.Label)
	case OpIfGR:
		return fmt.Sprintf("if %s > %s goto %s", i.Arg1, i.Arg2, i.Label)
	case OpIfLE:
		return fmt.Sprintf("if %s < %s goto %s", i.Arg1, i.Arg2, i.Label)
	case OpIfLEQ:
		return fmt.Sprintf("if %s <= %s goto %s", i.Arg1, i.Arg2, i.Label)
	case OpParam:
		return fmt.Sprintf("param %s", i.Arg1)
	case OpCall:
		return fmt.Sprintf("call %s", i.Arg1)
	case OpRet:
		return "ret"
	case OpHalt:
		return "halt"

	case OpAssign:
		// Arg1 is the source, Dest is the variable
		return fmt.Sprintf("%s = %s", i.Dest, i.Arg1)
	case OpAdd:
		// Arg1 + Arg2 => Dest
		return fmt.Sprintf("%s = %s + %s", i.Dest, i.Arg1, i.Arg2)
	case OpSub:
		// Arg1 - Arg2 => Dest
		return fmt.Sprintf("%s = %s - %s", i.Dest, i.Arg1, i.Arg2)
	case OpMul:
		// Arg1 + Arg2 => Dest
		return fmt.Sprintf("%s = %s * %s", i.Dest, i.Arg1, i.Arg2)
	case OpDiv:
		// Arg1 + Arg2 => Dest
		return fmt.Sprintf("%s = %s / %s", i.Dest, i.Arg1, i.Arg2)
	case OpMod:
		// Arg1 + Arg2 => Dest
		return fmt.Sprintf("%s = %s %s %s", i.Dest, i.Arg1, "%", i.Arg2)
	case OpRead:
		// Arg1 is the variable name
		return fmt.Sprintf("read %s", i.Arg1)
	case OpWrite:
		// Arg1 is the expression (or a temp with the value)
		return fmt.Sprintf("write %s", i.Arg1)

	}
	return "UNKNOWN"
}

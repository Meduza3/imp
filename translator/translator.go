package translator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Meduza3/imp/code"
	"github.com/Meduza3/imp/symboltable"
	"github.com/Meduza3/imp/tac"
)

type Translator struct {
	Output              []code.Instruction
	labelCounter        int
	St                  symboltable.SymbolTable
	tempCounter         int
	currentAddress      int
	currentFunctionName string
	procEntries         map[string]int // Adresy początków procedur
	labels              map[string]int // Etykiety na adresy
	errors              []string
}

func (t *Translator) Errors() []string {
	return t.errors
}

func New(st symboltable.SymbolTable) *Translator {
	return &Translator{St: st, procEntries: make(map[string]int), labels: make(map[string]int)}
}

func (t *Translator) Translate(tac []tac.Instruction) []code.Instruction {

	t.firstPass(tac)

	return t.Output
}

func (t *Translator) firstPass(inss []tac.Instruction) {
	for _, ins := range inss {
		// If this instruction has a label, you might record the final “machine code”
		// address in a separate map, or emit a comment for readability:
		if ins.Label != "" {
			// You can store t.labels[ins.Label] = t.currentAddress
			// or just emit a comment or no-op, e.g.:
			instruction := code.Instruction{
				Op:      code.COMM,
				Comment: ins.Label,
			}
			t.emit(instruction)
		}

		switch ins.Op {
		//----------------------------------------------------------------------
		// 1) Simple Assignments:  a = b
		//----------------------------------------------------------------------
		case tac.OpAssign:
			dest := ins.Destination // e.g. "a"
			src := ins.Arg1         // e.g. "b" or maybe "5"

			t.handleAssign(dest, src)

		//----------------------------------------------------------------------
		// 2) Arithmetic: a = b + c, a = b - c, etc.
		//----------------------------------------------------------------------
		case tac.OpAdd, tac.OpSub:
			// For example:  ins = { Op: OpAdd, Destination: "x", Arg1: "a", Arg2: "b" }
			t.handleAddSub(ins.Op, ins.Destination, ins.Arg1, ins.Arg2)

		case tac.OpGoto:
			// "goto L"
			t.handleGoto(ins.Destination)

		case tac.OpIfLT, tac.OpIfLE, tac.OpIfGT, tac.OpIfGE, tac.OpIfEQ, tac.OpIfNE:
			// e.g. "if< x, y goto L"
			t.handleIf(ins.Op, ins.Arg1, ins.Arg2, ins.Destination)
		case tac.OpRead:
			err := t.handleRead(ins.Arg1)
			if err != nil {
				t.errors = append(t.errors, fmt.Sprintf("Failed to translate %v: %v", ins, err))
			}
		case tac.OpWrite:
			err := t.handleWrite(ins.Arg1)
			if err != nil {
				t.errors = append(t.errors, fmt.Sprintf("Failed to translate %v: %v", ins, err))
			}
		case tac.OpHalt:
			t.handleHalt()
			//----------------------------------------------------------------------
			// 3) TODO: Multiplication, Division, Modulo expansions
			//----------------------------------------------------------------------
			// We will fill in the rest (OpIfXX, OpGoto, OpCall, etc.) later
		default:
			t.errors = append(t.errors, fmt.Sprintf("Failed to translate %v", ins))
		}

	}
}

func (t *Translator) handleRead(arg1 string) error {
	addr, err := t.getAddr(arg1)
	if err != nil {
		return fmt.Errorf("Failed to get addr of %q", arg1)
	}
	// emit "GET addr"
	t.emit(code.Instruction{Op: code.GET, HasOperand: true, Operand: addr})
	return nil
}

func (t *Translator) handleWrite(arg1 string) error {
	addr, err := t.getAddr(arg1)
	if err != nil {
		return fmt.Errorf("Failed to get addr of %q", arg1)
	}
	// emit "GET addr"
	t.emit(code.Instruction{Op: code.PUT, HasOperand: true, Operand: addr})
	return nil
}

func (t *Translator) handleAssign(dest, src string) error {

	destAddr, err := t.getAddr(dest)
	if err != nil {
		return fmt.Errorf("failed to get destination address: %v", err)
	}

	srcAddr, err := t.getAddr(src)
	if err != nil {
		return fmt.Errorf("failed to get source address: %v", err)
	}

	t.emit(code.Instruction{Op: code.LOAD, HasOperand: true, Operand: srcAddr})
	t.emit(code.Instruction{Op: code.STORE, HasOperand: true, Operand: destAddr})
	return nil
}

func (t *Translator) handleAddSub(op tac.Op, dest, left, right string) error {
	// Example:  x = a + b  =>  LOAD a;  ADD b;  STORE x
	// or        x = a - b  =>  LOAD a;  SUB b;  STORE x
	destAddr, err := t.getAddr(dest)
	if err != nil {
		return fmt.Errorf("failed to get destAddres: %v", err)
	}
	// 1. Load left into p0.
	// normal variable
	leftAddr, err := t.getAddr(left)
	if err != nil {
		return fmt.Errorf(("failed to get left address: %v"), err)
	}

	rightAddr, err := t.getAddr(right)
	if err != nil {
		return fmt.Errorf(("failed to get right address: %v"), err)
	}

	t.emit(code.Instruction{Op: code.LOAD, HasOperand: true, Operand: leftAddr})

	// 2. Add/sub the right side
	// normal variable
	if op == tac.OpAdd {
		t.emit(code.Instruction{Op: code.ADD, HasOperand: true, Operand: rightAddr})
	} else {
		t.emit(code.Instruction{Op: code.SUB, HasOperand: true, Operand: rightAddr})
	}

	// 3. Store p0 into dest
	t.emit(code.Instruction{Op: code.STORE, HasOperand: true, Operand: destAddr})
	return nil
}

func (t *Translator) handleMod(dest, left, right string) error {
	panic("unimplemented")
}

func (t *Translator) handleDiv(dest, left, right string) error {
	panic("unimplemented")
}

func (t *Translator) handleMul(dest, left, right string) error {
	panic("unimplemented")
}

func (t *Translator) handleParam() error {

	panic("unimplemented")
}

func (t *Translator) handleCall() error {

	panic("unimplemented")
}

func (t *Translator) handleRet() error {
	panic("unimplemented")
}

func (t *Translator) handleHalt() {
	t.emit(code.Instruction{Op: code.HALT})
}

func (t *Translator) handleGoto(labelName string) {
	t.emit(code.Instruction{Op: code.JUMP, HasOperand: true, Destination: labelName})
}
func (t *Translator) handleIf(op tac.Op, left, right, labelName string) error {
	//   LOAD left
	//   SUB right
	//   then JNEG label, JPOS label, JZERO label, etc. depending on op

	// 1) LOAD p0 = left
	leftAddr, err := t.getAddr(left)
	if err != nil {
		return fmt.Errorf("failed to get left addr: %v", err)
	}

	rightAddr, err := t.getAddr(right)
	if err != nil {
		return fmt.Errorf("failed to get right addr: %v", err)
	}
	// 3) Jump if condition is satisfied
	var jumpIns *code.Instruction
	var jumpIns2 *code.Instruction
	switch op {
	case tac.OpIfLT:
		jumpIns = &code.Instruction{Op: code.JNEG, HasOperand: true, Destination: labelName}
	case tac.OpIfLE:
		jumpIns = &code.Instruction{Op: code.JNEG, HasOperand: true, Destination: labelName}
		jumpIns2 = &code.Instruction{Op: code.JZERO, HasOperand: true, Destination: labelName}
	case tac.OpIfGT:
		jumpIns = &code.Instruction{Op: code.JPOS, HasOperand: true, Destination: labelName}
	case tac.OpIfGE:
		jumpIns = &code.Instruction{Op: code.JPOS, HasOperand: true, Destination: labelName}
		jumpIns2 = &code.Instruction{Op: code.JZERO, HasOperand: true, Destination: labelName}
	case tac.OpIfEQ:
		jumpIns = &code.Instruction{Op: code.JZERO, HasOperand: true, Destination: labelName}
	case tac.OpIfNE:
		jumpIns = &code.Instruction{Op: code.JPOS, HasOperand: true, Destination: labelName}
		jumpIns2 = &code.Instruction{Op: code.JZERO, HasOperand: true, Destination: labelName}
	}
	t.emit(code.Instruction{Op: code.LOAD, HasOperand: true, Operand: leftAddr})
	t.emit(code.Instruction{Op: code.SUB, HasOperand: true, Operand: rightAddr})
	if jumpIns != nil {
		t.emit(*jumpIns)
	}
	if jumpIns2 != nil {
		t.emit(*jumpIns2)
	}

	return nil
}

func (t *Translator) newLocalLabel() string {
	t.labelCounter++
	return fmt.Sprintf("LOCAL_%d", t.labelCounter)
}

func (t *Translator) emit(code code.Instruction) {
	t.Output = append(t.Output, code)
	t.currentAddress++
}

func (t *Translator) getAddr(name string) (int, error) {
	_, err := strconv.Atoi(name)
	if err == nil {
		symbol, err := t.St.Lookup(name, "main")
		if err != nil {
			return 0, fmt.Errorf("failed to find addr for %q: %v", name, err)
		}
		return symbol.Address, nil
	}
	if strings.Contains(name, ".") {
		parts := strings.SplitN(name, ".", 2)
		proc := parts[0]
		sym := parts[1]
		symbol, err := t.St.Lookup(sym, proc)
		if err != nil {
			return 0, fmt.Errorf("failed to find addr for %q: %v", name, err)
		}
		return symbol.Address, nil
	}

	symbol, err := t.St.Lookup(name, "main")
	if err != nil {
		return 0, fmt.Errorf("failed to find addr for %q: %v", name, err)
	}
	return symbol.Address, nil
}

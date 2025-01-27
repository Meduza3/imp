package translator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Meduza3/imp/symboltable"
	"github.com/Meduza3/imp/tac"
)

type Translator struct {
	Output         []string
	labelCounter   int
	St             symboltable.SymbolTable
	tempCounter    int
	currentAddress int
	procEntries    map[string]int // Adresy początków procedur
	labels         map[string]int // Etykiety na adresy
	errors         []string
}

func New(st symboltable.SymbolTable) *Translator {
	return &Translator{St: st, procEntries: make(map[string]int), labels: make(map[string]int)}
}

func (t *Translator) Translate(tac []tac.Instruction) []string {

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
			t.emit(fmt.Sprintf("# %s:", ins.Label))
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
			// read var
			addr, err := t.getAddr(ins.Arg1)
			if err != nil {
				t.errors = append(t.errors, fmt.Sprintf("failed to read OpRead ins: %v", err))
			}
			// emit "GET addr"
			t.emit(fmt.Sprintf("GET %d", addr))

		case tac.OpWrite:
			// write var
			addr, err := t.getAddr(ins.Arg1)
			if err != nil {
				t.errors = append(t.errors, fmt.Sprintf("failed to read OpWrite ins: %v", err))
			}
			t.emit(fmt.Sprintf("PUT %d", addr))
		//----------------------------------------------------------------------
		// 3) TODO: Multiplication, Division, Modulo expansions
		//----------------------------------------------------------------------
		default:
			// We will fill in the rest (OpIfXX, OpGoto, OpCall, etc.) later
		}
	}
}

func isNumber(num string) bool {
	_, err := strconv.Atoi(num)
	if err != nil {
		return false
	}
	return true
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

	t.emit(fmt.Sprintf("LOAD %d", srcAddr)) // p0 = p_src

	// Now store p0 into p_dest
	t.emit(fmt.Sprintf("STORE %d", destAddr)) // p_dest = p0
	return nil
}

func (t *Translator) handleGoto(labelName string) {
	// We don’t know the *relative* offset yet, so emit something symbolic:
	// e.g. "JUMP labelName"
	t.emit(fmt.Sprintf("JUMP %s", labelName))
}

func (t *Translator) handleIf(op tac.Op, left, right, labelName string) error {
	// Typically we do:
	//   LOAD left
	//   SUB right
	//   then JNEG label, JPOS label, JZERO label, etc. depending on op

	// 1) LOAD p0 = left
	if isNumber(left) {
		leftTemp := t.ensureConst(left)
		t.emit(fmt.Sprintf("LOAD %d", leftTemp))
	} else {
		leftAddr, err := t.getAddr((left))
		if err != nil {
			return fmt.Errorf("failed to get left addr: %v", err)
		}
		t.emit(fmt.Sprintf("LOAD %d", leftAddr))
	}

	// 2) SUB p0 -= right
	if isNumber(right) {
		rightTemp := t.ensureConst(right)
		t.emit(fmt.Sprintf("SUB %d", rightTemp))
	} else {
		rightAddr, err := t.getAddr((left))
		if err != nil {
			return fmt.Errorf("failed to get right addr: %v", err)
		}
		t.emit(fmt.Sprintf("SUB %d", rightAddr))
	}

	// 3) Jump if condition is satisfied
	switch op {
	case tac.OpIfLT:
		// if x < y => x-y < 0 => JNEG
		t.emit(fmt.Sprintf("JNEG %s", labelName))
	case tac.OpIfLE:
		// x <= y => x-y <= 0 => x-y < 0 or x-y == 0 => we can do
		//   JNEG labelName
		//   JZERO labelName
		// or handle with single approach.
		// Simpler: JNEG labelName; JZERO labelName
		// but that’s 2 instructions. If your VM has only JNEG, JZERO, JPOS,
		// you do exactly that:
		t.emit(fmt.Sprintf("JNEG %s", labelName))
		t.emit(fmt.Sprintf("JZERO %s", labelName))

	case tac.OpIfGT:
		// x > y => x-y > 0 => JPOS label
		t.emit(fmt.Sprintf("JPOS %s", labelName))

	case tac.OpIfGE:
		// x >= y => x-y >= 0 => x-y > 0 or x-y==0 => JPOS label; JZERO label
		t.emit(fmt.Sprintf("JPOS %s", labelName))
		t.emit(fmt.Sprintf("JZERO %s", labelName))

	case tac.OpIfEQ:
		// x == y => x-y == 0 => JZERO label
		t.emit(fmt.Sprintf("JZERO %s", labelName))

	case tac.OpIfNE:
		// x != y => x-y != 0 => NOT JZERO => we can do a
		//   JZERO skip
		//   JUMP label
		// skip:
		// but that means we also need a local label or do a single pass approach.
		// For purely symbolic, let's do:
		//   "JZERO <someTempLabel>"
		//   "JUMP labelName"
		//   "# <someTempLabel>:"
		skipLabel := t.newLocalLabel()
		t.emit(fmt.Sprintf("JZERO %s", skipLabel))
		// if not zero => jump to labelName
		t.emit(fmt.Sprintf("JUMP %s", labelName))
		// place skipLabel:
		t.emit(fmt.Sprintf("# %s:", skipLabel))
	}
	return nil
}

func (t *Translator) newLocalLabel() string {
	t.labelCounter++
	return fmt.Sprintf("LOCAL_%d", t.labelCounter)
}

func (v *Translator) emit(code string) {
	v.Output = append(v.Output, code)
	v.currentAddress++
}

func (t *Translator) handleAddSub(op tac.Op, dest, left, right string) error {
	// Example:  x = a + b  =>  LOAD a;  ADD b;  STORE x
	// or        x = a - b  =>  LOAD a;  SUB b;  STORE x
	destAddr, err := t.getAddr(dest)
	if err != nil {
		return fmt.Errorf("failed to get destAddres: %v", err)
	}
	// 1. Load left into p0.
	if isNumber(left) {
		// If left is literal, we must put it in p0 via SET, then store in a temp cell, then re-LOAD that cell
		leftTemp := t.ensureConst(left)
		t.emit(fmt.Sprintf("LOAD %d", leftTemp))
	} else {
		// normal variable
		leftAddr, err := t.getAddr(left)
		if err != nil {
			return fmt.Errorf(("failed to get left address: %v"), err)
		}
		t.emit(fmt.Sprintf("LOAD %d", leftAddr))
	}

	// 2. Add/sub the right side
	if isNumber(right) {
		rightTemp := t.ensureConst(right)
		if op == tac.OpAdd {
			t.emit(fmt.Sprintf("ADD %d", rightTemp))
		} else {
			t.emit(fmt.Sprintf("SUB %d", rightTemp))
		}
	} else {
		// normal variable
		rightAddr, err := t.getAddr(right)
		if err != nil {
			if op == tac.OpAdd {
				t.emit(fmt.Sprintf("ADD %d", rightAddr))
			} else {
				t.emit(fmt.Sprintf("SUB %d", rightAddr))
			}
		}
	}

	// 3. Store p0 into dest
	t.emit(fmt.Sprintf("STORE %d", destAddr))
	return nil
}

func (t *Translator) ensureConst(num string) int {
	// Option A: Reuse if we already have an address for this literal
	/*
		if addr, ok := t.constPool[num]; ok {
				return addr
		}
	*/
	// Option B: Always create a new address
	t.tempCounter++
	newAddr := 1000 + t.tempCounter
	// Steps to put `num` into that address:
	//
	// 1) p0 = num
	t.emit(fmt.Sprintf("SET %s", num))
	// 2) thatAddress = p0
	t.emit(fmt.Sprintf("STORE %d", newAddr))

	// If you want to cache it:
	// t.constPool[num] = newAddr
	return newAddr
}

func (t *Translator) getAddr(name string) (int, error) {
	_, err := strconv.Atoi(name)
	if err != nil {
		symbol, err := t.St.Lookup(name, "main")
		if err != nil {
			return 0, fmt.Errorf("failed to find addr for %q: %v", name, err)
		}
		return symbol.Address, nil
	} else {
		parts := strings.Split(name, ".")
		symbol, err := t.St.Lookup(parts[1], parts[0])
		if err != nil {
			return 0, fmt.Errorf("failed to find addr for %q: %v", name, err)
		}
		return symbol.Address, nil
	}
}

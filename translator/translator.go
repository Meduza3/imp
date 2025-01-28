package translator

import (
	"fmt"
	"sort"
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
	paramCount          int
}

func (t *Translator) Errors() []string {
	return t.errors
}

func New(st symboltable.SymbolTable) *Translator {
	return &Translator{St: st, procEntries: make(map[string]int), labels: make(map[string]int)}
}

func (t *Translator) Translate(tac []tac.Instruction) []code.Instruction {

	t.firstPass(tac)
	output := t.secondPass(t.Output)
	t.Output = output
	return t.Output
}

func (t *Translator) secondPass(input []code.Instruction) []code.Instruction {
	// 1) Collect all labels => line index
	labelAddress := make(map[string]int)
	for i := 0; i < len(input); i++ {
		if input[i].Label != "" {
			labelAddress[input[i].Label] = i
			input[i].Label = ""
		}
	}

	// 2) Patch up every instruction that has a Destination
	for i := 0; i < len(input); i++ {
		if input[i].Destination != "" {
			target := labelAddress[input[i].Destination]
			// If your machine uses JUMP j => (k = k + j), typically offset = target - i
			// or possibly: offset = target - (i + 1).  Choose whichever your VM expects:
			relAddr := target - i
			input[i].Destination = ""
			input[i].Operand = relAddr
			input[i].HasOperand = true
		}
	}
	return input
}
func (t *Translator) firstPass(inss []tac.Instruction) {
	for _, ins := range inss {
		// If this instruction has a label, you might record the final “machine code”
		// address in a separate map, or emit a comment for readability:
		var label string
		if ins.Label != "" {
			if !strings.HasPrefix(ins.Label, "L") {
				t.currentFunctionName = ins.Label // Update the current function name
			}
			// You can store t.labels[ins.Label] = t.currentAddress
			// or just emit a comment or no-op, e.g.:
			label = ins.Label
		}

		switch ins.Op {
		//----------------------------------------------------------------------
		// 1) Simple Assignments:  a = b
		//----------------------------------------------------------------------
		case tac.OpAssign:
			dest := ins.Destination // e.g. "a"
			src := ins.Arg1         // e.g. "b" or maybe "5"

			t.handleAssign(dest, src, label)

		//----------------------------------------------------------------------
		// 2) Arithmetic: a = b + c, a = b - c, etc.
		//----------------------------------------------------------------------
		case tac.OpAdd, tac.OpSub:
			// For example:  ins = { Op: OpAdd, Destination: "x", Arg1: "a", Arg2: "b" }
			t.handleAddSub(ins.Op, ins.Destination, ins.Arg1, ins.Arg2, label)
		case tac.OpMod:
			err := t.handleMod(ins.Destination, ins.Arg1, ins.Arg2, label)
			if err != nil {
				t.errors = append(t.errors, fmt.Sprintf("Failed to translate %v: %v", ins, err))
			}
		case tac.OpGoto:
			// "goto L"
			t.handleGoto(ins.Destination, label)

		case tac.OpIfLT, tac.OpIfLE, tac.OpIfGT, tac.OpIfGE, tac.OpIfEQ, tac.OpIfNE:
			// e.g. "if< x, y goto L"
			t.handleIf(ins.Op, ins.Arg1, ins.Arg2, ins.Destination, label)
		case tac.OpRead:
			err := t.handleRead(ins.Arg1, label)
			if err != nil {
				t.errors = append(t.errors, fmt.Sprintf("Failed to translate %v: %v", ins, err))
			}
		case tac.OpWrite:
			err := t.handleWrite(ins.Arg1, label)
			if err != nil {
				t.errors = append(t.errors, fmt.Sprintf("Failed to translate %v: %v", ins, err))
			}
		case tac.OpHalt:
			t.handleHalt(label)
			//----------------------------------------------------------------------
			// 3) TODO: Multiplication, Division, Modulo expansions
			//----------------------------------------------------------------------
			// We will fill in the rest (OpIfXX, OpGoto, OpCall, etc.) later
		case tac.OpCall:
			err := t.handleCall(ins.Arg1, ins.Arg2, label)
			if err != nil {
				t.errors = append(t.errors, fmt.Sprintf("Failed to translate %v: %v", ins, err))
			}
		case tac.OpParam:
			err := t.handleParam(ins.Arg1, label)
			if err != nil {
				t.errors = append(t.errors, fmt.Sprintf("Failed to translate %v: %v", ins, err))
			}
		case tac.OpRet:
			err := t.handleRet(label)
			if err != nil {
				t.errors = append(t.errors, fmt.Sprintf("Failed to translate %v: %v", ins, err))
			}
		default:
			t.errors = append(t.errors, fmt.Sprintf("Failed to translate %v", ins))
		}

	}
}

func (t *Translator) handleRead(arg1, label string) error {
	addr, err := t.getAddr(arg1)
	if err != nil {
		return fmt.Errorf("Failed to get addr of %q", arg1)
	}
	// emit "GET addr"
	ins := code.Instruction{Op: code.GET, HasOperand: true, Label: label, Operand: addr}
	t.emit(ins)
	return nil
}

func (t *Translator) handleWrite(arg1, label string) error {
	addr, err := t.getAddr(arg1)
	if err != nil {
		return fmt.Errorf("Failed to get addr of %q", arg1)
	}
	// emit "GET addr"
	ins := code.Instruction{Op: code.PUT, HasOperand: true, Label: label, Operand: addr}
	t.emit(ins)
	return nil
}

func (t *Translator) handleAssign(dest, src, label string) error {

	destAddr, err := t.getAddr(dest)
	if err != nil {
		return fmt.Errorf("failed to get destination address: %v", err)
	}

	srcAddr, err := t.getAddr(src)
	if err != nil {
		return fmt.Errorf("failed to get source address: %v", err)
	}

	t.emit(code.Instruction{Op: code.LOAD, Label: label, HasOperand: true, Operand: srcAddr})
	t.emit(code.Instruction{Op: code.STORE, HasOperand: true, Operand: destAddr})
	return nil
}

func (t *Translator) handleAddSub(op tac.Op, dest, left, right, label string) error {
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

	t.emit(code.Instruction{Op: code.LOAD, Label: label, HasOperand: true, Operand: leftAddr})

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

func (t *Translator) handleMod(dest, left, right, label string) error {
	leftAddr, err := t.getAddr(left)
	if err != nil {
		return fmt.Errorf("failed to get left (%s) address: %v", left, err)
	}
	rightAddr, err := t.getAddr(right)
	if err != nil {
		return fmt.Errorf("failed to get right (%s) address: %v", left, err)
	}
	destAddr, err := t.getAddr(dest)
	if err != nil {
		return fmt.Errorf("failed to get dest (%s) address: %v", dest, err)
	}
	t.emit(code.Instruction{Op: code.LOAD, HasOperand: true, Label: label, Operand: leftAddr})
	t.emit(code.Instruction{Op: code.SUB, HasOperand: true, Operand: rightAddr})
	t.emit(code.Instruction{Op: code.JNEG, HasOperand: true, Operand: 2})
	t.emit(code.Instruction{Op: code.JUMP, HasOperand: true, Operand: -2})
	t.emit(code.Instruction{Op: code.ADD, HasOperand: true, Operand: rightAddr})
	t.emit(code.Instruction{Op: code.STORE, HasOperand: true, Operand: destAddr})
	return nil
}

func (t *Translator) handleDiv(dest, left, right, label string) error {
	panic("unimplemented")
}

func (t *Translator) handleMul(dest, left, right, label string) error {
	panic("unimplemented")
}

func (t *Translator) handleParam(param, label string) error {
	paramAddr, err := t.getAddr(param)
	if err != nil {
		return fmt.Errorf("failed getting %s address: %v", param, err)
	}
	ins, _ := code.Make(code.SET, paramAddr)
	ins.Label = label
	t.emit(*ins)
	paramStore := t.St.CurrentOffset + t.paramCount
	t.paramCount++
	ins, _ = code.Make(code.STORE, paramStore)
	t.emit(*ins)
	return nil
}

func (t *Translator) handleCall(funcName, paramCount, label string) error {
	paramCountInt, err := strconv.Atoi(paramCount)
	if err != nil {
		return fmt.Errorf("failed to convert paramCount %s to int: %v", paramCount, err)
	}
	returnStore, err := t.St.Lookup(funcName+"_return", "xxFunctionsxx")
	if err != nil {
		return fmt.Errorf("failed to get return address of func %s: %v", funcName, err)
	}
	var arguments []symboltable.Symbol
	for _, symbol := range t.St.Table[funcName] {
		if symbol.Kind == symboltable.ARGUMENT {
			arguments = append(arguments, symbol)
		}
	}
	sort.Slice(arguments, func(i, j int) bool {
		return arguments[i].ArgumentIndex < arguments[j].ArgumentIndex
	})

	for i := paramCountInt; i > 0; i-- {
		ins, _ := code.Make(code.LOAD, arguments[paramCountInt-i].Address)
		ins.Label = label
		ins2, _ := code.Make(code.STORE, arguments[paramCountInt-i].Address)
		t.emit(*ins)
		t.emit(*ins2)
	}
	returnAddr := len(t.Output) + 3
	ins3, _ := code.Make(code.SET, returnAddr)
	ins4, _ := code.Make(code.STORE, returnStore.Address)
	t.emit(*ins3)
	t.emit(*ins4)

	ins5 := code.Instruction{Op: code.JUMP, Destination: funcName}
	t.emit(ins5)
	return nil
}

func (t *Translator) handleRet(label string) error {
	returnAddr, err := t.St.Lookup(t.currentFunctionName+"_return", "xxFunctionsxx")
	if err != nil {
		return fmt.Errorf("failed to get return for function %s: %v", t.currentFunctionName, err)
	}
	t.emit(code.Instruction{Op: code.RTRN, Label: label, HasOperand: true, Operand: returnAddr.Address})
	return nil
}

func (t *Translator) handleHalt(label string) {
	t.emit(code.Instruction{Label: label, Op: code.HALT})
}

func (t *Translator) handleGoto(labelName, label string) {
	t.emit(code.Instruction{Op: code.JUMP, Label: label, HasOperand: true, Destination: labelName})
}
func (t *Translator) handleIf(op tac.Op, left, right, labelName, label string) error {
	//   LOAD left
	//   SUB right
	//   then JNEG label, JPOSAlabel, JZERO label, etc. depending on op

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
	t.emit(code.Instruction{Op: code.LOAD, HasOperand: true, Label: label, Operand: leftAddr})
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

	symbol, err := t.St.Lookup(name, t.currentFunctionName)
	if err != nil {
		return 0, fmt.Errorf("failed to find addr for %q: %v", name, err)
	}
	return symbol.Address, nil
}

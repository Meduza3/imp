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
	pointerCell         int
	Output              []code.Instruction
	labelCounter        int
	St                  symboltable.SymbolTable
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
	return &Translator{pointerCell: 10, St: st, procEntries: make(map[string]int), labels: make(map[string]int)}
}

func (t *Translator) Translate(tac []tac.Instruction) []code.Instruction {
	for _, ins := range tac {
		fmt.Println("# ins: ", ins)
	}
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
			input[i].Operand = relAddr
			input[i].Destination = ""
			input[i].HasOperand = true
		}
	}
	return input
}

func (t *Translator) setupConstants() {
	// 1) First declare/store all nonnegative constants
	for _, table := range t.St.Table {
		for _, value := range table {
			val, _ := strconv.Atoi(value.Name)
			if value.Kind == symboltable.CONSTANT {
				t.emit(code.Instruction{
					Op:         code.SET,
					HasOperand: true,
					Operand:    val,
					Comment:    "declaring constant " + value.Name,
				})
				t.emit(code.Instruction{Op: code.STORE, HasOperand: true, Operand: value.Address})
			}
		}
	}
}

func (t *Translator) firstPass(inss []tac.Instruction) {
	t.setupConstants()

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
		case tac.OpAssign: // e.g. "b" or maybe "5"
			err := t.handleAssign(ins)
			if err != nil {
				t.errors = append(t.errors, fmt.Sprintf("failed to handle Assign %v: %v", ins, err))
			}

		//----------------------------------------------------------------------
		// 2) Arithmetic: a = b + c, a = b - c, etc.
		//----------------------------------------------------------------------
		case tac.OpAdd, tac.OpSub:
			// For example:  ins = { Op: OpAdd, Destination: "x", Arg1: "a", Arg2: "b" }
			err := t.handleAddSub(ins)
			if err != nil {
				t.errors = append(t.errors, fmt.Sprintf("failed to handleAddSub %v: %v", ins, err))
			}
		case tac.OpMod:
			err := t.handleMod(ins)
			if err != nil {
				t.errors = append(t.errors, fmt.Sprintf("Failed to translate %v: %v", ins, err))
			}
		case tac.OpGoto:
			// "goto L"
			destination := ins.JumpTo
			t.handleGoto(destination, label)

		case tac.OpIfLT, tac.OpIfLE, tac.OpIfGT, tac.OpIfGE, tac.OpIfEQ, tac.OpIfNE:
			// e.g. "if< x, y goto L"
			t.handleIf(ins)
		case tac.OpRead:
			err := t.handleRead(ins)
			if err != nil {
				t.errors = append(t.errors, fmt.Sprintf("Failed to translate %v: %v", ins, err))
			}
		case tac.OpWrite:
			err := t.handleWrite(ins)
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
			err := t.handleCall(ins)
			if err != nil {
				t.errors = append(t.errors, fmt.Sprintf("Failed to translate %v: %v", ins, err))
			}
		case tac.OpParam:
			err := t.handleParam(*ins.Arg1, label)
			if err != nil {
				t.errors = append(t.errors, fmt.Sprintf("Failed to translate %v: %v", ins, err))
			}
		case tac.OpRet:
			err := t.handleRet(label)
			if err != nil {
				t.errors = append(t.errors, fmt.Sprintf("Failed to translate %v: %v", ins, err))
			}
		case tac.OpArrayLoad:
			err := t.handleArrayLoad(ins)
			if err != nil {
				t.errors = append(t.errors, fmt.Sprintf("Failed to translate %v: %v", ins, err))
			}
		default:
			t.errors = append(t.errors, fmt.Sprintf("Failed to translate %v", ins))
		}

	}
}

func (t *Translator) handleRead(ins tac.Instruction) error {
	pointerCell := t.pointerCell
	if ins.Arg1 == nil {
		panic("nil arg1!!!")
	}
	if ins.Arg1.IsTable {
		idxSym, err := t.getSymbol(ins.Arg1Index)
		if err != nil {
			return err
		}
		t.emit(code.Instruction{
			Op:         code.SET,
			HasOperand: true,
			Operand:    ins.Arg1.Address,
			Label:      ins.Label,
			Comment:    ins.String(),
		})
		t.emit(code.Instruction{
			Op:         code.ADD,
			HasOperand: true,
			Operand:    idxSym.Address,
		})
		// 2) STORE pointerCell
		t.emit(code.Instruction{
			Op:         code.STORE,
			HasOperand: true,
			Operand:    pointerCell,
			Comment:    "pointerCell = address of x[n]",
		})

		// 3) GET scratchCell => memory[scratchCell] = input
		t.emit(code.Instruction{
			Op:         code.GET,
			HasOperand: true,
			Operand:    0,
			Comment:    "read input directly into memory[scratchCell]",
		})
		// 5) STOREI pointerCell => memory[memory[pointerCell]] = ACC
		t.emit(code.Instruction{
			Op:         code.STOREI,
			HasOperand: true,
			Operand:    pointerCell,
			Comment:    "x[n] = input",
		})

	} else {

		t.emit(code.Instruction{
			Op:         code.GET,
			HasOperand: true,
			Operand:    ins.Arg1.Address,
			Label:      ins.Label,
		})
	}
	return nil
}

func (t *Translator) handleWrite(ins tac.Instruction) error {
	if ins.Arg1 == nil {
		panic("WRITE instruction has nil Arg1")
	}
	if ins.Arg1.IsTable {
		idxSym, err := t.getSymbol(ins.Arg1Index)
		if err != nil {
			return fmt.Errorf("failed to getSymbol for index %q: %v", ins.Arg1Index, err)
		}
		t.emit(code.Instruction{
			Op:         code.SET,
			HasOperand: true,
			Operand:    ins.Arg1.Address,
			Label:      ins.Label,
			Comment:    ins.String(),
		})
		t.emit(code.Instruction{
			Op:         code.ADD,
			HasOperand: true,
			Operand:    idxSym.Address,
		})
		t.emit(code.Instruction{
			Op:         code.LOADI,
			HasOperand: true,
			Operand:    0,
			Comment:    "load x[n] into ACC",
		})
		t.emit(code.Instruction{
			Op:         code.PUT,
			HasOperand: true,
			Operand:    0,
			Comment:    "write x[n]",
		})
	} else {
		ins := code.Instruction{Op: code.PUT, HasOperand: true, Label: ins.Label, Operand: ins.Arg1.Address}
		t.emit(ins)
	}
	return nil
}

func (t *Translator) handleAssign(ins tac.Instruction) error {
	if ins.Arg1 == nil || ins.Arg2 == nil {
		return fmt.Errorf("nil argument in assignment instruction: %v", ins)
	}
	dest := ins.Arg1
	destIndex := ins.Arg1Index
	src := ins.Arg2
	srcIndex := ins.Arg2Index
	label := ins.Label
	if src.IsTable && dest.IsTable {
		// Case 4: Array Element to Array Element (x[n] := y[m])
		return t.handleArrayToArrayAssign(*dest, *src, destIndex, srcIndex, label)
	} else if src.IsTable {
		// Case 3: Array Element to Variable (a := x[n])
		return t.handleArrayToVarAssign(*dest, *src, srcIndex, label)
	} else if dest.IsTable {
		// Case 2: Variable to Array Element (x[n] := b)
		return t.handleVarToArrayAssign(*dest, *src, destIndex, label)
	} else {
		// Case 1: Variable to Variable (a := b)
		return t.handleVarToVarAssign(*dest, *src, label)
	}
}

func (t *Translator) handleVarToVarAssign(dest, src symboltable.Symbol, label string) error {
	fmt.Println("# called var to var with ", dest, src)
	// Load src into ACC
	t.emit(code.Instruction{
		Op:         code.LOAD,
		HasOperand: true,
		Operand:    src.Address,
		Comment:    fmt.Sprintf("load %v into ACC", src),
		Label:      label,
	})

	// Store ACC into dest
	t.emit(code.Instruction{
		Op:         code.STORE,
		HasOperand: true,
		Operand:    dest.Address,
		Comment:    fmt.Sprintf("store ACC into %v", dest),
	})

	return nil
}

func (t *Translator) handleVarToArrayAssign(dest, src symboltable.Symbol, destIndex string, label string) error {

	// Load the index value into ACC
	t.emit(code.Instruction{
		Op:         code.SET,
		Operand:    dest.Address,
		Comment:    fmt.Sprintf("ACC <== Address of %v[0]", dest.Name),
		Label:      label,
		HasOperand: true,
	})
	// Add the base address
	indexSymbol, err := t.getSymbol(destIndex)
	if err != nil {
		return err
	}
	t.emit(code.Instruction{
		Op:         code.ADD,
		Operand:    indexSymbol.Address,
		Comment:    fmt.Sprintf("add index at address to ACC"),
		HasOperand: true,
	})
	t.emit(code.Instruction{
		Op:         code.STORE,
		HasOperand: true,
		Operand:    t.pointerCell,
		Comment:    fmt.Sprintf("pc = &x[n]"),
	})
	t.emit(code.Instruction{
		Op:         code.LOAD,
		HasOperand: true,
		Operand:    src.Address,
		Comment:    fmt.Sprintf("load %v into ACC", src.Name),
	})
	t.emit(code.Instruction{
		Op:         code.STOREI,
		HasOperand: true,
		Operand:    t.pointerCell,
		Comment:    "p[pointerCell] <== ACC",
	})
	t.emit(code.Instruction{})

	return nil
}

func (t *Translator) handleArrayToVarAssign(dest, src symboltable.Symbol, srcIndex, label string) error {

	if err := t.loadOperandIndirect(src, srcIndex, label); err != nil {
		return err
	}
	t.emit(code.Instruction{
		Op:         code.STORE,
		HasOperand: true,
		Operand:    dest.Address,
		Comment:    "p[pointerCell] <== ACC",
	})
	return nil
}

func (t *Translator) handleArrayToArrayAssign(dest, src symboltable.Symbol, srcIndex, destIndex, label string) error {
	t.emit(code.Instruction{
		Op:         code.SET,
		Operand:    dest.Address,
		HasOperand: true,
		Label:      label,
	})
	indexSymbol, err := t.getSymbol(destIndex)
	if err != nil {
		return err
	}
	t.emit(code.Instruction{
		Op:         code.ADD,
		Operand:    indexSymbol.Address,
		HasOperand: true,
	})
	t.emit(code.Instruction{
		Op:         code.STORE,
		Operand:    t.pointerCell,
		HasOperand: true,
	})
	err = t.loadOperandIndirect(src, srcIndex, "")
	if err != nil {
		return err
	}
	t.emit(code.Instruction{
		Op:         code.STOREI,
		Operand:    t.pointerCell,
		HasOperand: true,
	})
	return nil
}

func (t *Translator) handleAddSub(ins tac.Instruction) error {
	dest := ins.Destination
	Arg1Index := ins.Arg1Index
	arg1 := ins.Arg1
	Arg2Index := ins.Arg2Index
	arg2 := ins.Arg2
	label := ins.Label
	if arg2.IsTable && dest.IsTable {
		return t.handleArrayAddSubArray(ins.Op, *dest, *arg1, *arg2, Arg1Index, Arg2Index, label)
	} else if arg2.IsTable {
		return t.handleVarAddSubArray(ins.Op, *dest, *arg1, *arg2, Arg2Index, label)
	} else if dest.IsTable {
		return t.handleArrayAddSubVar(ins.Op, *dest, *arg1, *arg2, Arg1Index, label)
	} else {
		return t.handleVarToVarAddSub(ins.Op, *dest, *arg1, *arg2, label)
	}
}

func (t *Translator) handleArrayAddSubArray(op tac.Op, dest symboltable.Symbol, arg1, arg2 symboltable.Symbol, Arg1Index, Arg2Index string, label string) error {
	t.emit(code.Instruction{
		Op:         code.SET,
		HasOperand: true,
		Operand:    arg1.Address,
		Label:      label,
	})
	idxSym, err := t.getSymbol(Arg2Index)
	if err != nil {
		return err
	}
	t.emit(code.Instruction{
		Op:         code.ADD,
		HasOperand: true,
		Operand:    idxSym.Address,
	})
	t.emit(code.Instruction{
		Op:         code.LOADI,
		HasOperand: true,
		Operand:    0,
	})

	t.emit(code.Instruction{
		Op:         code.SET,
		HasOperand: true,
		Operand:    arg2.Address,
	})
	idxSym, err = t.getSymbol(Arg2Index)
	if err != nil {
		return err
	}
	t.emit(code.Instruction{
		Op:         code.LOADI,
		HasOperand: true,
		Operand:    0,
	})
	if op == tac.OpAdd {

		t.emit(code.Instruction{
			Op:         code.ADD,
			HasOperand: true,
			Operand:    t.pointerCell,
		})
	} else {
		t.emit(code.Instruction{
			Op:         code.SUB,
			HasOperand: true,
			Operand:    t.pointerCell,
		})
	}
	t.emit(code.Instruction{
		Op:         code.STORE,
		HasOperand: true,
		Operand:    dest.Address,
	})
	return nil
}

func (t *Translator) handleVarAddSubArray(op tac.Op, dest symboltable.Symbol, arg1, arg2 symboltable.Symbol, Arg2Index string, label string) error {
	t.emit(code.Instruction{
		Op:         code.SET,
		HasOperand: true,
		Operand:    arg2.Address,
		Label:      label,
	})
	idxSym, err := t.getSymbol(Arg2Index)
	if err != nil {
		return err
	}
	t.emit(code.Instruction{
		Op:         code.ADD,
		HasOperand: true,
		Operand:    idxSym.Address,
	})

	t.emit(code.Instruction{
		Op:         code.LOADI,
		HasOperand: true,
		Operand:    0,
	})
	if op == tac.OpAdd {
		t.emit(code.Instruction{
			Op:         code.ADD,
			HasOperand: true,
			Operand:    arg1.Address,
		})
	} else {
		t.emit(code.Instruction{
			Op:         code.SUB,
			HasOperand: true,
			Operand:    arg1.Address,
		})
	}

	t.emit(code.Instruction{
		Op:         code.STORE,
		HasOperand: true,
		Operand:    dest.Address,
	})
	return nil
}

func (t *Translator) handleArrayAddSubVar(op tac.Op, dest symboltable.Symbol, arg1, arg2 symboltable.Symbol, Arg1Index string, label string) error {
	t.emit(code.Instruction{
		Op:         code.SET,
		HasOperand: true,
		Operand:    arg1.Address,
		Label:      label,
	})
	idxSym, err := t.getSymbol(Arg1Index)
	if err != nil {
		return err
	}
	t.emit(code.Instruction{
		Op:         code.ADD,
		HasOperand: true,
		Operand:    idxSym.Address,
	})

	if op == tac.OpAdd {
		t.emit(code.Instruction{
			Op:         code.ADD,
			HasOperand: true,
			Operand:    arg2.Address,
		})
	} else {
		t.emit(code.Instruction{
			Op:         code.SUB,
			HasOperand: true,
			Operand:    arg2.Address,
		})
	}
	t.emit(code.Instruction{
		Op:         code.STORE,
		HasOperand: true,
		Operand:    dest.Address,
	})
	return nil
}

func (t *Translator) handleVarToVarAddSub(op tac.Op, dest symboltable.Symbol, arg1 symboltable.Symbol, arg2 symboltable.Symbol, label string) error {
	t.emit(code.Instruction{
		Op:         code.LOAD,
		HasOperand: true,
		Operand:    arg1.Address,
		Label:      label,
	})
	if op == tac.OpAdd {
		t.emit(code.Instruction{
			Op:         code.ADD,
			HasOperand: true,
			Operand:    arg2.Address,
		})
	} else {
		t.emit(code.Instruction{
			Op:         code.SUB,
			HasOperand: true,
			Operand:    arg2.Address,
		})
	}
	t.emit(code.Instruction{
		Op:         code.STORE,
		HasOperand: true,
		Operand:    dest.Address,
	})
	return nil
}

func (t *Translator) ArrayMinusArrayLoad(arg1, arg2 symboltable.Symbol, Arg1Index, Arg2Index string, label string) error {
	t.emit(code.Instruction{
		Op:         code.SET,
		HasOperand: true,
		Operand:    arg1.Address,
		Label:      label,
	})
	idxSym, err := t.getSymbol(Arg2Index)
	if err != nil {
		return err
	}
	t.emit(code.Instruction{
		Op:         code.ADD,
		HasOperand: true,
		Operand:    idxSym.Address,
	})
	t.emit(code.Instruction{
		Op:         code.LOADI,
		HasOperand: true,
		Operand:    0,
	})

	t.emit(code.Instruction{
		Op:         code.SET,
		HasOperand: true,
		Operand:    arg2.Address,
	})
	idxSym, err = t.getSymbol(Arg2Index)
	if err != nil {
		return err
	}
	t.emit(code.Instruction{
		Op:         code.LOADI,
		HasOperand: true,
		Operand:    0,
	})
	t.emit(code.Instruction{
		Op:         code.SUB,
		HasOperand: true,
		Operand:    t.pointerCell,
	})

	return nil
}

func (t *Translator) VarMinusArrayLoad(arg1, arg2 symboltable.Symbol, Arg2Index string, label string) error {
	t.emit(code.Instruction{
		Op:         code.SET,
		HasOperand: true,
		Operand:    arg2.Address,
		Label:      label,
	})
	idxSym, err := t.getSymbol(Arg2Index)
	if err != nil {
		return err
	}
	t.emit(code.Instruction{
		Op:         code.ADD,
		HasOperand: true,
		Operand:    idxSym.Address,
	})

	t.emit(code.Instruction{
		Op:         code.LOADI,
		HasOperand: true,
		Operand:    0,
	})

	t.emit(code.Instruction{
		Op:         code.SUB,
		HasOperand: true,
		Operand:    arg1.Address,
	})
	return nil
}

func (t *Translator) ArrayMinusVarLoad(arg1, arg2 symboltable.Symbol, Arg1Index string, label string) error {
	t.emit(code.Instruction{
		Op:         code.SET,
		HasOperand: true,
		Operand:    arg1.Address,
		Label:      label,
	})
	idxSym, err := t.getSymbol(Arg1Index)
	if err != nil {
		return err
	}
	t.emit(code.Instruction{
		Op:         code.ADD,
		HasOperand: true,
		Operand:    idxSym.Address,
	})

	t.emit(code.Instruction{
		Op:         code.SUB,
		HasOperand: true,
		Operand:    arg2.Address,
	})

	return nil
}

func (t *Translator) VarMinusVarLoad(arg1 symboltable.Symbol, arg2 symboltable.Symbol, label string) error {
	t.emit(code.Instruction{
		Op:         code.LOAD,
		HasOperand: true,
		Operand:    arg1.Address,
		Label:      label,
	})

	t.emit(code.Instruction{
		Op:         code.SUB,
		HasOperand: true,
		Operand:    arg2.Address,
	})

	return nil
}

// loadOperandIndirect computes the address of the array element and uses LOADI to load its value into ACC.
func (t *Translator) loadOperandIndirect(operand symboltable.Symbol, operandIndex, label string) error {

	// Compute the address: baseAddr + (index - fromVal)
	indexSymbol, err := t.St.Lookup(fmt.Sprintf("%d", operandIndex), "main")
	if err != nil {
		return fmt.Errorf("failed to lookup the symbol for the index %d: %v", operandIndex, err)
	}

	// Load the index value into ACC
	t.emit(code.Instruction{
		Op:         code.SET,
		Operand:    operand.Address,
		Comment:    fmt.Sprintf("ACC <== Address of %v[0]", operand.Name),
		Label:      label,
		HasOperand: true,
	})
	// Add the base address
	t.emit(code.Instruction{
		Op:         code.ADD,
		Operand:    indexSymbol.Address,
		Comment:    fmt.Sprintf("add index at address to ACC"),
		HasOperand: true,
	})
	// Use LOADI to load the value from the computed address
	t.emit(code.Instruction{
		Op:         code.LOADI,
		Operand:    0,
		Comment:    fmt.Sprintf("ACC' <== %v[ACC]", operand.Name),
		HasOperand: true,
	})
	return nil
}

func (t *Translator) handleArrayLoad(ins tac.Instruction) error {
	// Calculate array address
	base := ins.Arg2.Address
	indexSym, err := t.getSymbol(ins.Arg2Index)
	if err != nil {
		return err
	}

	t.emit(code.Instruction{
		Op:         code.SET,
		HasOperand: true,

		Operand: base,
		Comment: "Array base address",
	})

	t.emit(code.Instruction{
		Op:         code.ADD,
		HasOperand: true,
		Operand:    indexSym.Address,
		Comment:    "Add index offset",
	})

	t.emit(code.Instruction{
		Op:         code.LOADI,
		HasOperand: true,

		Operand: 0,
		Comment: "Load array element",
	})

	t.emit(code.Instruction{
		Op:         code.STORE,
		HasOperand: true,
		Operand:    ins.Arg1.Address,
	})

	return nil
}

func (t *Translator) handleMod(ins tac.Instruction) error {
	panic("unimplemented")
}

func (t *Translator) handleDiv(ins tac.Instruction) error {
	panic("unimplemented")
}

func (t *Translator) handleMul(ins tac.Instruction) error {
	panic("unimplemented")
}

func (t *Translator) handleParam(param symboltable.Symbol, label string) error {

	panic("unimplemented")
}

func (t *Translator) handleCall(ins tac.Instruction) error {
	panic("unimplemented")
}

func (t *Translator) handleRet(label string) error {
	returnAddr, err := t.St.Lookup(t.currentFunctionName+"_return", "xxFunctionsxx")
	if err != nil {
		return fmt.Errorf("failed to get return for function %v: %v", t.currentFunctionName, err)
	}
	t.emit(code.Instruction{Op: code.RTRN, Comment: "ret", Label: label, HasOperand: true, Operand: returnAddr.Address})
	return nil
}

func (t *Translator) handleHalt(label string) {
	t.emit(code.Instruction{Label: label, Comment: "halt", Op: code.HALT})
}

func (t *Translator) handleGoto(labelName, label string) {
	t.emit(code.Instruction{Op: code.JUMP, Comment: "goto " + labelName, Label: label, HasOperand: true, Destination: labelName})
}
func (t *Translator) handleIf(ins tac.Instruction) error {

	if ins.Arg1.IsTable && ins.Arg2.IsTable {
		t.ArrayMinusArrayLoad(*ins.Arg1, *ins.Arg2, ins.Arg1Index, ins.Arg2Index, ins.Label)
	} else if ins.Arg2.IsTable {
		t.VarMinusArrayLoad(*ins.Arg1, *ins.Arg2, ins.Arg2Index, ins.Label)
	} else if ins.Arg1.IsTable {
		t.ArrayMinusVarLoad(*ins.Arg1, *ins.Arg2, ins.Arg1Index, ins.Label)
	} else {
		t.VarMinusVarLoad(*ins.Arg1, *ins.Arg2, ins.Label)
	}
	// 3) Jump if condition is satisfied
	var jumpIns *code.Instruction
	var jumpIns2 *code.Instruction
	switch ins.Op {
	case tac.OpIfLT:
		jumpIns = &code.Instruction{Op: code.JNEG, HasOperand: true, Comment: "to " + ins.Label, Destination: ins.JumpTo}
	case tac.OpIfLE:
		jumpIns = &code.Instruction{Op: code.JNEG, HasOperand: true, Comment: "to " + ins.Label, Destination: ins.JumpTo}
		jumpIns2 = &code.Instruction{Op: code.JZERO, HasOperand: true, Comment: "to " + ins.Label, Destination: ins.JumpTo}
	case tac.OpIfGT:
		jumpIns = &code.Instruction{Op: code.JPOS, HasOperand: true, Comment: "to " + ins.Label, Destination: ins.JumpTo}
	case tac.OpIfGE:
		jumpIns = &code.Instruction{Op: code.JPOS, HasOperand: true, Comment: "to " + ins.Label, Destination: ins.JumpTo}
		jumpIns2 = &code.Instruction{Op: code.JZERO, HasOperand: true, Comment: "to " + ins.Label, Destination: ins.JumpTo}
	case tac.OpIfEQ:
		jumpIns = &code.Instruction{Op: code.JZERO, HasOperand: true, Comment: "to " + ins.Label, Destination: ins.JumpTo}
	case tac.OpIfNE:
		jumpIns = &code.Instruction{Op: code.JPOS, HasOperand: true, Comment: "to " + ins.Label, Destination: ins.JumpTo}
		jumpIns2 = &code.Instruction{Op: code.JNEG, HasOperand: true, Comment: "to " + ins.Label, Destination: ins.JumpTo}
	}
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

func (t *Translator) getSymbol(name string) (*symboltable.Symbol, error) {
	_, err := strconv.Atoi(name)
	if err == nil {
		symbol, err := t.St.Lookup(name, "main")
		if err != nil {
			return nil, fmt.Errorf("failed to find addr for %q: %v", name, err)
		}
		return symbol, nil
	}
	if strings.Contains(name, ".") {
		parts := strings.SplitN(name, ".", 2)
		proc := parts[0]
		sym := parts[1]
		symbol, err := t.St.Lookup(sym, proc)
		if err != nil {
			return nil, fmt.Errorf("failed to find addr for %q: %v", name, err)
		}
		return symbol, nil
	}

	symbol, err := t.St.Lookup(name, t.currentFunctionName)
	if err != nil {
		return nil, fmt.Errorf("failed to find addr for %q: %v", name, err)
	}
	return symbol, nil
}

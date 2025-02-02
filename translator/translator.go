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
	pointerCell              int
	Output                   []code.Instruction
	St                       symboltable.SymbolTable
	currentAddress           int
	currentFunctionName      string
	currentOlderFunctionName string
	procEntries              map[string]int // Adresy początków procedur
	labels                   map[string]int // Etykiety na adresy
	errors                   []string
	paramCount               int
}

func (t *Translator) Errors() []string {
	return t.errors
}

func New(st symboltable.SymbolTable) *Translator {
	return &Translator{pointerCell: st.CurrentOffset + 10, St: st, procEntries: make(map[string]int), labels: make(map[string]int)}
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
		if len(input[i].Labels) != 0 {
			for j, _ := range input[i].Labels {
				labelAddress[input[i].Labels[j]] = i
			}
			input[i].Labels = []string{}
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
		var labels []string
		if len(ins.Labels) != 0 {
			for _, label := range ins.Labels {
				if !strings.HasPrefix(label, "L") && !strings.HasPrefix(label, "built_in") {
					t.currentFunctionName = label // Update the current function name
				} else if !strings.HasPrefix(label, "L") {
					t.currentOlderFunctionName = label
				}
				// You can store t.labels[ins.Label] = t.currentAddress
				// or just emit a comment or no-op, e.g.:
				labels = append(labels, label)
			}
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
			t.handleGoto(destination, labels)

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
			t.handleHalt(labels)
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
			err := t.handleParam(*ins.Arg1, labels)
			if err != nil {
				t.errors = append(t.errors, fmt.Sprintf("Failed to translate %v: %v", ins, err))
			}
		case tac.OpRet:
			err := t.handleRet(labels)
			if err != nil {
				t.errors = append(t.errors, fmt.Sprintf("Failed to translate %v: %v", ins, err))
			}
		case tac.OpDiv:
			err := t.handleDiv(ins)
			if err != nil {
				t.errors = append(t.errors, fmt.Sprintf("Failed to translate %v: %v", ins, err))
			}
		default:
			t.errors = append(t.errors, fmt.Sprintf("Failed to translate %v", ins))
		}

	}
}

func (t *Translator) handleRead(ins tac.Instruction) error {
	if ins.Arg1 == nil {
		panic("nil arg1!!!")
	}
	if ins.Arg1.Kind == symboltable.ARGUMENT {
		if ins.Arg1.IsTable {
			idxSym, err := t.getSymbol(ins.Arg1Index)
			if err != nil {
				return fmt.Errorf("failed to getSymbol for index %q: %v", ins.Arg1Index, err)
			}

			t.emit(code.Instruction{
				Op:         code.LOAD,
				HasOperand: true,
				Operand:    ins.Arg1.Address,
				Labels:     ins.Labels,
				Comment:    "ACC <== &arg1",
			})
			t.emit(code.Instruction{
				Op:         code.ADD,
				HasOperand: true,
				Operand:    idxSym.Address,
			})
			t.emit(code.Instruction{
				Op:         code.STORE,
				HasOperand: true,
				Operand:    t.pointerCell,
				Comment:    fmt.Sprintf("pointerCell = address of %s[%s]", ins.Arg1.Name, ins.Arg1Index),
			})
			t.emit(code.Instruction{
				Op:         code.GET,
				HasOperand: true,
				Operand:    0,
				Comment:    "read input directly into memory[0]",
			})
			t.emit(code.Instruction{
				Op:         code.STOREI,
				HasOperand: true,
				Operand:    t.pointerCell,
				Comment:    "x[n] = input",
			})
		} else {
			t.emit(code.Instruction{
				Op:         code.GET,
				HasOperand: true,

				Labels:  ins.Labels,
				Operand: 0,
				Comment: fmt.Sprintf("memory[0] <== input"),
			})
			t.emit(code.Instruction{
				Op:         code.STOREI,
				HasOperand: true,
				Operand:    ins.Arg1.Address,
				Comment:    fmt.Sprintf("input ==> *%s", ins.Arg1.Name),
			})
			return nil
		}
	} else {
		if ins.Arg1.IsTable {
			//isTable
			t.emit(code.Instruction{
				Op:         code.SET,
				HasOperand: true,
				Operand:    ins.Arg1.Address,
				Labels:     ins.Labels,
			})
			idxSym, err := t.getSymbol(ins.Arg1Index)
			if err != nil {
				return fmt.Errorf("failed to getSymbol for index %q: %v", ins.Arg1Index, err)
			}
			t.emit(code.Instruction{
				Op:         code.ADD,
				HasOperand: true,
				Operand:    idxSym.Address,
			})
			t.emit(code.Instruction{
				Op:         code.STORE,
				HasOperand: true,
				Operand:    t.pointerCell,
				Comment:    "pointerCell = address of x[n]",
			})
			t.emit(code.Instruction{
				Op:         code.GET,
				HasOperand: true,
				Operand:    0,
				Comment:    "read input directly into memory[0]",
			})
			t.emit(code.Instruction{
				Op:         code.STOREI,
				HasOperand: true,
				Operand:    t.pointerCell,
				Comment:    "x[n] = input",
			})
			return nil
		} else {
			//Declaration !isTable
			t.emit(code.Instruction{
				Op:         code.GET,
				HasOperand: true,
				Operand:    ins.Arg1.Address,
				Labels:     ins.Labels,
			})
			return nil
		}
	}
	return nil
}

func (t *Translator) handleWrite(ins tac.Instruction) error {
	if ins.Arg1 == nil {
		panic("WRITE instruction has nil Arg1")
	}
	if ins.Arg1.Kind == symboltable.ARGUMENT {
		if ins.Arg1.IsTable {
			idxSym, err := t.getSymbol(ins.Arg1Index)
			if err != nil {
				return fmt.Errorf("failed to getSymbol for index %q: %v", ins.Arg1Index, err)
			}
			t.emit(code.Instruction{
				Op:         code.LOAD,
				HasOperand: true,
				Operand:    ins.Arg1.Address,
				Labels:     ins.Labels,
				Comment:    "ACC <== &arg1",
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
				Comment:    "ACC = val of x[n]",
			})
		} else {
			t.emit(code.Instruction{
				Op:         code.LOADI,
				HasOperand: true,
				Operand:    ins.Arg1.Address,
				Labels:     ins.Labels,
				Comment:    "load what argument is pointint to into ACC",
			})
		}

		t.emit(code.Instruction{
			Op:         code.PUT,
			HasOperand: true,
			Operand:    0,
			Comment:    "write x[n]",
		})
		return nil
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
			Labels:     ins.Labels,
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
		ins := code.Instruction{Op: code.PUT, HasOperand: true, Labels: ins.Labels, Operand: ins.Arg1.Address}
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
	label := ins.Labels
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

func (t *Translator) handleVarToVarAssign(dest, src symboltable.Symbol, label []string) error {
	destArgument := dest.Kind == symboltable.ARGUMENT
	srcArgument := src.Kind == symboltable.ARGUMENT
	// Load src into ACC
	if srcArgument {
		t.emit(code.Instruction{
			Op:         code.LOADI,
			HasOperand: true,
			Operand:    src.Address,
			Comment:    fmt.Sprintf("load %v into ACC", src),
			Labels:     label,
		})
	} else {
		t.emit(code.Instruction{
			Op:         code.LOAD,
			HasOperand: true,
			Operand:    src.Address,
			Comment:    fmt.Sprintf("load %v into ACC", src),
			Labels:     label,
		})
	}

	// Store ACC into dest
	if destArgument {
		t.emit(code.Instruction{
			Op:         code.STOREI,
			HasOperand: true,
			Operand:    dest.Address,
			Comment:    fmt.Sprintf("store ACC into %v", dest),
		})
	} else {
		t.emit(code.Instruction{
			Op:         code.STORE,
			HasOperand: true,
			Operand:    dest.Address,
			Comment:    fmt.Sprintf("store ACC into %v", dest),
		})
	}
	return nil
}

func (t *Translator) handleVarToArrayAssign(dest, src symboltable.Symbol, destIndex string, label []string) error {
	destArgument := dest.Kind == symboltable.ARGUMENT
	srcArgument := src.Kind == symboltable.ARGUMENT

	if destArgument {
		t.emit(code.Instruction{
			Op:         code.LOAD,
			Operand:    dest.Address,
			Comment:    fmt.Sprintf("ACC <== Address of %v[0]", dest.Name),
			Labels:     label,
			HasOperand: true,
		})
	} else {
		// Load the index value into ACC
		t.emit(code.Instruction{
			Op:         code.SET,
			Operand:    dest.Address,
			Comment:    fmt.Sprintf("ACC <== Address of %v[0]", dest.Name),
			Labels:     label,
			HasOperand: true,
		})
	}
	// Add the base address
	indexSymbol, err := t.getSymbol(destIndex)
	if err != nil {
		return fmt.Errorf("NIEPRAWIDLOWE UZYCIE TABLICY: %v", err)
	}
	if indexSymbol.Kind == symboltable.ARGUMENT {
		t.emit(code.Instruction{
			Op:         code.STORE,
			Operand:    t.pointerCell,
			HasOperand: true,
		})
		t.emit(code.Instruction{
			Op:         code.LOADI,
			Operand:    indexSymbol.Address,
			HasOperand: true,
		})
		t.emit(code.Instruction{
			Op:         code.ADD,
			Operand:    t.pointerCell,
			HasOperand: true,
		})
	} else {
		t.emit(code.Instruction{
			Op:         code.ADD,
			Operand:    indexSymbol.Address,
			Comment:    fmt.Sprintf("add index (%s) to ACC", destIndex),
			HasOperand: true,
		})
	}

	t.emit(code.Instruction{
		Op:         code.STORE,
		HasOperand: true,
		Operand:    t.pointerCell,
		Comment:    fmt.Sprintf("pc = &%s[%s]", dest.Name, destIndex),
	})
	if srcArgument {
		t.emit(code.Instruction{
			Op:         code.LOADI,
			HasOperand: true,
			Operand:    src.Address,
			Comment:    fmt.Sprintf("load %v into ACC", src.Name),
		})
	} else {
		t.emit(code.Instruction{
			Op:         code.LOAD,
			HasOperand: true,
			Operand:    src.Address,
			Comment:    fmt.Sprintf("load %v into ACC", src.Name),
		})
	}
	t.emit(code.Instruction{
		Op:         code.STOREI,
		HasOperand: true,
		Operand:    t.pointerCell,
		Comment:    "p[pointerCell] <== ACC",
	})

	return nil
}

func (t *Translator) handleArrayToVarAssign(dest, src symboltable.Symbol, srcIndex string, label []string) error {
	destArgument := dest.Kind == symboltable.ARGUMENT
	if err := t.loadOperandIndirect(src, srcIndex, label); err != nil {
		return err
	}
	if destArgument {
		t.emit(code.Instruction{
			Op:         code.STOREI,
			HasOperand: true,
			Operand:    dest.Address,
			Comment:    "dest <== ACC (src)",
			Labels:     label,
		})
	} else {
		t.emit(code.Instruction{
			Op:         code.STORE,
			HasOperand: true,
			Operand:    dest.Address,
			Comment:    "dest <== ACC (src)",
		})
	}
	return nil
}

func (t *Translator) handleArrayToArrayAssign(dest, src symboltable.Symbol, srcIndex, destIndex string, labels []string) error {
	destArgument := dest.Kind == symboltable.ARGUMENT
	if destArgument {
		t.emit(code.Instruction{
			Op:         code.LOAD,
			Operand:    dest.Address,
			HasOperand: true,
			Labels:     labels,
		})
	} else {
		t.emit(code.Instruction{
			Op:         code.SET,
			Operand:    dest.Address,
			HasOperand: true,
			Labels:     labels,
		})
	}
	indexSymbol, err := t.getSymbol(destIndex)
	if err != nil {
		return fmt.Errorf("NIEPRAWIDLOWE UZYCIE TABLICY: %v", err)

	}
	if indexSymbol.Kind == symboltable.ARGUMENT {
		t.emit(code.Instruction{
			Op:         code.STORE,
			Operand:    t.pointerCell,
			HasOperand: true,
		})
		t.emit(code.Instruction{
			Op:         code.LOADI,
			Operand:    indexSymbol.Address,
			HasOperand: true,
		})
		t.emit(code.Instruction{
			Op:         code.ADD,
			Operand:    t.pointerCell,
			HasOperand: true,
		})
	} else {
		t.emit(code.Instruction{
			Op:         code.ADD,
			Operand:    indexSymbol.Address,
			Comment:    fmt.Sprintf("add index at address to ACC"),
			HasOperand: true,
		})
	}
	t.emit(code.Instruction{
		Op:         code.STORE,
		Operand:    t.pointerCell,
		HasOperand: true,
	})
	err = t.loadOperandIndirect(src, srcIndex, []string{})
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
	label := ins.Labels
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

func (t *Translator) handleArrayAddSubArray(op tac.Op, dest symboltable.Symbol, arg1, arg2 symboltable.Symbol, Arg1Index, Arg2Index string, labels []string) error {
	destArgument := dest.Kind == symboltable.ARGUMENT
	arg1Argument := arg1.Kind == symboltable.ARGUMENT
	arg2Argument := arg2.Kind == symboltable.ARGUMENT
	if arg1Argument {
		t.emit(code.Instruction{
			Op:         code.LOAD,
			HasOperand: true,
			Operand:    arg1.Address,
			Labels:     labels,
		})
	} else {
		t.emit(code.Instruction{
			Op:         code.SET,
			HasOperand: true,
			Operand:    arg1.Address,
			Labels:     labels,
		})
	}
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
		Op:         code.LOADI,
		HasOperand: true,
		Operand:    0,
	})

	t.emit(code.Instruction{
		Op:         code.STORE,
		HasOperand: true,
		Operand:    t.pointerCell,
	})
	if arg2Argument {
		t.emit(code.Instruction{
			Op:         code.LOAD,
			HasOperand: true,
			Operand:    arg2.Address,
		})
	} else {
		t.emit(code.Instruction{
			Op:         code.SET,
			HasOperand: true,
			Operand:    arg2.Address,
		})
	}
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
	if destArgument {
		t.emit(code.Instruction{
			Op:         code.STOREI,
			HasOperand: true,
			Operand:    dest.Address,
		})
	} else {

		t.emit(code.Instruction{
			Op:         code.STORE,
			HasOperand: true,
			Operand:    dest.Address,
		})
	}
	return nil
}

func (t *Translator) handleVarAddSubArray(op tac.Op, dest symboltable.Symbol, arg1, arg2 symboltable.Symbol, Arg2Index string, labels []string) error {
	destArgument := dest.Kind == symboltable.ARGUMENT
	arg1Argument := arg1.Kind == symboltable.ARGUMENT
	arg2Argument := arg2.Kind == symboltable.ARGUMENT
	if arg2Argument {
		t.emit(code.Instruction{
			Op:         code.LOAD,
			HasOperand: true,
			Operand:    arg2.Address,
			Labels:     labels,
		})
	} else {
		t.emit(code.Instruction{
			Op:         code.SET,
			HasOperand: true,
			Operand:    arg2.Address,
			Labels:     labels,
		})
	}
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
		if arg1Argument {
			t.emit(code.Instruction{
				Op:         code.ADDI,
				HasOperand: true,
				Operand:    arg1.Address,
			})
		} else {
			t.emit(code.Instruction{
				Op:         code.ADD,
				HasOperand: true,
				Operand:    arg1.Address,
			})
		}
	} else {
		if arg2Argument {
			t.emit(code.Instruction{
				Op:         code.SUBI,
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
	}
	if destArgument {
		t.emit(code.Instruction{
			Op:         code.STOREI,
			HasOperand: true,
			Operand:    dest.Address,
		})
	} else {
		t.emit(code.Instruction{
			Op:         code.STORE,
			HasOperand: true,
			Operand:    dest.Address,
		})
	}
	return nil
}

func (t *Translator) handleArrayAddSubVar(op tac.Op, dest symboltable.Symbol, arg1, arg2 symboltable.Symbol, Arg1Index string, labels []string) error {
	destArgument := dest.Kind == symboltable.ARGUMENT
	arg1Argument := arg1.Kind == symboltable.ARGUMENT
	arg2Argument := arg2.Kind == symboltable.ARGUMENT
	if arg1Argument {
		t.emit(code.Instruction{
			Op:         code.LOAD,
			HasOperand: true,
			Operand:    arg1.Address,
			Labels:     labels,
		})
	} else {
		t.emit(code.Instruction{
			Op:         code.SET,
			HasOperand: true,
			Operand:    arg1.Address,
			Labels:     labels,
		})
	}
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
		if arg2Argument {
			t.emit(code.Instruction{
				Op:         code.ADDI,
				HasOperand: true,
				Operand:    arg2.Address,
			})
		} else {
			t.emit(code.Instruction{
				Op:         code.ADD,
				HasOperand: true,
				Operand:    arg2.Address,
			})
		}
	} else {
		if arg2Argument {
			t.emit(code.Instruction{
				Op:         code.SUBI,
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
	}
	if destArgument {
		t.emit(code.Instruction{
			Op:         code.STOREI,
			HasOperand: true,
			Operand:    dest.Address,
		})
	} else {
		t.emit(code.Instruction{
			Op:         code.STORE,
			HasOperand: true,
			Operand:    dest.Address,
		})
	}
	return nil
}

func (t *Translator) handleVarToVarAddSub(op tac.Op, dest symboltable.Symbol, arg1 symboltable.Symbol, arg2 symboltable.Symbol, labels []string) error {
	destArgument := dest.Kind == symboltable.ARGUMENT
	arg1Argument := arg1.Kind == symboltable.ARGUMENT
	arg2Argument := arg2.Kind == symboltable.ARGUMENT
	if arg1Argument {
		t.emit(code.Instruction{
			Op:         code.LOADI,
			HasOperand: true,
			Operand:    arg1.Address,
			Labels:     labels,
		})
	} else {
		t.emit(code.Instruction{
			Op:         code.LOAD,
			HasOperand: true,
			Operand:    arg1.Address,
			Labels:     labels,
		})
	}
	if arg2Argument {
		if op == tac.OpAdd {
			t.emit(code.Instruction{
				Op:         code.ADDI,
				HasOperand: true,
				Operand:    arg2.Address,
			})
		} else {
			t.emit(code.Instruction{
				Op:         code.SUBI,
				HasOperand: true,
				Operand:    arg2.Address,
			})
		}
	} else {
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
	}
	if destArgument {
		t.emit(code.Instruction{
			Op:         code.STOREI,
			HasOperand: true,
			Operand:    dest.Address,
		})
	} else {
		t.emit(code.Instruction{
			Op:         code.STORE,
			HasOperand: true,
			Operand:    dest.Address,
		})
	}
	return nil
}

func (t *Translator) ArrayMinusArrayLoad(arg1, arg2 symboltable.Symbol, Arg1Index, Arg2Index string, labels []string) error {
	t.emit(code.Instruction{
		Op:         code.SET,
		HasOperand: true,
		Operand:    arg1.Address,
		Labels:     labels,
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

func (t *Translator) VarMinusArrayLoad(arg1, arg2 symboltable.Symbol, Arg2Index string, labels []string) error {
	t.emit(code.Instruction{
		Op:         code.SET,
		HasOperand: true,
		Operand:    arg2.Address,
		Labels:     labels,
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

func (t *Translator) ArrayMinusVarLoad(arg1, arg2 symboltable.Symbol, Arg1Index string, labels []string) error {
	t.emit(code.Instruction{
		Op:         code.SET,
		HasOperand: true,
		Operand:    arg1.Address,
		Labels:     labels,
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

func (t *Translator) VarMinusVarLoad(arg1 symboltable.Symbol, arg2 symboltable.Symbol, labels []string) error {
	t.emit(code.Instruction{
		Op:         code.LOAD,
		HasOperand: true,
		Operand:    arg1.Address,
		Labels:     labels,
	})

	t.emit(code.Instruction{
		Op:         code.SUB,
		HasOperand: true,
		Operand:    arg2.Address,
	})

	return nil
}

// loadOperandIndirect computes the address of the array element and uses LOADI to load its value into ACC.
func (t *Translator) loadOperandIndirect(operand symboltable.Symbol, operandIndex string, labels []string) error {

	// Compute the address: baseAddr + (index - fromVal)
	indexSymbol, err := t.St.Lookup(operandIndex, t.currentFunctionName)
	if err != nil {
		return fmt.Errorf("NIEWLASCIWE UZYCIE TABLICY: failed to lookup the symbol for the index %d: %v", operandIndex, err)
	}
	if operand.Kind == symboltable.ARGUMENT {
		t.emit(code.Instruction{
			Op:         code.LOAD,
			Operand:    operand.Address,
			Comment:    fmt.Sprintf("ACC <== Address of %v[0]", operand.Name),
			Labels:     labels,
			HasOperand: true,
		})
	} else {
		// Load the index value into ACC
		t.emit(code.Instruction{
			Op:         code.SET,
			Operand:    operand.Address,
			Comment:    fmt.Sprintf("ACC <== Address of %v[0]", operand.Name),
			Labels:     labels,
			HasOperand: true,
		})
	}
	// Add the base address
	if indexSymbol.Kind == symboltable.ARGUMENT {
		t.emit(code.Instruction{
			Op:         code.STORE,
			Operand:    t.pointerCell,
			HasOperand: true,
		})
		t.emit(code.Instruction{
			Op:         code.LOADI,
			Operand:    indexSymbol.Address,
			HasOperand: true,
		})
		t.emit(code.Instruction{
			Op:         code.ADD,
			Operand:    t.pointerCell,
			HasOperand: true,
		})
	} else {
		t.emit(code.Instruction{
			Op:         code.ADD,
			Operand:    indexSymbol.Address,
			Comment:    fmt.Sprintf("add index at address to ACC"),
			HasOperand: true,
		})
	}
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
	if ins.Arg2.Name == "2" {
		destArgument := ins.Destination.Kind == symboltable.ARGUMENT
		arg1Argument := ins.Arg1.Kind == symboltable.ARGUMENT
		if arg1Argument {
			t.emit(code.Instruction{
				Op:         code.LOADI,
				HasOperand: true,
				Labels:     ins.Labels,
				Operand:    ins.Arg1.Address,
			})
		} else {
			t.emit(code.Instruction{
				Op:         code.LOAD,
				HasOperand: true,
				Labels:     ins.Labels,
				Operand:    ins.Arg1.Address,
			})
		}
		t.emit(code.Instruction{
			Op: code.HALF,
		})
		if destArgument {
			t.emit(code.Instruction{
				Op:         code.STOREI,
				HasOperand: true,
				Operand:    ins.Destination.Address,
			})
		} else {
			t.emit(code.Instruction{
				Op:         code.STORE,
				HasOperand: true,
				Operand:    ins.Destination.Address,
			})
		}
		return nil
	}
	return fmt.Errorf("unimplemented :)")
}

func (t *Translator) handleParam(param symboltable.Symbol, labels []string) error {
	if param.Kind == symboltable.ARGUMENT {
		t.emit(code.Instruction{
			Op:         code.LOAD,
			HasOperand: true,
			Labels:     labels,
			Operand:    param.Address,
		})
	} else {
		t.emit(code.Instruction{
			Op:         code.SET,
			HasOperand: true,
			Labels:     labels,
			Operand:    param.Address,
		})
	}
	t.emit(code.Instruction{
		Op:         code.STORE,
		HasOperand: true,
		Operand:    t.pointerCell + 10 + t.paramCount,
	})
	t.paramCount++
	return nil
}

func (t *Translator) handleCall(ins tac.Instruction) error {
	procSym, err := t.St.Lookup(ins.Arg1.Name, "xxFunctionsxx")
	if err != nil {
		return fmt.Errorf("failed finding a functon called %s", ins.Arg1.Name)
	}
	argCount := procSym.ArgCount
	for i := argCount; i > 0; i-- {
		if i == argCount {
			t.emit(code.Instruction{
				Op:         code.LOAD,
				HasOperand: true,
				Labels:     ins.Labels,
				Operand:    t.pointerCell + 10 + t.paramCount - i,
			})
		} else {
			t.emit(code.Instruction{
				Op:         code.LOAD,
				HasOperand: true,
				Operand:    t.pointerCell + 10 + t.paramCount - i,
			})
		}
		t.emit(code.Instruction{
			Op:         code.STORE,
			HasOperand: true,
			Operand:    procSym.Arguments[argCount-i].Address,
		})
	}
	returnSymbol, err := t.St.Lookup(ins.Arg1.Name+"_return", "xxFunctionsxx")
	if err != nil {
		return fmt.Errorf("failed finding a functon return called %s", ins.Arg1.Name)
	}
	if argCount == 0 {
		t.emit(code.Instruction{
			Op:         code.SET,
			HasOperand: true,
			Operand:    t.currentAddress + 3,
			Labels:     ins.Labels,
		})
	} else {
		t.emit(code.Instruction{
			Op:         code.SET,
			HasOperand: true,
			Operand:    len(t.Output) + 3,
		})
	}
	t.emit(code.Instruction{
		Op:         code.STORE,
		HasOperand: true,
		Operand:    returnSymbol.Address,
	})
	t.emit(code.Instruction{
		Op:          code.JUMP,
		Destination: procSym.Name,
	})

	return nil
}

func (t *Translator) handleRet(labels []string) error {
	returnAddr, err := t.St.Lookup(t.currentFunctionName+"_return", "xxFunctionsxx")
	if err != nil {
		returnAddr, err = t.St.Lookup(t.currentOlderFunctionName+"_return", "xxFunctionsxx")
		if err != nil {
			return fmt.Errorf("failed to get return for function %v: %v", t.currentOlderFunctionName, err)
		}
	}
	t.emit(code.Instruction{Op: code.RTRN, Comment: "ret", Labels: labels, HasOperand: true, Operand: returnAddr.Address})
	return nil
}

func (t *Translator) handleHalt(labels []string) {
	t.emit(code.Instruction{Labels: labels, Comment: "halt", Op: code.HALT})
}

func (t *Translator) handleGoto(labelName string, labels []string) {
	t.emit(code.Instruction{Op: code.JUMP, Comment: "goto " + labelName, Labels: labels, HasOperand: true, Destination: labelName})
}
func (t *Translator) handleIf(ins tac.Instruction) error {

	if ins.Arg1.IsTable && ins.Arg2.IsTable {
		t.ArrayMinusArrayLoad(*ins.Arg1, *ins.Arg2, ins.Arg1Index, ins.Arg2Index, ins.Labels)
	} else if ins.Arg2.IsTable {
		t.VarMinusArrayLoad(*ins.Arg1, *ins.Arg2, ins.Arg2Index, ins.Labels)
	} else if ins.Arg1.IsTable {
		t.ArrayMinusVarLoad(*ins.Arg1, *ins.Arg2, ins.Arg1Index, ins.Labels)
	} else {
		t.VarMinusVarLoad(*ins.Arg1, *ins.Arg2, ins.Labels)
	}
	// 3) Jump if condition is satisfied
	var jumpIns *code.Instruction
	var jumpIns2 *code.Instruction
	switch ins.Op {
	case tac.OpIfLT:
		jumpIns = &code.Instruction{Op: code.JNEG, HasOperand: true, Comment: "to " + ins.JumpTo, Destination: ins.JumpTo}
	case tac.OpIfLE:
		jumpIns = &code.Instruction{Op: code.JNEG, HasOperand: true, Comment: "to " + ins.JumpTo, Destination: ins.JumpTo}
		jumpIns2 = &code.Instruction{Op: code.JZERO, HasOperand: true, Comment: "to " + ins.JumpTo, Destination: ins.JumpTo}
	case tac.OpIfGT:
		jumpIns = &code.Instruction{Op: code.JPOS, HasOperand: true, Comment: "to " + ins.JumpTo, Destination: ins.JumpTo}
	case tac.OpIfGE:
		jumpIns = &code.Instruction{Op: code.JPOS, HasOperand: true, Comment: "to " + ins.JumpTo, Destination: ins.JumpTo}
		jumpIns2 = &code.Instruction{Op: code.JZERO, HasOperand: true, Comment: "to " + ins.JumpTo, Destination: ins.JumpTo}
	case tac.OpIfEQ:
		jumpIns = &code.Instruction{Op: code.JZERO, HasOperand: true, Comment: "to " + ins.JumpTo, Destination: ins.JumpTo}
	case tac.OpIfNE:
		jumpIns = &code.Instruction{Op: code.JPOS, HasOperand: true, Comment: "to " + ins.JumpTo, Destination: ins.JumpTo}
		jumpIns2 = &code.Instruction{Op: code.JNEG, HasOperand: true, Comment: "to " + ins.JumpTo, Destination: ins.JumpTo}
	}
	if jumpIns != nil {
		t.emit(*jumpIns)
	}
	if jumpIns2 != nil {
		t.emit(*jumpIns2)
	}

	return nil
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

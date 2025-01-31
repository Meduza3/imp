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
	pointerCell         int
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
	return &Translator{pointerCell: 69, St: st, procEntries: make(map[string]int), labels: make(map[string]int)}
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
		fmt.Println("# ins: ", ins)
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

			err := t.handleAssign(dest, src, label)
			if err != nil {
				t.errors = append(t.errors, fmt.Sprintf("failed to handle Assign %v: %v", ins, err))
			}
		//----------------------------------------------------------------------
		// 2) Arithmetic: a = b + c, a = b - c, etc.
		//----------------------------------------------------------------------
		case tac.OpAdd, tac.OpSub:
			// For example:  ins = { Op: OpAdd, Destination: "x", Arg1: "a", Arg2: "b" }
			err := t.handleAddSub(ins.Op, ins.Destination, ins.Arg1, ins.Arg2, label)
			if err != nil {
				t.errors = append(t.errors, fmt.Sprintf("failed to handleAddSub %s: %v", ins, err))
			}
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
	if isArrayRef(arg1) {
		base, idx, err := parseArrayRef(arg1)
		if err != nil {
			return fmt.Errorf("failed parsing array ref: %v", err)
		}
		arrSymbol, err := t.getSymbol(base)
		if err != nil {
			return err
		}
		baseAddr := arrSymbol.Address
		fromVal := arrSymbol.From
		fromValSymbol, err := t.getSymbol(fmt.Sprintf("%d", fromVal))
		if err != nil {
			return err
		}
		idxSym, err := t.getSymbol(idx)
		if err != nil {
			return err
		}
		pointerCell := t.pointerCell
		scratchCell := t.pointerCell + 1 // or any free cell

		// --- Compute pointerCell = base + (idx - fromVal) ---
		t.emit(code.Instruction{
			Op:         code.LOAD,
			HasOperand: true,
			Operand:    idxSym.Address,
			Label:      label,
			Comment:    "index variable " + idx,
		})
		if fromVal != 0 {
			t.emit(code.Instruction{
				Op:         code.SUB,
				HasOperand: true,
				Operand:    fromValSymbol.Address,
				Comment:    "subtract array.from for " + arrSymbol.Name,
			})
		}
		t.emit(code.Instruction{
			Op:         code.ADD,
			HasOperand: true,
			Operand:    baseAddr,
			Comment:    "add base address for " + arrSymbol.Name,
		})
		t.emit(code.Instruction{
			Op:         code.STORE,
			HasOperand: true,
			Operand:    pointerCell,
		})

		// --- Read into scratchCell ---
		t.emit(code.Instruction{
			Op:         code.GET,
			HasOperand: true,
			Operand:    scratchCell,
			Comment:    "read -> scratch cell",
		})

		// --- Then copy scratchCell => memory[memory[pointerCell]] ---
		t.emit(code.Instruction{
			Op:         code.LOAD,
			HasOperand: true,
			Operand:    scratchCell,
		})
		t.emit(code.Instruction{
			Op:         code.STOREI,
			HasOperand: true,
			Operand:    pointerCell,
			Comment:    "store scratch => array element",
		})
	} else {
		// Plain variable read
		addr, err := t.getAddr(arg1)
		if err != nil {
			return fmt.Errorf("Failed to get addr of %q", arg1)
		}
		t.emit(code.Instruction{
			Op:         code.GET,
			HasOperand: true,
			Operand:    addr,
			Comment:    "read " + arg1,
			Label:      label,
		})
	}
	return nil
}

func (t *Translator) handleWrite(arg1, label string) error {
	if isArrayRef(arg1) {
		base, idx, err := parseArrayRef(arg1)
		if err != nil {
			return fmt.Errorf("failed to parse arg1 array ref: %v", err)
		}
		arrSymbol, err := t.getSymbol(base)
		if err != nil {
			return fmt.Errorf("failed to get symbol for array base: %v", err)
		}
		baseAddr := arrSymbol.Address
		fromVal := arrSymbol.From
		fromValSymbol, err := t.getSymbol(fmt.Sprintf("%d", fromVal))
		if err != nil {
			return fmt.Errorf("failed to get address of fromVal %d", fromVal)
		}
		idxSym, err := t.getSymbol(idx)
		if err != nil {
			return fmt.Errorf("failed to get symbol for array index: %v", err)
		}
		pointerCell := t.pointerCell // Use a dedicated temporary cell

		// Compute the address: baseAddr + (idx - fromVal)
		t.emit(code.Instruction{
			Op:         code.LOAD,
			HasOperand: true,
			Operand:    idxSym.Address,
			Comment:    "load index variable " + idx,
		})
		if fromVal != 0 {
			t.emit(code.Instruction{
				Op:         code.SUB,
				HasOperand: true,
				Operand:    fromValSymbol.Address,
				Comment:    "subtract array.from for " + arrSymbol.Name,
			})
		}
		t.emit(code.Instruction{
			Op:         code.ADD,
			HasOperand: true,
			Operand:    baseAddr,
			Comment:    "add base address for " + arrSymbol.Name,
		})
		t.emit(code.Instruction{
			Op:         code.STORE,
			HasOperand: true,
			Operand:    pointerCell,
			Comment:    "store computed address in pointerCell",
		})

		// Load the value from the computed address
		t.emit(code.Instruction{
			Op:         code.LOADI,
			HasOperand: true,
			Operand:    pointerCell,
			Comment:    "load x[n] into ACC",
		})
		t.emit(code.Instruction{
			Op:         code.STORE,
			HasOperand: true,
			Operand:    pointerCell + 1, // Use a separate temp for storing the value
			Comment:    "store ACC to temp",
		})
		t.emit(code.Instruction{
			Op:         code.PUT,
			HasOperand: true,
			Operand:    pointerCell + 1, // Reference the correct temp cell
			Comment:    "write x[n]",
		})
	} else {
		addr, err := t.getAddr(arg1)
		if err != nil {
			return fmt.Errorf("Failed to get addr of %q", arg1)
		}
		// emit "PUT addr"
		ins := code.Instruction{Op: code.PUT, Comment: "write " + arg1, HasOperand: true, Label: label, Operand: addr}
		t.emit(ins)
	}
	return nil
}

func isArrayRef(s string) bool {
	return strings.Contains(s, "[") && strings.Contains(s, "]")
}

// parseArrayRef("main.a[i]") => ("main.a", "i")
// or parseArrayRef("a[x]")   => ("a", "x")
func parseArrayRef(s string) (base string, index string, err error) {
	// e.g. s = "main.a[i]" or "a[i]"
	// find '[' and ']'
	lb := strings.IndexRune(s, '[')
	rb := strings.IndexRune(s, ']')
	if lb < 0 || rb < 0 || lb > rb {
		return "", "", fmt.Errorf("Invalid array syntax: %q", s)
	}
	base = s[:lb]        // "main.a"
	index = s[lb+1 : rb] // "i"
	return base, index, nil
}

func (t *Translator) handleAssign(dest, src, label string) error {
	srcIsArray := isArrayRef(src)
	destIsArray := isArrayRef(dest)

	if srcIsArray && destIsArray {
		// Case 4: Array Element to Array Element (x[n] := y[m])
		return t.handleArrayToArrayAssign(dest, src, label)
	} else if srcIsArray {
		// Case 3: Array Element to Variable (a := x[n])
		return t.handleArrayToVarAssign(dest, src, label)
	} else if destIsArray {
		// Case 2: Variable to Array Element (x[n] := b)
		return t.handleVarToArrayAssign(dest, src, label)
	} else {
		// Case 1: Variable to Variable (a := b)
		return t.handleVarToVarAssign(dest, src, label)
	}
}

func (t *Translator) handleVarToVarAssign(dest, src, label string) error {
	srcAddr, err := t.getAddr(src)
	if err != nil {
		return fmt.Errorf("failed to get src address: %v", err)
	}
	destAddr, err := t.getAddr(dest)
	if err != nil {
		return fmt.Errorf("failed to get dest address: %v", err)
	}

	// Load src into ACC
	t.emit(code.Instruction{
		Op:         code.LOAD,
		HasOperand: true,
		Operand:    srcAddr,
		Comment:    fmt.Sprintf("load %s into ACC", src),
		Label:      label,
	})

	// Store ACC into dest
	t.emit(code.Instruction{
		Op:         code.STORE,
		HasOperand: true,
		Operand:    destAddr,
		Comment:    fmt.Sprintf("store ACC into %s", dest),
	})

	return nil
}

func (t *Translator) handleVarToArrayAssign(dest, src, label string) error {
	base, idx, err := parseArrayRef(dest)
	if err != nil {
		return fmt.Errorf("failed to parse array ref for dest %s: %v", dest, err)
	}
	arrSymbol, err := t.getSymbol(base)
	if err != nil {
		return fmt.Errorf("failed to get symbol for array base %s: %v", base, err)
	}
	baseAddr := arrSymbol.Address
	fromVal := arrSymbol.From
	fromValSymbol, err := t.getSymbol(fmt.Sprintf("%d", fromVal))
	if err != nil {
		return fmt.Errorf("failed to get symbol for array lower bound: %v", err)
	}
	idxSym, err := t.getSymbol(idx)
	if err != nil {
		return fmt.Errorf("failed to get symbol for index %s: %v", idx, err)
	}
	pointerCell := t.pointerCell   // e.g., 69
	tempValue := t.pointerCell + 1 // e.g., 70

	// Step 1: Load src into ACC
	srcAddr, err := t.getAddr(src)
	if err != nil {
		return fmt.Errorf("failed to get src address: %v", err)
	}
	t.emit(code.Instruction{
		Op:         code.LOAD,
		HasOperand: true,
		Operand:    srcAddr,
		Label:      label,
		Comment:    fmt.Sprintf("load %s into ACC", src),
	})

	// Step 2: Store ACC into tempValue
	t.emit(code.Instruction{
		Op:         code.STORE,
		HasOperand: true,
		Operand:    tempValue,
		Comment:    "store ACC into tempValue",
	})

	// Step 3: Compute the address of x[n]
	t.emit(code.Instruction{
		Op:         code.LOAD,
		HasOperand: true,
		Operand:    idxSym.Address,
		Comment:    fmt.Sprintf("load index variable %s", idx),
	})
	if fromVal != 0 {
		t.emit(code.Instruction{
			Op:         code.SUB,
			HasOperand: true,
			Operand:    fromValSymbol.Address,
			Comment:    fmt.Sprintf("subtract array.from (%d) for %s", fromVal, base),
		})
	}
	t.emit(code.Instruction{
		Op:         code.ADD,
		HasOperand: true,
		Operand:    baseAddr,
		Comment:    fmt.Sprintf("add base address (%d) for %s", baseAddr, base),
	})
	t.emit(code.Instruction{
		Op:         code.STORE,
		HasOperand: true,
		Operand:    pointerCell,
		Comment:    "store computed address into pointerCell",
	})

	// Step 4: Load preserved value from tempValue into ACC
	t.emit(code.Instruction{
		Op:         code.LOAD,
		HasOperand: true,
		Operand:    tempValue,
		Comment:    "load preserved value from tempValue into ACC",
	})

	// Step 5: Store ACC into x[n] via STOREI
	t.emit(code.Instruction{
		Op:         code.STOREI,
		HasOperand: true,
		Operand:    pointerCell,
		Comment:    fmt.Sprintf("store ACC into %s[%s]", base, idx),
	})

	return nil
}

func (t *Translator) handleArrayToVarAssign(dest, src, label string) error {
	base, idx, err := parseArrayRef(src)
	if err != nil {
		return fmt.Errorf("failed to parse array ref for src %s: %v", src, err)
	}
	arrSymbol, err := t.getSymbol(base)
	if err != nil {
		return fmt.Errorf("failed to get symbol for array base %s: %v", base, err)
	}
	baseAddr := arrSymbol.Address
	fromVal := arrSymbol.From
	fromValSymbol, err := t.getSymbol(fmt.Sprintf("%d", fromVal))
	if err != nil {
		return fmt.Errorf("failed to get symbol for array lower bound: %v", err)
	}
	idxSym, err := t.getSymbol(idx)
	if err != nil {
		return fmt.Errorf("failed to get symbol for index %s: %v", idx, err)
	}
	pointerCell := t.pointerCell   // e.g., 69
	tempValue := t.pointerCell + 1 // e.g., 70

	// Step 1: Compute the address of x[n]
	t.emit(code.Instruction{
		Op:         code.LOAD,
		HasOperand: true,
		Operand:    idxSym.Address,
		Label:      label,
		Comment:    fmt.Sprintf("load index variable %s", idx),
	})
	if fromVal != 0 {
		t.emit(code.Instruction{
			Op:         code.SUB,
			HasOperand: true,
			Operand:    fromValSymbol.Address,
			Comment:    fmt.Sprintf("subtract array.from (%d) for %s", fromVal, base),
		})
	}
	t.emit(code.Instruction{
		Op:         code.ADD,
		HasOperand: true,
		Operand:    baseAddr,
		Comment:    fmt.Sprintf("add base address (%d) for %s", baseAddr, base),
	})
	t.emit(code.Instruction{
		Op:         code.STORE,
		HasOperand: true,
		Operand:    pointerCell,
		Comment:    "store computed address into pointerCell",
	})

	// Step 2: Load value from x[n] into ACC via LOADI
	t.emit(code.Instruction{
		Op:         code.LOADI,
		HasOperand: true,
		Operand:    pointerCell,
		Comment:    fmt.Sprintf("load %s[%s] into ACC", base, idx),
	})

	// Step 3: Store ACC into tempValue
	t.emit(code.Instruction{
		Op:         code.STORE,
		HasOperand: true,
		Operand:    tempValue,
		Comment:    "store ACC into tempValue",
	})

	// Step 4: Load tempValue into ACC
	t.emit(code.Instruction{
		Op:         code.LOAD,
		HasOperand: true,
		Operand:    tempValue,
		Comment:    "load preserved value from tempValue into ACC",
	})

	// Step 5: Store ACC into dest variable
	destAddr, err := t.getAddr(dest)
	if err != nil {
		return fmt.Errorf("failed to get dest address: %v", err)
	}
	t.emit(code.Instruction{
		Op:         code.STORE,
		HasOperand: true,
		Operand:    destAddr,
		Comment:    fmt.Sprintf("store ACC into %s", dest),
	})

	return nil
}

func (t *Translator) handleArrayToArrayAssign(dest, src, label string) error {
	// Parse destination array reference
	destBase, destIdx, err := parseArrayRef(dest)
	if err != nil {
		return fmt.Errorf("failed to parse array ref for dest %s: %v", dest, err)
	}
	destArrSymbol, err := t.getSymbol(destBase)
	if err != nil {
		return fmt.Errorf("failed to get symbol for dest array base %s: %v", destBase, err)
	}
	destBaseAddr := destArrSymbol.Address
	destFromVal := destArrSymbol.From
	destFromValSymbol, err := t.getSymbol(fmt.Sprintf("%d", destFromVal))
	if err != nil {
		return fmt.Errorf("failed to get symbol for dest array lower bound: %v", err)
	}
	destIdxSym, err := t.getSymbol(destIdx)
	if err != nil {
		return fmt.Errorf("failed to get symbol for dest index %s: %v", destIdx, err)
	}
	destPointerCell := t.pointerCell // e.g., 69
	_ = t.pointerCell + 1            // e.g., 70

	// Parse source array reference
	srcBase, srcIdx, err := parseArrayRef(src)
	if err != nil {
		return fmt.Errorf("failed to parse array ref for src %s: %v", src, err)
	}
	srcArrSymbol, err := t.getSymbol(srcBase)
	if err != nil {
		return fmt.Errorf("failed to get symbol for src array base %s: %v", srcBase, err)
	}
	srcBaseAddr := srcArrSymbol.Address
	srcFromVal := srcArrSymbol.From
	srcFromValSymbol, err := t.getSymbol(fmt.Sprintf("%d", srcFromVal))
	if err != nil {
		return fmt.Errorf("failed to get symbol for src array lower bound: %v", err)
	}
	srcIdxSym, err := t.getSymbol(srcIdx)
	if err != nil {
		return fmt.Errorf("failed to get symbol for src index %s: %v", srcIdx, err)
	}
	srcPointerCell := t.pointerCell + 2 // e.g., 71
	srcTempValue := t.pointerCell + 3   // e.g., 72

	// Step 1: Compute the address of y[m]
	t.emit(code.Instruction{
		Op:         code.LOAD,
		Label:      label,
		HasOperand: true,
		Operand:    srcIdxSym.Address,
		Comment:    fmt.Sprintf("load src index variable %s", srcIdx),
	})
	if srcFromVal != 0 {
		t.emit(code.Instruction{
			Op:         code.SUB,
			HasOperand: true,
			Operand:    srcFromValSymbol.Address,
			Comment:    fmt.Sprintf("subtract src array.from (%d) for %s", srcFromVal, srcBase),
		})
	}
	t.emit(code.Instruction{
		Op:         code.ADD,
		HasOperand: true,
		Operand:    srcBaseAddr,
		Comment:    fmt.Sprintf("add src base address (%d) for %s", srcBaseAddr, srcBase),
	})
	t.emit(code.Instruction{
		Op:         code.STORE,
		HasOperand: true,
		Operand:    srcPointerCell,
		Comment:    "store computed src address into srcPointerCell",
	})

	// Step 2: Load value from y[m] via LOADI
	t.emit(code.Instruction{
		Op:         code.LOADI,
		HasOperand: true,
		Operand:    srcPointerCell,
		Comment:    fmt.Sprintf("load %s[%s] into ACC", srcBase, srcIdx),
	})

	// Step 3: Store ACC into srcTempValue
	t.emit(code.Instruction{
		Op:         code.STORE,
		HasOperand: true,
		Operand:    srcTempValue,
		Comment:    "store ACC into srcTempValue",
	})

	// Step 4: Compute the address of x[n]
	t.emit(code.Instruction{
		Op:         code.LOAD,
		HasOperand: true,
		Operand:    destIdxSym.Address,
		Comment:    fmt.Sprintf("load dest index variable %s", destIdx),
	})
	if destFromVal != 0 {
		t.emit(code.Instruction{
			Op:         code.SUB,
			HasOperand: true,
			Operand:    destFromValSymbol.Address,
			Comment:    fmt.Sprintf("subtract dest array.from (%d) for %s", destFromVal, destBase),
		})
	}
	t.emit(code.Instruction{
		Op:         code.ADD,
		HasOperand: true,
		Operand:    destBaseAddr,
		Comment:    fmt.Sprintf("add dest base address (%d) for %s", destBaseAddr, destBase),
	})
	t.emit(code.Instruction{
		Op:         code.STORE,
		HasOperand: true,
		Operand:    destPointerCell,
		Comment:    "store computed dest address into destPointerCell",
	})

	// Step 5: Load preserved value from srcTempValue into ACC
	t.emit(code.Instruction{
		Op:         code.LOAD,
		HasOperand: true,
		Operand:    srcTempValue,
		Comment:    "load preserved value from srcTempValue into ACC",
	})

	// Step 6: Store ACC into x[n] via STOREI
	t.emit(code.Instruction{
		Op:         code.STOREI,
		HasOperand: true,
		Operand:    destPointerCell,
		Comment:    fmt.Sprintf("store ACC into %s[%s]", destBase, destIdx),
	})

	return nil
}
func (t *Translator) handleAddSub(op tac.Op, dest, left, right, label string) error {
	isLeftArray := isArrayRef(left)
	isRightArray := isArrayRef(right)

	// Handle Left Operand
	if isLeftArray {
		err := t.loadOperandIndirect(left)
		if err != nil {
			return fmt.Errorf("failed to load left operand %s: %v", left, err)
		}
	} else {
		err := t.loadOperand(left, label)
		if err != nil {
			return fmt.Errorf("failed to load left operand %s: %v", left, err)
		}
	}

	// Handle Right Operand
	if isRightArray {
		// Load array element into ACC
		err := t.loadOperandIndirect(right)
		if err != nil {
			return fmt.Errorf("failed to load right operand %s: %v", right, err)
		}
		// Store it temporarily
		tempRightAddr := t.allocateTemp()
		t.emit(code.Instruction{
			Op:         code.STORE,
			Operand:    tempRightAddr,
			Comment:    "store right operand value into temp",
			HasOperand: true,
		})
		// Load it back for the operation
		t.emit(code.Instruction{
			Op:         code.LOAD,
			Operand:    tempRightAddr,
			Comment:    "load temp right operand into ACC",
			HasOperand: true,
		})
		// Perform the operation
		if op == tac.OpAdd {
			t.emit(code.Instruction{
				Op:         code.ADD,
				Operand:    tempRightAddr,
				Comment:    fmt.Sprintf("add %s", right),
				HasOperand: true,
			})
		} else {
			t.emit(code.Instruction{
				Op:         code.SUB,
				Operand:    tempRightAddr,
				Comment:    fmt.Sprintf("subtract %s", right),
				HasOperand: true,
			})
		}
	} else {
		// Right operand is a variable; perform ADD or SUB directly
		rightAddr, err := t.getAddr(right)
		if err != nil {
			return fmt.Errorf("failed to get address of right operand %s: %v", right, err)
		}
		var opCode code.Opcode
		var comment string
		if op == tac.OpAdd {
			opCode = code.ADD
			comment = fmt.Sprintf("add %s", right)
		} else {
			opCode = code.SUB
			comment = fmt.Sprintf("subtract %s", right)
		}
		t.emit(code.Instruction{
			Op:         opCode,
			Operand:    rightAddr,
			Comment:    comment,
			HasOperand: true,
		})
	}

	// Step 4: Store the result into destination
	destAddr, err := t.getAddr(dest)
	if err != nil {
		return fmt.Errorf("failed to get address of destination %s: %v", dest, err)
	}
	t.emit(code.Instruction{
		Op:         code.STORE,
		Operand:    destAddr,
		Comment:    fmt.Sprintf("store ACC into %s", dest),
		HasOperand: true,
	})

	return nil
}
func (t *Translator) handleArrayLoad(dest, array, index, label string) error {
	arrSymbol, err := t.getSymbol(array)
	if err != nil {
		return fmt.Errorf("array %s not found: %v", array, err)
	}

	idxAddr, err := t.getAddr(index)
	if err != nil {
		return fmt.Errorf("index %s not found: %v", index, err)
	}

	// Calculate address: array_base + (index - from)
	t.emit(code.Instruction{
		Op:         code.LOAD,
		HasOperand: true,
		Operand:    idxAddr,
		Label:      label,
	})

	if arrSymbol.From != 0 {
		fromAddr, err := t.getAddr(strconv.Itoa(arrSymbol.From))
		if err != nil {
			return fmt.Errorf("from value not found: %v", err)
		}
		t.emit(code.Instruction{
			Op:         code.SUB,
			HasOperand: true,
			Operand:    fromAddr,
		})
	}

	t.emit(code.Instruction{
		Op:         code.ADD,
		HasOperand: true,
		Operand:    arrSymbol.Address,
	})

	ptrCell := t.pointerCell
	t.emit(code.Instruction{
		Op:         code.STORE,
		HasOperand: true,
		Operand:    ptrCell,
	})

	// Load value from array
	t.emit(code.Instruction{
		Op:         code.LOADI,
		HasOperand: true,
		Operand:    ptrCell,
	})

	destAddr, err := t.getAddr(dest)
	if err != nil {
		return fmt.Errorf("destination %s not found: %v", dest, err)
	}

	t.emit(code.Instruction{
		Op:         code.STORE,
		HasOperand: true,
		Operand:    destAddr,
	})

	return nil
}

// loadOperand loads the operand into ACC.
// If it's an array reference, it computes the address and uses LOADI.
// If it's a variable, it uses LOAD.
func (t *Translator) loadOperand(operand string, label string) error {
	if isArrayRef(operand) {
		// Operand is an array element; compute address and use LOADI
		return t.loadOperandIndirect(operand)
	} else {
		// Operand is a variable; use LOAD
		addr, err := t.getAddr(operand)
		if err != nil {
			return fmt.Errorf("failed to get address for operand %s: %v", operand, err)
		}
		ins := code.Instruction{
			Op:         code.LOAD,
			Operand:    addr,
			Comment:    fmt.Sprintf("load %s into ACC", operand),
			HasOperand: true,
		}
		if label != "" {
			ins.Label = label
		}
		t.emit(ins)
		return nil
	}
}

// loadOperandIndirect computes the address of the array element and uses LOADI to load its value into ACC.
func (t *Translator) loadOperandIndirect(arrayRef string) error {
	base, index, err := parseArrayRef(arrayRef)
	if err != nil {
		return fmt.Errorf("invalid array reference %s: %v", arrayRef, err)
	}

	arrSymbol, err := t.getSymbol(base)
	if err != nil {
		return fmt.Errorf("failed to get symbol for array base %s: %v", base, err)
	}

	// Compute the address: baseAddr + (index - fromVal)
	indexAddr, err := t.getAddr(index)
	if err != nil {
		return fmt.Errorf("failed to get address for index %s: %v", index, err)
	}

	// Load the index value into ACC
	t.emit(code.Instruction{
		Op:         code.LOAD,
		Operand:    indexAddr,
		Comment:    fmt.Sprintf("load index %s for array %s", index, base),
		HasOperand: true,
	})

	// Subtract the lower bound if necessary
	if arrSymbol.From != 0 {
		fromValStr := fmt.Sprintf("%d", arrSymbol.From)
		fromValAddr, err := t.getAddr(fromValStr)
		if err != nil {
			return fmt.Errorf("failed to get address for array lower bound %s: %v", fromValStr, err)
		}
		t.emit(code.Instruction{
			Op:         code.SUB,
			Operand:    fromValAddr,
			Comment:    fmt.Sprintf("subtract lower bound %d for array %s", arrSymbol.From, base),
			HasOperand: true,
		})
	}

	// Add the base address
	t.emit(code.Instruction{
		Op:         code.ADD,
		Operand:    arrSymbol.Address,
		Comment:    fmt.Sprintf("add base address %d for array %s", arrSymbol.Address, base),
		HasOperand: true,
	})

	// Store the computed address into a temporary cell
	tempAddr := t.allocateTemp()
	t.emit(code.Instruction{
		Op:         code.STORE,
		Operand:    tempAddr,
		Comment:    "store computed address into temp cell",
		HasOperand: true,
	})

	// Use LOADI to load the value from the computed address
	t.emit(code.Instruction{
		Op:         code.LOADI,
		Operand:    tempAddr,
		Comment:    fmt.Sprintf("load %s[%s] into ACC", base, index),
		HasOperand: true,
	})

	return nil
}

// storeOperandIndirect computes the address of the array element and uses STOREI to store ACC's value.
func (t *Translator) storeOperandIndirect(arrayRef string) error {
	base, index, err := parseArrayRef(arrayRef)
	if err != nil {
		return fmt.Errorf("invalid array reference %s: %v", arrayRef, err)
	}

	arrSymbol, err := t.getSymbol(base)
	if err != nil {
		return fmt.Errorf("failed to get symbol for array base %s: %v", base, err)
	}

	// Compute the address: baseAddr + (index - fromVal)
	indexAddr, err := t.getAddr(index)
	if err != nil {
		return fmt.Errorf("failed to get address for index %s: %v", index, err)
	}

	// Load the index value into ACC
	t.emit(code.Instruction{
		Op:         code.LOAD,
		Operand:    indexAddr,
		Comment:    fmt.Sprintf("load index %s for array %s", index, base),
		HasOperand: true,
	})

	// Subtract the lower bound if necessary
	if arrSymbol.From != 0 {
		fromValStr := fmt.Sprintf("%d", arrSymbol.From)
		fromValAddr, err := t.getAddr(fromValStr)
		if err != nil {
			return fmt.Errorf("failed to get address for array lower bound %s: %v", fromValStr, err)
		}
		t.emit(code.Instruction{
			Op:         code.SUB,
			Operand:    fromValAddr,
			Comment:    fmt.Sprintf("subtract lower bound %d for array %s", arrSymbol.From, base),
			HasOperand: true,
		})
	}

	// Add the base address
	t.emit(code.Instruction{
		Op:         code.ADD,
		Operand:    arrSymbol.Address,
		Comment:    fmt.Sprintf("add base address %d for array %s", arrSymbol.Address, base),
		HasOperand: true,
	})

	// Store the computed address into a temporary cell
	tempAddr := t.allocateTemp()
	t.emit(code.Instruction{
		Op:         code.STORE,
		Operand:    tempAddr,
		Comment:    "store computed address into temp cell for STOREI",
		HasOperand: true,
	})

	// Use STOREI to store ACC's value into the array element
	t.emit(code.Instruction{
		Op:         code.STOREI,
		Operand:    tempAddr,
		Comment:    fmt.Sprintf("store ACC into %s[%s]", base, index),
		HasOperand: true,
	})

	return nil
}

// allocateTemp allocates a new temporary cell and returns its address.
func (t *Translator) allocateTemp() int {
	t.tempCounter++
	// Optionally, add to symbol table if needed
	tempName := fmt.Sprintf("temp%d", t.tempCounter)
	t.St.Declare(tempName, t.currentFunctionName, symboltable.Symbol{Name: tempName, Kind: symboltable.TEMP})
	addr, _ := t.getAddr(tempName)
	return addr
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
	t.emit(code.Instruction{Op: code.LOAD, Comment: dest + " := " + left + " % " + right, HasOperand: true, Label: label, Operand: leftAddr})
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
	ins.Comment = "param " + param
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
		ins.Comment = "call " + funcName + " " + paramCount
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
	t.emit(code.Instruction{Op: code.RTRN, Comment: "ret", Label: label, HasOperand: true, Operand: returnAddr.Address})
	return nil
}

func (t *Translator) handleHalt(label string) {
	t.emit(code.Instruction{Label: label, Comment: "halt", Op: code.HALT})
}

func (t *Translator) handleGoto(labelName, label string) {
	t.emit(code.Instruction{Op: code.JUMP, Comment: "goto " + labelName, Label: label, HasOperand: true, Destination: labelName})
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
		jumpIns2 = &code.Instruction{Op: code.JNEG, HasOperand: true, Destination: labelName}
	}
	t.emit(code.Instruction{Op: code.LOAD, Comment: string(op) + " " + left + " " + right, HasOperand: true, Label: label, Operand: leftAddr})
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

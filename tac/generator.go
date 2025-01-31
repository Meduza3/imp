package tac

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Meduza3/imp/ast"
	"github.com/Meduza3/imp/symboltable"
)

type Generator struct {
	SymbolTable  *symboltable.SymbolTable
	Instructions []Instruction
	Errors       []string

	labelCount int
	tempCount  int

	currentMemoryOffset int
	currentProc         string
}

func NewGenerator() *Generator {
	return &Generator{
		SymbolTable:         symboltable.New(),
		Errors:              make([]string, 0),
		currentMemoryOffset: 1,
	}
}

func (g *Generator) GetSymbolTable() *symboltable.SymbolTable {
	return g.SymbolTable
}

func (g *Generator) newLabel() string {
	g.labelCount++
	return fmt.Sprintf("L%d", g.labelCount)
}

func (g *Generator) newTemp() string {
	g.tempCount++
	name := fmt.Sprintf("t%d", g.tempCount)
	g.SymbolTable.Declare(name, g.currentProc, symboltable.Symbol{
		Name: name,
		Kind: symboltable.TEMP,
	})
	return fmt.Sprintf("t%d", g.tempCount)
}

func opFromToken(tk ast.MathExpression) Op {
	switch tk.Operator.Literal {
	case "+":
		return OpAdd
	case "-":
		return OpSub
	case "*":
		return OpMul
	case "/":
		return OpDiv
	case "%":
		return OpMod
	default:
		// If it's not recognized, default to "assign" or throw error
		return OpAssign
	}
}

func (g *Generator) Generate(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		g.SymbolTable.Declare("1", "main", symboltable.Symbol{Name: "1", Kind: symboltable.CONSTANT})
		g.emit(Instruction{Op: OpGoto, Destination: "main"})
		for _, procedure := range node.Procedures {
			if procedure != nil {
				_ = g.Generate(procedure)
			}
		}
		if node.Main != nil {
			_ = g.Generate(node.Main)
		}
		g.emit(Instruction{
			Op: OpHalt,
		})

	case *ast.Procedure:

		oldProc := g.currentProc
		g.currentProc = node.ProcHead.Name.Value // e.g. "de"
		g.SymbolTable.Declare(g.currentProc, "xxFunctionsxx", symboltable.Symbol{Name: g.currentProc, Kind: symboltable.PROCEDURE, ArgCount: len(node.ProcHead.ArgsDecl)})
		g.SymbolTable.Declare(g.currentProc+"_return", "xxFunctionsxx", symboltable.Symbol{Name: g.currentProc + "_return", Kind: symboltable.RETURNADDR})
		g.emit(Instruction{Label: node.ProcHead.Name.Value})
		for _, decl := range node.ProcHead.ArgsDecl {
			err := g.DeclareArgProcedure(decl, g.currentProc)
			if err != nil {
				g.Errors = append(g.Errors, err.Error())
			}
		}
		for _, decl := range node.Declarations {
			err := g.DeclareProcedure(decl, g.currentProc)
			if err != nil {
				g.Errors = append(g.Errors, err.Error())
			}
		}
		for _, comm := range node.Commands {
			err := g.Generate(comm)
			if err != nil {
				g.Errors = append(g.Errors, err.Error())
			}
		}
		g.emit(Instruction{Op: OpRet})
		g.currentProc = oldProc
	case *ast.Main:
		oldProc := g.currentProc
		g.currentProc = "main"
		g.emit(Instruction{Label: "main"})
		for _, decl := range node.Declarations {
			err := g.DeclareMain(decl)
			if err != nil {
				g.Errors = append(g.Errors, err.Error())
			}
		}

		for _, comm := range node.Commands {
			err := g.Generate(comm)
			if err != nil {
				g.Errors = append(g.Errors, err.Error())
			}
		}
		g.currentProc = oldProc
	case *ast.AssignCommand:
		// 1. Generate a place (temp or variable) for the right-hand side
		place, err := g.generateMathExpression(&node.MathExpression)
		if err != nil {
			g.Errors = append(g.Errors, err.Error())
			return nil // or return err, up to you
		}

		// 2. Emit a final assignment: identifier = place
		qualifiedDest := g.qualifyVar(node.Identifier.String())
		g.emit(Instruction{
			Op:          OpAssign,
			Destination: qualifiedDest,
			Arg1:        place,
		})
	case *ast.WriteCommand:
		val := g.qualifyVarOrNumber(node.Value.String())
		if isNumber(val) {
			g.SymbolTable.Declare(val, "main", symboltable.Symbol{Name: val, Kind: symboltable.CONSTANT})
		}
		g.emit(Instruction{Op: OpWrite, Arg1: val})
	case *ast.ReadCommand:
		ident := g.qualifyVar(node.Identifier.String())
		g.emit(Instruction{Op: OpRead, Arg1: ident})
	case *ast.WhileCommand:
		labelStart := g.newLabel() // e.g. "L1"
		labelBody := g.newLabel()  // e.g. "L2"
		labelEnd := g.newLabel()   // e.g. "L3"

		// 2. Emit labelStart at the top of the loop
		g.emit(Instruction{Label: labelStart})
		err := g.generateCondition(node.Condition, labelBody, labelEnd)
		if err != nil {
			g.Errors = append(g.Errors, err.Error())
			return nil
		}
		g.emit(Instruction{Label: labelBody})
		for _, cmd := range node.Commands {
			if err := g.Generate(cmd); err != nil {
				g.Errors = append(g.Errors, err.Error())
			}
		}
		g.emit(Instruction{
			Op:          OpGoto,
			Destination: labelStart,
		})
		g.emit(Instruction{Label: labelEnd})

	case *ast.ForCommand:
		iteratorName := node.Iterator.Value
		g.SymbolTable.Declare(iteratorName, g.currentProc, symboltable.Symbol{Name: iteratorName, Kind: symboltable.DECLARATION})

		startVal := node.From.String()
		endVal := node.To.String()
		// TODO: this may be a table to check for being a table
		_, err := strconv.Atoi(startVal)
		if err == nil {
			g.SymbolTable.Declare(startVal, "main", symboltable.Symbol{Name: startVal, Kind: symboltable.CONSTANT})
		} else {
			g.SymbolTable.Declare(startVal, g.currentProc, symboltable.Symbol{Name: startVal, Kind: symboltable.DECLARATION})
		}
		_, err = strconv.Atoi(endVal)
		if err == nil {
			g.SymbolTable.Declare(endVal, "main", symboltable.Symbol{Name: endVal, Kind: symboltable.CONSTANT})
		} else {
			g.SymbolTable.Declare(endVal, g.currentProc, symboltable.Symbol{Name: endVal, Kind: symboltable.DECLARATION})
		}

		labelTest := g.newLabel() // e.g. "L1"
		labelBody := g.newLabel() // e.g. "L2"
		labelEnd := g.newLabel()  // e.g. "L3"

		g.emit(Instruction{
			Op:          OpAssign, // "="
			Destination: iteratorName,
			Arg1:        startVal,
		})
		g.emit(Instruction{
			Op:          OpGoto,
			Destination: labelTest,
		})
		g.emit(Instruction{Label: labelTest})
		if !node.IsDownTo {
			// ascending
			g.emit(Instruction{
				Op:          OpIfLE, // "if<="
				Arg1:        iteratorName,
				Arg2:        endVal,
				Destination: labelBody,
			})
		} else {
			// descending
			g.emit(Instruction{
				Op:          OpIfGE, // "if>="
				Arg1:        iteratorName,
				Arg2:        endVal,
				Destination: labelBody,
			})
		}
		g.emit(Instruction{
			Op:          OpGoto,
			Destination: labelEnd,
		})
		g.emit(Instruction{Label: labelBody})
		for _, cmd := range node.Commands {
			if err := g.Generate(cmd); err != nil {
				g.Errors = append(g.Errors, err.Error())
			}
		}
		if !node.IsDownTo {
			// i = i + 1
			tmp := g.newTemp()
			g.emit(Instruction{
				Op:          OpAdd, // "tX = i + 1"
				Destination: tmp,
				Arg1:        iteratorName,
				Arg2:        "1",
			})
			// i = tX
			g.emit(Instruction{
				Op:          OpAssign,
				Destination: iteratorName,
				Arg1:        tmp,
			})
		} else {
			// i = i - 1
			tmp := g.newTemp()
			g.emit(Instruction{
				Op:          OpSub, // "tX = i - 1"
				Destination: tmp,
				Arg1:        iteratorName,
				Arg2:        "1",
			})
			g.emit(Instruction{
				Op:          OpAssign,
				Destination: iteratorName,
				Arg1:        tmp,
			})
		}

		// 9. Jump back to labelTest
		g.emit(Instruction{
			Op:          OpGoto,
			Destination: labelTest,
		})
		g.emit(Instruction{Label: labelEnd})
	case *ast.ProcCallCommand:
		for _, arg := range node.Args {
			argName := arg.String()
			symbol, err := g.SymbolTable.Lookup(argName, g.currentProc)
			if err == nil {
				g.emit(Instruction{
					Op:   OpParam,
					Arg1: g.qualifyVarOrNumber(argName),
					Arg2: fmt.Sprintf("%d", symbol.ArgCount),
				})
			} else {
				fmt.Printf("failed calling %s: %v\n", argName, err)
				return fmt.Errorf("failed calling %s: %v", argName, err)
			}
		}

		procName := node.Name.String()
		numArgs := len(node.Args)
		_, err := g.SymbolTable.Lookup(procName, "xxFunctionsxx")
		if err == nil {
			g.emit(Instruction{
				Op:   OpCall,
				Arg1: procName, // the procedure label/name
				Arg2: fmt.Sprintf("%d", numArgs),
			})
		} else {
			fmt.Printf("failed calling %s: %v\n", procName, err)

			return fmt.Errorf("failed calling %s: %v", procName, err)

		}

	case *ast.RepeatCommand:
		labelStart := g.newLabel()
		labelEnd := g.newLabel()
		g.emit(Instruction{Label: labelStart})
		for _, cmd := range node.Commands {
			if err := g.Generate(cmd); err != nil {
				g.Errors = append(g.Errors, err.Error())
			}
		}

		// 4. Generate code to test the condition
		//    If the condition is true, jump to labelEnd; otherwise jump to labelStart
		err := g.generateCondition(node.Condition, labelEnd, labelStart)
		if err != nil {
			g.Errors = append(g.Errors, err.Error())
			return nil
		}

		// 5. Emit labelEnd (the exit point)
		g.emit(Instruction{Label: labelEnd})

	case *ast.IfCommand:
		labelThen := g.newLabel() // e.g. "L1"
		labelEnd := g.newLabel()  // e.g. "L2"
		var labelElse string
		hasElse := (len(node.ElseCommands) > 0)
		if hasElse {
			labelElse = g.newLabel() // e.g. "L3"
		} else {
			labelElse = labelEnd // If no else, "else" is effectively the end
		}

		// 1. Evaluate condition, produce branch instructions
		//    "if condition => labelThen"
		//    "goto labelElse"
		err := g.generateCondition(node.Condition, labelThen, labelElse)
		if err != nil {
			g.Errors = append(g.Errors, err.Error())
			return nil
		}

		g.emit(Instruction{Label: labelThen})
		for _, cmd := range node.ThenCommands {
			if err := g.Generate(cmd); err != nil {
				g.Errors = append(g.Errors, err.Error())
			}
		}

		if hasElse {
			g.emit(Instruction{
				Op:          OpGoto,
				Destination: labelEnd,
			})
			g.emit(Instruction{
				Label: labelElse,
			})
			for _, cmd := range node.ElseCommands {
				if err := g.Generate(cmd); err != nil {
					g.Errors = append(g.Errors, err.Error())
				}
			}
		}

		g.emit(Instruction{Label: labelEnd})

	default:
		return nil
	}
	return nil
}

// This helper qualifies either a plain identifier "a" => "currentProc.a",
// or an array "arr[i]" => "currentProc.arr[i]" if your language uses that approach.
func (g *Generator) qualifyVar(plainName string) string {
	// If plainName is something like "x", we do "proc.x"
	// If it already has a dot, or is a temp or number, skip.
	// But since your code calls this only for actual variables, we can do:
	if plainName == "" {
		return "" // safety
	}
	if strings.Contains(plainName, "[") {
		idx := extractIndex(plainName)
		_, err := strconv.Atoi(idx)
		if err == nil {
			g.SymbolTable.Declare(idx, "main", symboltable.Symbol{Kind: symboltable.CONSTANT, Name: idx})
		}
	}
	// If it’s already got ".", or is "t1", just return it as-is
	if hasDot(plainName) || isTemp(plainName) || isNumber(plainName) {
		return plainName
	}
	// Otherwise prepend currentProc + "."
	return g.currentProc + "." + plainName
}

func extractIndex(input string) string {
	start := strings.Index(input, "[")
	end := strings.LastIndex(input, "]")

	if start == -1 || end == -1 || start+1 >= end {
		return "" // Return empty if brackets are missing or invalid
	}

	return input[start+1 : end] // Extract content between [ and ]
}

// If it’s a number, we keep it as-is. Otherwise qualify it.
func (g *Generator) qualifyVarOrNumber(s string) string {
	if isNumber(s) || isTemp(s) {
		return s
	}
	return g.qualifyVar(s)
}

func isNumber(str string) bool {
	_, err := strconv.Atoi(str)
	return (err == nil)
}

func isTemp(str string) bool {
	return len(str) > 0 && str[0] == 't'
}

func hasDot(str string) bool {
	return len(str) > 0 && func() bool {
		for i := range str {
			if str[i] == '.' {
				return true
			}
		}
		return false
	}()
}
func MergeLabelOnlyInstructions(inss []Instruction) []Instruction {
	var result []Instruction
	i := 0
	for i < len(inss) {
		ins := inss[i]

		// Check if it's really just a "label" line,
		// i.e. it has a label and absolutely no other data.
		isLabelOnly := (ins.Label != "" &&
			ins.Op == "" &&
			ins.Arg1 == "" &&
			ins.Arg2 == "" &&
			ins.Destination == "")

		if isLabelOnly {
			// Merge this label with the *next* instruction if one exists
			if i+1 < len(inss) {
				next := inss[i+1]
				if next.Label == "" {
					next.Label = ins.Label
				} else {
					// If the next instruction also has a label, you could decide
					// how to handle multiple labels. For simplicity, join them:
					next.Label = ins.Label + " " + next.Label
				}
				inss[i+1] = next
			} else {
				// If there's no “next” instruction, keep this label as a “no‐op.”
				result = append(result, ins)
			}
			i++
		} else {
			// Normal instruction or label+operation/args
			result = append(result, ins)
			i++
		}
	}
	return result
}

func (g *Generator) generateCondition(cond ast.Condition, labelTrue, labelFalse string) error {
	// 1. Generate code for left and right
	left, err := g.generateValue(cond.Left)
	if err != nil {
		return err
	}
	right, err := g.generateValue(cond.Right)
	if err != nil {
		return err
	}

	// 2. Map the condition's operator token to an OpIfXX
	op, err := mapConditionOp(cond.Operator.Literal)
	if err != nil {
		return err
	}

	// 3. Emit the branch:
	//    if< left, right => labelTrue
	//    goto labelFalse
	g.emit(Instruction{
		Op:          op,
		Arg1:        left,
		Arg2:        right,
		Destination: labelTrue,
	})

	// Unconditional jump to labelFalse
	g.emit(Instruction{
		Op:          OpGoto,
		Destination: labelFalse,
	})
	return nil
}

// Helper that maps e.g. "=" => OpIfEQ, "<" => OpIfLT, ...
func mapConditionOp(op string) (Op, error) {
	switch op {
	case "=":
		return OpIfEQ, nil
	case "!=":
		return OpIfNE, nil
	case "<":
		return OpIfLT, nil
	case "<=":
		return OpIfLE, nil
	case ">":
		return OpIfGT, nil
	case ">=":
		return OpIfGE, nil
	}
	return "", fmt.Errorf("unknown condition operator %q", op)
}

// generateValue returns a string “place” for the given Value.
// If it’s just a number literal or identifier, we can return the string directly.
// If your language has array-index computations, you’d handle them here.
// generateValue returns a string “place” for the given Value.
// If it’s an array reference, it emits instructions to compute the address and load the value.
func (g *Generator) generateValue(v ast.Value) (string, error) {
	switch val := v.(type) {
	case *ast.NumberLiteral:
		numStr := val.String()
		// Declare the number as a constant
		g.SymbolTable.Declare(numStr, "main", symboltable.Symbol{
			Name: numStr,
			Kind: symboltable.CONSTANT,
		})
		// If negative, also declare its absolute value
		if strings.HasPrefix(numStr, "-") {
			posVal := numStr[1:]
			g.SymbolTable.Declare(posVal, "main", symboltable.Symbol{
				Name: posVal,
				Kind: symboltable.CONSTANT,
			})
		}
		return numStr, nil

	case *ast.Identifier:

		return g.qualifyVar(val.String()), nil

	default:
		return "", fmt.Errorf("unhandled Value type %T", v)
	}
}

// generateMathExpression returns the place holding the result of the expression.
// If expression.Right is nil, there's no operator, so just return Left's place.
// Otherwise, emit an instruction to combine Left and Right.
func (g *Generator) generateMathExpression(me *ast.MathExpression) (string, error) {
	leftPlace, err := g.generateValue(me.Left)
	if err != nil {
		return "", err
	}

	// If there's no operator, it's just a single operand
	if me.Right == nil {
		return leftPlace, nil
	}

	// We have an operator and a right operand
	rightPlace, err := g.generateValue(me.Right)
	if err != nil {
		return "", err
	}

	// Create a fresh temporary for the operation result
	tmp := g.newTemp()

	// Map the token to our Op enum
	op := opFromToken(*me)

	// Emit the arithmetic instruction:
	//     tmp = leftPlace op rightPlace
	g.emit(Instruction{
		Op:          op,
		Destination: tmp,
		Arg1:        leftPlace,
		Arg2:        rightPlace,
	})

	return tmp, nil
}

func (g *Generator) emit(ins Instruction) {
	g.Instructions = append(g.Instructions, ins)
}

func (g *Generator) DeclareArgProcedure(decl ast.ArgDecl, procName string) error {
	var argCount = 0
	for _, symbol := range g.GetSymbolTable().Table[procName] {
		if symbol.Kind == symboltable.ARGUMENT {
			argCount++
		}
	}
	name := decl.Name.Value
	isTable := decl.IsTable
	symbol := symboltable.Symbol{
		Name:          name,
		Kind:          symboltable.ARGUMENT,
		IsTable:       isTable,
		Address:       g.currentMemoryOffset,
		ArgumentIndex: argCount + 1,
	}
	err := g.SymbolTable.Declare(name, procName, symbol)
	if err != nil {
		return fmt.Errorf("failed to declare argument %v in procedure %s: %v", decl, procName, err)
	}
	g.currentMemoryOffset++
	return nil
}

func (g *Generator) DeclareProcedure(decl ast.Declaration, procName string) error {
	name := decl.Pidentifier.Value
	symbol := symboltable.Symbol{
		Name:    name,
		Kind:    symboltable.DECLARATION,
		Address: g.currentMemoryOffset,
	}
	got, _ := g.SymbolTable.Lookup(name, procName)
	if got == nil {
		err := g.SymbolTable.Declare(name, procName, symbol)
		if err != nil {
			return fmt.Errorf("failed to declare %v in procedure %s: %v", decl, procName, err)
		}
	}

	g.currentMemoryOffset++
	return nil
}

func (g *Generator) DeclareMain(decl ast.Declaration) error {
	name := decl.Pidentifier.Value
	var symbol symboltable.Symbol
	var nextMemory int
	if decl.IsTable {
		from, err := strconv.Atoi(decl.From.Value)
		g.SymbolTable.Declare(decl.From.Value, "main", symboltable.Symbol{Name: decl.From.Value, Kind: symboltable.CONSTANT})
		if err != nil {
			return fmt.Errorf("failed parsing from value in main declaration %v. value: %s", decl, decl.From.Value)
		}
		to, err := strconv.Atoi(decl.To.Value)
		g.SymbolTable.Declare(decl.To.Value, "main", symboltable.Symbol{Name: decl.To.Value, Kind: symboltable.CONSTANT})
		if err != nil {
			return fmt.Errorf("failed parsing to value in main declaration %v. value: %s", decl, decl.To.Value)
		}
		symbol = symboltable.Symbol{
			Name:    name,
			Kind:    symboltable.DECLARATION,
			IsTable: true,
			From:    from,
			To:      to,
			Size:    to - from + 1,
			Address: g.currentMemoryOffset,
		}
		nextMemory = g.currentMemoryOffset + symbol.Size
	} else {
		symbol = symboltable.Symbol{
			Name:    name,
			Kind:    symboltable.DECLARATION,
			Address: g.currentMemoryOffset,
		}
		nextMemory = g.currentMemoryOffset + 1
	}
	err := g.SymbolTable.Declare(name, "main", symbol)
	if err != nil {
		return fmt.Errorf("failed to declare in main: %v", err)
	}
	g.currentMemoryOffset = nextMemory
	return nil
}

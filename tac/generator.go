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

	currentProc string
}

func NewGenerator() *Generator {
	return &Generator{
		SymbolTable: symboltable.New(),
		Errors:      make([]string, 0),
	}
}

func (g *Generator) GetSymbolTable() *symboltable.SymbolTable {
	return g.SymbolTable
}

func (g *Generator) newLabel() string {
	g.labelCount++
	return fmt.Sprintf("L%d", g.labelCount)
}

func (g *Generator) newTemp() *symboltable.Symbol {
	g.tempCount++
	name := fmt.Sprintf("t%d", g.tempCount)

	sym, _ := g.SymbolTable.Declare(name, g.currentProc, symboltable.Symbol{
		Name: name,
		Kind: symboltable.TEMP,
	})
	return sym
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
		g.SymbolTable.Declare("built_in_left", "main", symboltable.Symbol{Name: "built_in_left", Kind: symboltable.DECLARATION})
		g.SymbolTable.Declare("built_in_right", "main", symboltable.Symbol{Name: "built_in_right", Kind: symboltable.DECLARATION})
		g.SymbolTable.Declare("built_in_result", "main", symboltable.Symbol{Name: "built_in_result", Kind: symboltable.DECLARATION})
		g.emit(Instruction{Op: OpGoto, JumpTo: "main"})
		for _, procedure := range node.Procedures {
			if procedure != nil {
				err := g.Generate(procedure)
				if err != nil {
					g.Errors = append(g.Errors, fmt.Sprintf("failed to generate procedure: %v", err))
				}
			}
		}
		if node.Main != nil {
			err := g.Generate(node.Main)
			if err != nil {
				g.Errors = append(g.Errors, fmt.Sprintf("failed to generate main: %v", err))
			}
		}
		g.emit(Instruction{
			Op: OpHalt,
		})

	case *ast.Procedure:

		oldProc := g.currentProc
		g.currentProc = node.ProcHead.Name.Value // e.g. "de"
		funcSym, _ := g.SymbolTable.Declare(g.currentProc, "xxFunctionsxx", symboltable.Symbol{Name: g.currentProc, Kind: symboltable.PROCEDURE, ArgCount: len(node.ProcHead.ArgsDecl)})
		g.SymbolTable.Declare(g.currentProc+"_return", "xxFunctionsxx", symboltable.Symbol{Name: g.currentProc + "_return", Kind: symboltable.RETURNADDR})
		g.emit(Instruction{Labels: []string{node.ProcHead.Name.Value}})
		for _, decl := range node.ProcHead.ArgsDecl {
			sym, err := g.DeclareArgProcedure(decl, g.currentProc)
			if err != nil {
				fmt.Println(funcSym.ArgCount, funcSym.Arguments)
				g.Errors = append(g.Errors, err.Error())
			} else {
				funcSym.Arguments = append(funcSym.Arguments, sym)
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
		g.emit(Instruction{Labels: []string{"main"}})
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
			return nil // or return err, up to you
		}
		if place == nil {
			return fmt.Errorf("failed to generate RHS for assignment")
		}
		// 2. Emit a final assignment: identifier = place
		idSymbol, err := g.SymbolTable.Lookup(node.Identifier.Value, g.currentProc)
		if err != nil {
			return fmt.Errorf("failed to lookup for idSymbol: %v", err)
		}
		if isNumber(node.Identifier.Index) {
			g.SymbolTable.Declare(node.Identifier.Index, "main", symboltable.Symbol{Name: node.Identifier.Index, Kind: symboltable.CONSTANT})
		}
		if idSymbol.IsTable {
			g.emit(Instruction{
				Op:        OpAssign,
				Arg1:      idSymbol,
				Arg1Index: node.Identifier.Index,
				Arg2:      place,
			})
		} else {
			if idSymbol == nil {
				return fmt.Errorf("nil idSymbol")
			}
			g.emit(Instruction{
				Op:   OpAssign,
				Arg1: idSymbol,
				Arg2: place,
			})
		}

	case *ast.WriteCommand:
		val := node.Value

		var sym *symboltable.Symbol
		var err error
		if isNumber(val.String()) {
			sym, err = g.SymbolTable.Declare(val.String(), "main", symboltable.Symbol{Name: val.String(), Kind: symboltable.CONSTANT})
			if err != nil {
			}
			g.emit(Instruction{Op: OpWrite, Arg1: sym})

			return nil
		}
		switch value := val.(type) {
		case *ast.Identifier:
			sym, err = g.SymbolTable.Lookup(value.Value, g.currentProc)
			if err != nil {
				return fmt.Errorf("failed to generate Write for %v: %v", node, err)
			}
			if isNumber(value.Index) {
				_, err = g.SymbolTable.Declare(value.Index, "main", symboltable.Symbol{Name: value.Index, Kind: symboltable.CONSTANT})
				if err != nil {
				}
			}
			g.emit(Instruction{Op: OpWrite, Arg1: sym, Arg1Index: value.Index})
		}
	case *ast.ReadCommand:
		val := node.Identifier
		var sym *symboltable.Symbol
		if isNumber(val.String()) {
			sym, _ = g.SymbolTable.Declare(val.Value, "main", symboltable.Symbol{Name: val.Value, Kind: symboltable.CONSTANT})
			g.emit(Instruction{Op: OpRead, Arg1: sym})
			return nil
		}
		sym, _ = g.SymbolTable.Lookup(val.Value, g.currentProc)
		if isNumber(val.Index) {
			g.SymbolTable.Declare(val.Index, "main", symboltable.Symbol{Name: val.Index, Kind: symboltable.CONSTANT})
		}
		g.emit(Instruction{Op: OpRead, Arg1: sym, Arg1Index: val.Index})

	case *ast.WhileCommand:
		labelStart := g.newLabel() // e.g. "L1"
		labelBody := g.newLabel()  // e.g. "L2"
		labelEnd := g.newLabel()   // e.g. "L3"

		// 2. Emit labelStart at the top of the loop
		g.emit(Instruction{Labels: []string{labelStart}})

		err := g.generateCondition(node.Condition, labelBody, labelEnd)
		if err != nil {
			return err
		}
		g.emit(Instruction{Labels: []string{labelBody}})

		for _, cmd := range node.Commands {
			if err := g.Generate(cmd); err != nil {
				g.Errors = append(g.Errors, err.Error())
			}
		}

		g.emit(Instruction{
			Op:     OpGoto,
			JumpTo: labelStart,
		})
		g.emit(Instruction{Labels: []string{labelEnd}})

	case *ast.ForCommand:
		iteratorName := node.Iterator.Value
		iteratorSymbol, _ := g.SymbolTable.Declare(iteratorName, g.currentProc, symboltable.Symbol{Name: iteratorName, Kind: symboltable.DECLARATION})

		startVal := node.From.String()
		endVal := node.To.String()
		// TODO: this may be a table to check for being a table
		var startSymbol *symboltable.Symbol
		_, err := strconv.Atoi(startVal)
		if err == nil {
			startSymbol, _ = g.SymbolTable.Declare(startVal, "main", symboltable.Symbol{Name: startVal, Kind: symboltable.CONSTANT})
		} else {
			startSymbol, _ = g.SymbolTable.Declare(startVal, g.currentProc, symboltable.Symbol{Name: startVal, Kind: symboltable.DECLARATION})
		}
		if startSymbol == nil {
			return fmt.Errorf("nil startSymbol")
		}
		var endSymbol *symboltable.Symbol
		_, err = strconv.Atoi(endVal)
		if err == nil {
			endSymbol, _ = g.SymbolTable.Declare(endVal, "main", symboltable.Symbol{Name: endVal, Kind: symboltable.CONSTANT})
		} else {
			endSymbol, _ = g.SymbolTable.Declare(endVal, g.currentProc, symboltable.Symbol{Name: endVal, Kind: symboltable.DECLARATION})
		}

		oneSymbol, err := g.SymbolTable.Lookup("1", "main")
		if err != nil {
			return fmt.Errorf("failed to lookup symbol 1")
		}
		if iteratorSymbol == nil {
			return fmt.Errorf("nil iteratorSymbol")
		}
		labelTest := g.newLabel() // e.g. "L1"
		labelBody := g.newLabel() // e.g. "L2"
		labelEnd := g.newLabel()  // e.g. "L3"
		g.emit(Instruction{
			Op:   OpAssign, // "="
			Arg1: iteratorSymbol,
			Arg2: startSymbol,
		})
		g.emit(Instruction{
			Op:     OpGoto,
			JumpTo: labelTest,
		})
		g.emit(Instruction{Labels: []string{labelTest}})
		if !node.IsDownTo {
			// ascending
			g.emit(Instruction{
				Op:     OpIfLE, // "if<="
				Arg1:   iteratorSymbol,
				Arg2:   endSymbol,
				JumpTo: labelBody,
			})
		} else {
			// descending
			g.emit(Instruction{
				Op:     OpIfGE, // "if>="
				Arg1:   iteratorSymbol,
				Arg2:   endSymbol,
				JumpTo: labelBody,
			})
		}
		g.emit(Instruction{
			Op:     OpGoto,
			JumpTo: labelEnd,
		})
		g.emit(Instruction{Labels: []string{labelBody}})
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
				Arg1:        iteratorSymbol,
				Arg2:        oneSymbol,
			})
			// i = tX
			g.emit(Instruction{
				Op:   OpAssign,
				Arg1: iteratorSymbol,
				Arg2: tmp,
			})
		} else {
			// i = i - 1
			tmp := g.newTemp()
			g.emit(Instruction{
				Op:          OpSub, // "tX = i - 1"
				Destination: tmp,
				Arg1:        iteratorSymbol,
				Arg2:        oneSymbol,
			})
			g.emit(Instruction{
				Op:   OpAssign,
				Arg1: iteratorSymbol,
				Arg2: tmp,
			})
		}

		// 9. Jump back to labelTest
		g.emit(Instruction{
			Op:     OpGoto,
			JumpTo: labelTest,
		})
		g.emit(Instruction{Labels: []string{labelEnd}})
	case *ast.ProcCallCommand:
		funcSym, err := g.SymbolTable.Lookup(node.Name.String(), "xxFunctionsxx")
		if err != nil {
			return fmt.Errorf("failed looking up function symbol: %v", err)
		}
		for _, arg := range node.Args {
			argName := arg.String()
			symbol, err := g.SymbolTable.Lookup(argName, g.currentProc)
			if err == nil {
				g.emit(Instruction{
					Op:   OpParam,
					Arg1: symbol,
				})
			} else {
				return fmt.Errorf("failed calling %s: %v", argName, err)
			}
		}

		g.emit(Instruction{
			Op:   OpCall,
			Arg1: funcSym,
		})

	case *ast.RepeatCommand:
		labelStart := g.newLabel()
		labelEnd := g.newLabel()
		g.emit(Instruction{Labels: []string{labelStart}})
		for _, cmd := range node.Commands {
			if err := g.Generate(cmd); err != nil {
				g.Errors = append(g.Errors, err.Error())
			}
		}

		// 4. Generate code to test the condition
		//    If the condition is true, jump to labelEnd; otherwise jump to labelStart
		err := g.generateCondition(node.Condition, labelEnd, labelStart)
		if err != nil {
			return err
		}

		// 5. Emit labelEnd (the exit point)
		g.emit(Instruction{Labels: []string{labelEnd}})

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
			return err
		}

		g.emit(Instruction{Labels: []string{labelThen}})
		for _, cmd := range node.ThenCommands {
			if err := g.Generate(cmd); err != nil {
				g.Errors = append(g.Errors, err.Error())
			}
		}

		if hasElse {
			g.emit(Instruction{
				Op:     OpGoto,
				JumpTo: labelEnd,
			})
			g.emit(Instruction{
				Labels: []string{labelElse},
			})
			for _, cmd := range node.ElseCommands {
				if err := g.Generate(cmd); err != nil {
					g.Errors = append(g.Errors, err.Error())
				}
			}
		}

		g.emit(Instruction{Labels: []string{labelEnd}})

	default:
		return nil
	}
	return nil
}

func isNumber(str string) bool {
	_, err := strconv.Atoi(str)
	return (err == nil)
}

func MergeLabelOnlyInstructions(inss []Instruction) []Instruction {
	// for _, ins := range inss {
	// 	fmt.Println(ins)
	// }
	var result []Instruction
	i := 0
	for i < len(inss) {
		ins := inss[i]

		// Check if it's really just a "label" line,
		// i.e. it has a label and absolutely no other data.
		isLabelOnly := (len(ins.Labels) > 0 && ins.Op == "")
		if isLabelOnly {
			// Merge this label with the *next* instruction if one exists
			if i+1 < len(inss) {
				next := inss[i+1]
				if len(next.Labels) == 0 {
					next.Labels = ins.Labels
				} else {
					// If the next instruction also has a label, you could decide
					// how to handle multiple labels. For simplicity, join them:
					for _, label := range ins.Labels {
						next.Labels = append(next.Labels, label)
					}
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
		Op:     op,
		Arg1:   &left,
		Arg2:   &right,
		JumpTo: labelTrue,
	})

	// Unconditional jump to labelFalse
	g.emit(Instruction{
		Op:     OpGoto,
		JumpTo: labelFalse,
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

// generateValue returns a symbol “place” for the given Value.
// If it’s just a number literal or identifier, we can return the string directly.
// If your language has array-index computations, you’d handle them here.
// generateValue returns a string “place” for the given Value.
// If it’s an array reference, it emits instructions to compute the address and load the value.
func (g *Generator) generateValue(v ast.Value) (symboltable.Symbol, error) {
	switch val := v.(type) {
	case *ast.NumberLiteral:
		numStr := val.String()
		// Declare the number as a constant
		sym, _ := g.SymbolTable.Declare(numStr, "main", symboltable.Symbol{
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
		return *sym, nil

	case *ast.Identifier:
		// Handle array indices
		if val.Index != "" {
			// Generate code for array element access
			arrSym, err := g.SymbolTable.Lookup(val.Value, g.currentProc)
			if err != nil {
				return symboltable.Symbol{}, err
			}

			// Create temporary for the loaded value
			tmp := g.newTemp()

			// Generate index calculation
			var indexSym *symboltable.Symbol
			if isNumber(val.Index) {
				indexSym, _ = g.SymbolTable.Declare(val.Index, "main", symboltable.Symbol{Name: val.Index, Kind: symboltable.CONSTANT})
			} else {
				indexSym, err = g.SymbolTable.Lookup(val.Index, g.currentProc)
			}
			if err != nil {
				return symboltable.Symbol{}, err
			}

			// Emit array load instructions
			g.emit(Instruction{
				Op:        OpAssign,
				Arg1:      tmp,
				Arg2:      arrSym,
				Arg2Index: indexSym.Name,
			})

			return *tmp, nil
		}
		sym, err := g.SymbolTable.Lookup(val.String(), g.currentProc)
		if err != nil {
			return symboltable.Symbol{}, fmt.Errorf("failed to lookup symbol %s :%v", v.String(), err)
		}
		return *sym, nil

	default:
		return symboltable.Symbol{}, fmt.Errorf("unhandled Value type %T", v)
	}
}

// generateMathExpression returns the place holding the result of the expression.
// If expression.Right is nil, there's no operator, so just return Left's place.
// Otherwise, emit an instruction to combine Left and Right.
func (g *Generator) generateMathExpression(me *ast.MathExpression) (*symboltable.Symbol, error) {
	leftPlace, err := g.generateValue(me.Left)
	if err != nil {
		return nil, err
	}

	// If there's no operator, it's just a single operand
	if me.Right == nil {
		return &leftPlace, nil
	}

	// We have an operator and a right operand
	rightPlace, err := g.generateValue(me.Right)
	if err != nil {
		return nil, err
	}

	// Create a fresh temporary for the operation result
	tmp := g.newTemp()

	// Map the token to our Op enum
	op := opFromToken(*me)
	if op == OpDiv {
		if me.Right.String() == "2" {
			goto exit
		} else {
			funcSym, err := g.SymbolTable.Lookup("built_in_div", "xxFunctionsxx")
			if err != nil {
				return nil, err
			}
			leftSym, err := g.SymbolTable.Lookup("built_in_left", "main")
			if err != nil {
				return nil, err
			}
			rightSym, err := g.SymbolTable.Lookup("built_in_right", "main")
			if err != nil {
				return nil, err
			}
			resultSym, err := g.SymbolTable.Lookup("built_in_result", "main")
			if err != nil {
				return nil, err
			}
			g.emit(Instruction{
				Op:   OpAssign,
				Arg1: leftSym,
				Arg2: &leftPlace,
			})
			g.emit(Instruction{
				Op:   OpAssign,
				Arg1: rightSym,
				Arg2: &rightPlace,
			})
			g.emit(Instruction{
				Op:   OpCall,
				Arg1: funcSym,
			})
			g.emit(Instruction{
				Op:   OpAssign,
				Arg1: tmp,
				Arg2: resultSym,
			})
			return tmp, nil
		}
	} else if op == OpMul {
		funcSym, err := g.SymbolTable.Lookup("built_in_mult", "xxFunctionsxx")
		if err != nil {
			return nil, err
		}
		leftSym, err := g.SymbolTable.Lookup("built_in_left", "main")
		if err != nil {
			return nil, err
		}
		rightSym, err := g.SymbolTable.Lookup("built_in_right", "main")
		if err != nil {
			return nil, err
		}
		resultSym, err := g.SymbolTable.Lookup("built_in_result", "main")
		if err != nil {
			return nil, err
		}
		g.emit(Instruction{
			Op:   OpAssign,
			Arg1: leftSym,
			Arg2: &leftPlace,
		})
		g.emit(Instruction{
			Op:   OpAssign,
			Arg1: rightSym,
			Arg2: &rightPlace,
		})
		g.emit(Instruction{
			Op:   OpCall,
			Arg1: funcSym,
		})
		g.emit(Instruction{
			Op:   OpAssign,
			Arg1: tmp,
			Arg2: resultSym,
		})
		return tmp, nil
	} else if op == OpMod {
		funcSym, err := g.SymbolTable.Lookup("built_in_mod", "xxFunctionsxx")
		if err != nil {
			return nil, err
		}
		leftSym, err := g.SymbolTable.Lookup("built_in_left", "main")
		if err != nil {
			return nil, err
		}
		rightSym, err := g.SymbolTable.Lookup("built_in_right", "main")
		if err != nil {
			return nil, err
		}
		resultSym, err := g.SymbolTable.Lookup("built_in_result", "main")
		if err != nil {
			return nil, err
		}
		g.emit(Instruction{
			Op:   OpAssign,
			Arg1: leftSym,
			Arg2: &leftPlace,
		})
		g.emit(Instruction{
			Op:   OpAssign,
			Arg1: rightSym,
			Arg2: &rightPlace,
		})
		g.emit(Instruction{
			Op:   OpCall,
			Arg1: funcSym,
		})
		g.emit(Instruction{
			Op:   OpAssign,
			Arg1: tmp,
			Arg2: resultSym,
		})
		return tmp, nil
	}

	// Emit the arithmetic instruction:
	//     tmp = leftPlace op rightPlace
exit:
	g.emit(Instruction{
		Op:          op,
		Destination: tmp,
		Arg1:        &leftPlace,
		Arg2:        &rightPlace,
	})

	return tmp, nil
}

func (g *Generator) emit(ins Instruction) {
	g.Instructions = append(g.Instructions, ins)
}

func (g *Generator) DeclareArgProcedure(decl ast.ArgDecl, procName string) (*symboltable.Symbol, error) {
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
		ArgumentIndex: argCount + 1,
	}
	sym, err := g.SymbolTable.Declare(name, procName, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to declare argument %v in procedure %s: %v", decl, procName, err)
	}
	return sym, nil
}

func (g *Generator) DeclareProcedure(decl ast.Declaration, procName string) error {
	name := decl.Pidentifier.Value
	symbol := symboltable.Symbol{
		Name: name,
		Kind: symboltable.DECLARATION,
	}
	got, _ := g.SymbolTable.Lookup(name, procName)
	if got == nil {
		_, err := g.SymbolTable.Declare(name, procName, symbol)
		if err != nil {
			return fmt.Errorf("failed to declare %v in procedure %s: %v", decl, procName, err)
		}
	}

	return nil
}

func (g *Generator) DeclareMain(decl ast.Declaration) error {
	name := decl.Pidentifier.Value
	var symbol symboltable.Symbol
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
		}
	} else {
		symbol = symboltable.Symbol{
			Name: name,
			Kind: symboltable.DECLARATION,
		}
	}
	_, err := g.SymbolTable.Declare(name, "main", symbol)
	if err != nil {
		return fmt.Errorf("failed to declare in main: %v", err)
	}
	return nil
}

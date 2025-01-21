package tac

import (
	"fmt"
	"strconv"

	"github.com/Meduza3/imp/ast"
	"github.com/Meduza3/imp/symboltable"
)

type Generator struct {
	SymbolTable  *symboltable.SymbolTable
	Instructions []Instruction
	Errors       []string

	labelCount int
	tempCount  int
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

func (g *Generator) newTemp() string {
	g.tempCount++
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
		for _, procedure := range node.Procedures {
			if procedure != nil {
				g.Generate(procedure)
			}
		}
		if node.Main != nil {
			g.Generate(node.Main)
		}
		g.emit(Instruction{
			Op: OpHalt,
		})

	case *ast.Procedure:
		g.emit(Instruction{Label: node.ProcHead.Name.Value})
		for _, decl := range node.ProcHead.ArgsDecl {
			err := g.DeclareArgProcedure(decl, node.ProcHead.Name.Value)
			if err != nil {
				g.Errors = append(g.Errors, err.Error())
			}
		}
		for _, decl := range node.Declarations {
			err := g.DeclareProcedure(decl, node.ProcHead.Name.Value)
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
	case *ast.Main:
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
	case *ast.AssignCommand:
		// 1. Generate a place (temp or variable) for the right-hand side
		place, err := g.generateMathExpression(&node.MathExpression)
		if err != nil {
			g.Errors = append(g.Errors, err.Error())
			return nil // or return err, up to you
		}

		// 2. Emit a final assignment: identifier = place
		g.emit(Instruction{
			Op:          OpAssign,
			Destination: node.Identifier.String(),
			Arg1:        place,
			Arg2:        "", // Not needed for a pure assignment
		})
	case *ast.WriteCommand:
		g.emit(Instruction{Op: OpWrite, Arg1: node.Value.String()})
	case *ast.ReadCommand:
		g.emit(Instruction{Op: OpRead, Arg1: node.Identifier.Value})
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

		startVal := node.From.String()
		endVal := node.To.String()

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
			g.emit(Instruction{
				Op:   OpParam,
				Arg1: argName,
			})
		}

		procName := node.Name.String()
		numArgs := len(node.Args)
		g.emit(Instruction{
			Op:   OpCall,
			Arg1: procName, // the procedure label/name
			Arg2: fmt.Sprintf("%d", numArgs),
		})

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
func (g *Generator) generateValue(v ast.Value) (string, error) {
	switch val := v.(type) {
	case *ast.NumberLiteral:
		// Return the integer text directly
		return val.String(), nil

	case *ast.Identifier:
		// Return something like "x" or "x[i]" if you track arrays
		return val.String(), nil

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
	name := decl.Name.Value
	isTable := decl.IsTable
	symbol := symboltable.Symbol{
		Name:    name,
		Kind:    symboltable.ARGUMENT,
		IsTable: isTable,
	}
	err := g.SymbolTable.DeclareProcedure(name, procName, symbol)
	if err != nil {
		return fmt.Errorf("failed to declare argument %v in procedure %s: %v", decl, procName)
	}
	return nil
}

func (g *Generator) DeclareProcedure(decl ast.Declaration, procName string) error {
	name := decl.Pidentifier.Value
	symbol := symboltable.Symbol{
		Name: name,
		Kind: symboltable.DECLARATION,
	}
	err := g.SymbolTable.DeclareProcedure(name, procName, symbol)
	if err != nil {
		return fmt.Errorf("failed to declare %v in procedure %s: %v", decl, procName)
	}
	return nil
}

func (g *Generator) DeclareMain(decl ast.Declaration) error {
	name := decl.Pidentifier.Value
	var symbol symboltable.Symbol
	if decl.IsTable {
		from, err := strconv.Atoi(decl.From.Value)
		if err != nil {
			return fmt.Errorf("failed parsing from value in main declaration %v. value:", decl, decl.From.Value)
		}
		to, err := strconv.Atoi(decl.To.Value)
		if err != nil {
			return fmt.Errorf("failed parsing to value in main declaration %v. value:", decl, decl.To.Value)
		}
		symbol = symboltable.Symbol{
			Name:    name,
			Kind:    symboltable.DECLARATION,
			IsTable: true,
			From:    from,
			To:      to,
			Size:    to - from,
		}
	} else {
		symbol = symboltable.Symbol{
			Name: name,
			Kind: symboltable.DECLARATION,
		}
	}
	err := g.SymbolTable.DeclareMain(name, symbol)
	if err != nil {
		return fmt.Errorf("failed to declare in main: %v", err)
	}
	return nil
}

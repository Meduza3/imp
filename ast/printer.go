package ast

import (
	"fmt"
	"strings"
)

type Printer struct {
	currentIndent int
	sb            strings.Builder
}

func NewPrinter() *Printer {
	return &Printer{
		currentIndent: 0,
	}
}

func (p *Printer) Indent() {
	p.currentIndent++
}

func (p *Printer) Dedent() {
	if p.currentIndent > 0 {
		p.currentIndent--
	}
}

func (p *Printer) getIndent() string {
	return strings.Repeat("  ", p.currentIndent) // 2 spaces per indent level
}

func (p *Printer) writeLine(line string) {
	p.sb.WriteString(p.getIndent())
	p.sb.WriteString(line)
	p.sb.WriteString("\n")
}

func (p *Printer) Print(program *Program) string {
	// Print Procedures
	for _, proc := range program.Procedures {
		p.printProcedure(proc)
		p.sb.WriteString("\n") // Add a newline between procedures
	}

	// Print Main
	p.printMain(program.Main)

	return p.sb.String()
}

func (p *Printer) printProcedure(proc *Procedure) {
	// Procedure Header
	p.writeLine(fmt.Sprintf("PROCEDURE %s IS", proc.ProcHead.String()))

	// Indent for Declarations and BEGIN
	p.Indent()

	// Print Declarations (if any)
	fmt.Println(len(proc.Declarations))
	if len(proc.Declarations) > 0 {
		declStrings := []string{}
		for _, decl := range proc.Declarations {
			declStrings = append(declStrings, decl.String())
		}
		p.writeLine(strings.Join(declStrings, ", "))
	}

	// Print BEGIN
	p.Dedent()
	p.writeLine("BEGIN")
	p.Indent()

	// Print Commands
	for _, cmd := range proc.Commands {
		p.printCommand(cmd)
	}

	p.Dedent() // Outdent after commands

	// Print END
	p.writeLine("END")

	p.Dedent() // Outdent after PROCEDURE
}

func (p *Printer) printMain(main *Main) {
	p.writeLine("PROGRAM IS")
	p.Indent()

	// Print Declarations (if any)
	if len(main.Declarations) > 0 {
		declStrings := []string{}
		for _, decl := range main.Declarations {
			declStrings = append(declStrings, decl.String())
		}
		p.writeLine(strings.Join(declStrings, ", "))
	}
	p.Dedent()
	// Print BEGIN
	p.writeLine("BEGIN")
	p.Indent()

	// Print Commands
	for _, cmd := range main.Commands {
		p.printCommand(cmd)
	}

	p.Dedent() // Outdent after commands

	// Print END
	p.writeLine("END")

	p.Dedent() // Outdent after PROGRAM
}

func (p *Printer) printCommand(cmd Command) {
	switch c := cmd.(type) {
	case *AssignCommand:
		p.printAssignCommand(c)
	case *IfCommand:
		p.printIfCommand(c)
	case *WhileCommand:
		p.printWhileCommand(c)
	case *ForCommand:
		p.printForCommand(c)
	case *ProcCallCommand:
		p.printProcCallCommand(c)
	case *ReadCommand:
		p.printReadCommand(c)
	case *WriteCommand:
		p.printWriteCommand(c)
	case *RepeatCommand:
		p.printRepeatCommand(c)
	default:
		p.writeLine(fmt.Sprintf("// Unhandled command type: %T", c))
	}
}

func (p *Printer) printAssignCommand(ac *AssignCommand) {
	var expr string
	if ac.MathExpression.Right == nil {
		expr = ac.MathExpression.Left.String()
	} else {
		expr = fmt.Sprintf("%s %s %s",
			ac.MathExpression.Left.String(),
			ac.MathExpression.Operator.Literal,
			ac.MathExpression.Right.String(),
		)
	}
	p.writeLine(fmt.Sprintf("%s := %s;", ac.Identifier.String(), expr))
}

func (p *Printer) printIfCommand(ic *IfCommand) {
	p.writeLine(fmt.Sprintf("IF %s THEN", ic.Condition.String()))
	p.Indent()

	// Print Then Commands
	for _, cmd := range ic.ThenCommands {
		p.printCommand(cmd)
	}

	p.Dedent()

	// Print Else Commands (if any)
	if len(ic.ElseCommands) > 0 {
		p.writeLine("ELSE")
		p.Indent()
		for _, cmd := range ic.ElseCommands {
			p.printCommand(cmd)
		}
		p.Dedent()
	}

	// Print ENDIF
	p.writeLine("ENDIF")
}

func (p *Printer) printWhileCommand(wc *WhileCommand) {
	p.writeLine(fmt.Sprintf("WHILE %s DO", wc.Condition.String()))
	p.Indent()

	// Print Commands within WHILE
	for _, cmd := range wc.Commands {
		p.printCommand(cmd)
	}

	p.Dedent()
	p.writeLine("ENDWHILE")
}

func (p *Printer) printForCommand(fc *ForCommand) {
	direction := "TO"
	if fc.IsDownTo {
		direction = "DOWNTO"
	}
	p.writeLine(fmt.Sprintf("FOR %s FROM %s %s %s DO",
		fc.Iterator.String(),
		fc.From.String(),
		direction,
		fc.To.String(),
	))
	p.Indent()

	// Print Commands within FOR
	for _, cmd := range fc.Commands {
		p.printCommand(cmd)
	}

	p.Dedent()
	p.writeLine("ENDFOR")
}

func (p *Printer) printProcCallCommand(pcc *ProcCallCommand) {
	args := []string{}
	for _, arg := range pcc.Args {
		args = append(args, arg.String())
	}
	p.writeLine(fmt.Sprintf("%s(%s);", pcc.Name.String(), strings.Join(args, ", ")))
}

func (p *Printer) printReadCommand(rc *ReadCommand) {
	p.writeLine(fmt.Sprintf("READ %s;", rc.Identifier.String()))
}

func (p *Printer) printWriteCommand(wc *WriteCommand) {
	p.writeLine(fmt.Sprintf("WRITE %s;", wc.Value.String()))
}

func (p *Printer) printRepeatCommand(rc *RepeatCommand) {
	p.writeLine("REPEAT")
	p.Indent()

	// Print Commands within REPEAT
	for _, cmd := range rc.Commands {
		p.printCommand(cmd)
	}

	p.Dedent()
	p.writeLine(fmt.Sprintf("UNTIL %s;", rc.Condition.String()))
}

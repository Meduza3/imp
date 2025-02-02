package code

import "fmt"

type Opcode = string

const (
	GET    Opcode = "GET"
	PUT           = "PUT"
	LOAD          = "LOAD"
	STORE         = "STORE"
	LOADI         = "LOADI"
	STOREI        = "STOREI"
	ADD           = "ADD"
	SUB           = "SUB"
	ADDI          = "ADDI"
	SUBI          = "SUBI"
	SET           = "SET"
	HALF          = "HALF"
	JUMP          = "JUMP"
	JPOS          = "JPOS"
	JZERO         = "JZERO"
	JNEG          = "JNEG"
	RTRN          = "RTRN"
	HALT          = "HALT"
	COMM          = "#"
)

type Definition struct {
	Name        string
	NumOperands int
}

var definitions = map[Opcode]*Definition{
	GET:    {"GET", 1},
	PUT:    {"PUT", 1},
	LOAD:   {"LOAD", 1},
	STORE:  {"STORE", 1},
	LOADI:  {"LOADI", 1},
	STOREI: {"STOREI", 1},
	ADD:    {"ADD", 1},
	SUB:    {"SUB", 1},
	ADDI:   {"ADDI", 1},
	SUBI:   {"SUBI", 1},
	SET:    {"SET", 1},
	HALF:   {"HALF", 0},
	JUMP:   {"JUMP", 1},
	JPOS:   {"JPOS", 1},
	JZERO:  {"JZERO", 1},
	JNEG:   {"JNEG", 1},
	RTRN:   {"RTRN", 1},
	HALT:   {"HALT", 0},
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}

type Instruction struct {
	Op          Opcode
	HasOperand  bool
	Operand     int
	Destination string
	Labels      []string
	Comment     string
}

func (i Instruction) String() string {
	var string string
	if len(i.Labels) != 0 {
		for _, label := range i.Labels {
			string += label + ": "
		}
	}
	if i.HasOperand {
		string += i.Op + " " + fmt.Sprintf("%d", i.Operand)
	} else {
		string += i.Op
	}
	if i.Destination != "" {
		string += " " + i.Destination
	}
	if i.Comment != "" {
		string += " # " + i.Comment
	}
	return string
}

func Make(op Opcode, operands ...int) (*Instruction, error) {
	def, ok := definitions[op]
	if !ok {
		return nil, fmt.Errorf("code %s not defined", op)
	}
	switch len(operands) {
	case 0:
		if def.NumOperands == 0 {
			return &Instruction{Op: op, HasOperand: false}, nil
		} else {
			return nil, fmt.Errorf("missing operand for opcode %s", op)
		}
	case 1:
		if def.NumOperands == 1 {
			return &Instruction{Op: op, HasOperand: true, Operand: operands[0]}, nil
		} else {
			return nil, fmt.Errorf("unnecesary operand for opcode %s", op)
		}
	default:
		return nil, fmt.Errorf("only 0 or 1 operands")
	}
}

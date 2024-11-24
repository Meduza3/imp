package code

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type Instructions []byte

type Opcode byte

const (
	GET Opcode = iota
	PUT
	LOAD
	STORE
	LOADI
	STOREI
	ADD
	SUB
	ADDI
	SUBI
	SET
	HALF
	JUMP
	JPOS
	JZERO
	JNEG
	RTRN
	HALT
	OpConstant
)

type Definition struct {
	Name          string
	OperandWidths []int // Number of bytes each operand uses up.
}

// All instruction definitions matching the C++ VM
var definitions = map[Opcode]*Definition{
	GET:    {"GET", []int{8}},    // 8 bytes for long long address
	PUT:    {"PUT", []int{8}},    // 8 bytes for long long address
	LOAD:   {"LOAD", []int{8}},   // 8 bytes for long long address
	STORE:  {"STORE", []int{8}},  // 8 bytes for long long address
	LOADI:  {"LOADI", []int{8}},  // 8 bytes for long long address
	STOREI: {"STOREI", []int{8}}, // 8 bytes for long long address
	ADD:    {"ADD", []int{8}},    // 8 bytes for long long address
	SUB:    {"SUB", []int{8}},    // 8 bytes for long long address
	ADDI:   {"ADDI", []int{8}},   // 8 bytes for long long address
	SUBI:   {"SUBI", []int{8}},   // 8 bytes for long long address
	SET:    {"SET", []int{8}},    // 8 bytes for long long value
	HALF:   {"HALF", []int{8}},   // 8 bytes for long long address
	JUMP:   {"JUMP", []int{8}},   // 8 bytes for long long offset
	JPOS:   {"JPOS", []int{8}},   // 8 bytes for long long offset
	JZERO:  {"JZERO", []int{8}},  // 8 bytes for long long offset
	JNEG:   {"JNEG", []int{8}},   // 8 bytes for long long offset
	RTRN:   {"RTRN", []int{8}},   // 8 bytes for long long address
	HALT:   {"HALT", []int{}},    // No operands
}

func (ins Instructions) String() string {
	var out strings.Builder

	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d $s\n", i, ins.formatInstruction(def, operands))

		i += 1 + read
	}

	return out.String()
}

func (ins Instructions) formatInstruction(def *Definition, operands []int64) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand length %d does not match defined %d",
			len(operands), operandCount)
	}

	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}

	return fmt.Sprintf("ERROR: unhandled operand count for %s", def.Name)
}

func ReadOperands(def *Definition, ins Instructions) ([]int64, int) {
	operands := make([]int64, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 8:
			operands[i] = int64(binary.BigEndian.Uint64(ins[offset:]))
			offset += 8
		}
	}

	return operands, offset
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}

func Make(op Opcode, operands ...int64) ([]byte, error) {
	def, ok := definitions[op]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}

	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	for i, operand := range operands {
		if i >= len(def.OperandWidths) {
			return nil, fmt.Errorf("too many operands provided for %s", def.Name)
		}
		width := def.OperandWidths[i]

		switch width {
		case 8:
			binary.BigEndian.PutUint64(instruction[offset:], uint64(operand))
		}
		offset += width
	}
	return instruction, nil
}

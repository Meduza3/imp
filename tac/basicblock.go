package tac

type BasicBlock struct {
	ID           int
	Instructions []Instruction
	Predecessors []*BasicBlock
	Successors   []*BasicBlock
}

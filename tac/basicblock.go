package tac

type BasicBlock struct {
	ID           int
	Instructions []Instruction
	Predecessors []*BasicBlock
	Successors   []*BasicBlock
}

func SplitIntoBasicBlocks(inss []Instruction) []BasicBlock {
	if len(inss) == 0 {
		return nil
	}

	// Step 1: Identify leader instructions
	leaders := make([]bool, len(inss))
	leaders[0] = true // The first instruction is always a leader

	for i, ins := range inss {
		if ins.Label != "" {
			leaders[i] = true
		}
		// Check if the current instruction is a branch (goto or conditional)
		switch ins.Op {
		case OpGoto, OpIfEQ, OpIfNE, OpIfLT, OpIfLE, OpIfGT, OpIfGE, OpCall:
			if i+1 < len(inss) {
				leaders[i+1] = true
			}
		}
	}

	// Collect indices of all leaders
	var leaderIndices []int
	for i, isLeader := range leaders {
		if isLeader {
			leaderIndices = append(leaderIndices, i)
		}
	}

	// Split the instructions into basic blocks based on leaders
	var blocks []BasicBlock
	for idx, start := range leaderIndices {
		end := len(inss)
		if idx < len(leaderIndices)-1 {
			end = leaderIndices[idx+1]
		}
		blockInss := inss[start:end]
		blocks = append(blocks, BasicBlock{
			ID:           len(blocks),
			Instructions: blockInss,
			Predecessors: nil,
			Successors:   nil,
		})
	}

	return blocks
}

func BuildFlowGraph(blocks []BasicBlock) []BasicBlock {
	// Step 1: Create label to block mapping
	labelToBlock := make(map[string]*BasicBlock)
	for i := range blocks {
		block := &blocks[i]
		if len(block.Instructions) > 0 && block.Instructions[0].Label != "" {
			labelToBlock[block.Instructions[0].Label] = block
		}
	}

	// Step 2: Connect blocks based on control flow
	for i := range blocks {
		block := &blocks[i]
		if len(block.Instructions) == 0 {
			continue
		}

		lastIns := block.Instructions[len(block.Instructions)-1]
		var successors []*BasicBlock

		switch lastIns.Op {
		case OpGoto:
			if target := labelToBlock[lastIns.JumpTo]; target != nil {
				successors = append(successors, target)
			}

		case OpIfEQ, OpIfNE, OpIfLT, OpIfLE, OpIfGT, OpIfGE:
			// Conditional branch has two successors:
			// 1. The explicit target block
			// 2. Implicit fall-through to next block
			if target := labelToBlock[lastIns.JumpTo]; target != nil {
				successors = append(successors, target)
			}
			if i+1 < len(blocks) {
				successors = append(successors, &blocks[i+1])
			}

		case OpRet, OpHalt:
			// No successors for return/halt

		default:
			// Fall through to next block
			if i+1 < len(blocks) {
				successors = append(successors, &blocks[i+1])
			}
		}

		// Update successor/predecessor relationships
		for _, succ := range successors {
			block.Successors = append(block.Successors, succ)
			succ.Predecessors = append(succ.Predecessors, block)
		}
	}

	return blocks
}

func GetBlockIDs(blocks []*BasicBlock) []int {
	var ids []int
	for _, b := range blocks {
		ids = append(ids, b.ID)
	}
	return ids
}

package ir

import (
	"fmt"
)

type BasicBlock struct {
	Label        string          // Optional: a display name or discovered label.
	Instrs       []IRInstruction // The IR instructions in this block
	Predecessors []*BasicBlock
	Successors   []*BasicBlock
}

func BuildBasicBlocks(instrs []IRInstruction) []*BasicBlock {
	if len(instrs) == 0 {
		return nil
	}

	// Step 1: Identify block boundaries.
	// We'll keep an index -> bool map for "does a new block start here?"
	startsBlock := make([]bool, len(instrs))

	// The first instruction is always a block start
	startsBlock[0] = true

	// Mark block boundaries
	for i := 0; i < len(instrs); i++ {
		op := instrs[i].Op

		if op == OpLabel {
			// A label always begins a new block
			startsBlock[i] = true
			// If there’s a preceding instruction, the next one after it
			// might also be forced to start a block— but typically the label
			// itself is the first instruction of the new block.
		}

		// If the current instruction can transfer control
		// unconditionally or otherwise, the next instruction starts a block
		if op == OpGoto || op == OpIfEQ || op == OpIfNEQ || op == OpIfGEQ || op == OpIfGR ||
			op == OpIfLE || op == OpIfLEQ || op == OpRet || op == OpHalt {

			if i+1 < len(instrs) {
				startsBlock[i+1] = true
			}
		}
	}

	// Step 2: Create the BasicBlock objects by grouping instructions.
	var blocks []*BasicBlock
	var current *BasicBlock

	for i := 0; i < len(instrs); i++ {
		if startsBlock[i] {
			// Begin a new basic block
			current = &BasicBlock{}

			blocks = append(blocks, current)
			// If this instruction is OpLabel, we can store it as the block label
			if instrs[i].Op == OpLabel {
				current.Label = instrs[i].Label
			}
		}

		// Add this instruction to the current block
		current.Instrs = append(current.Instrs, instrs[i])
	}

	// Step 3: Build edges between blocks.
	// A convenient way is to map from label -> block so we can find jumps easily.
	labelToBlock := make(map[string]*BasicBlock)
	for _, blk := range blocks {
		if blk.Label != "" {
			labelToBlock[blk.Label] = blk
		}
	}

	// Helper to link two blocks
	link := func(from, to *BasicBlock) {
		from.Successors = append(from.Successors, to)
		to.Predecessors = append(to.Predecessors, from)
	}

	// Now iterate over each block, look at its last instruction:
	for i, blk := range blocks {
		if len(blk.Instrs) == 0 {
			continue // skip empty block, though typically shouldn't happen
		}
		last := blk.Instrs[len(blk.Instrs)-1]

		switch last.Op {
		case OpGoto:
			// Unconditional jump to last.Label
			if target, ok := labelToBlock[last.Label]; ok {
				link(blk, target)
			}
		case OpIfEQ, OpIfNEQ, OpIfGEQ, OpIfGR, OpIfLE, OpIfLEQ:
			// Conditional jump =>
			// 1) link to the block labeled by last.Label
			// 2) link to the next block in sequence (fallthrough), if any
			if target, ok := labelToBlock[last.Label]; ok {
				link(blk, target)
			}
			// If there's a next block in the list, link to it
			if i+1 < len(blocks) {
				link(blk, blocks[i+1])
			}
		case OpRet, OpHalt:
			// No successors
		default:
			// If not an explicit jump/ret/halt, we fall through to next block (if any)
			if i+1 < len(blocks) {
				link(blk, blocks[i+1])
			}
		}
	}

	// Return the list of blocks
	return blocks
}
func PrintBasicBlocks(blocks []*BasicBlock) {
	for i, blk := range blocks {
		fmt.Printf("Block %d:\n", i)

		// If the block has a label, display it.
		if blk.Label != "" {
			fmt.Printf("  Label: %s\n", blk.Label)
		}

		// Show the instructions
		fmt.Println("  Instructions:")
		for _, instr := range blk.Instrs {
			fmt.Printf("    %s\n", instr.String())
		}

		// Show the successor blocks
		if len(blk.Successors) > 0 {
			fmt.Printf("  Successors: ")
			for _, s := range blk.Successors {
				// Find the index of s in blocks. We'll do a quick linear search
				// here, but you could also store indexes in the block structure.
				idx := indexOfBlock(blocks, s)
				fmt.Printf("%d ", idx)
			}
			fmt.Println()
		}

		// Show the predecessor blocks
		if len(blk.Predecessors) > 0 {
			fmt.Printf("  Predecessors: ")
			for _, p := range blk.Predecessors {
				idx := indexOfBlock(blocks, p)
				fmt.Printf("%d ", idx)
			}
			fmt.Println()
		}

		fmt.Println()
	}
}

// indexOfBlock returns the index of 'b' in 'blocks' or -1 if not found.
// This is just a helper to print the block index in successors/predecessors.
func indexOfBlock(blocks []*BasicBlock, b *BasicBlock) int {
	for i, blk := range blocks {
		if blk == b {
			return i
		}
	}
	return -1
}

// ExtractProcedures takes a *full* IR listing (potentially with multiple
// procedures + main code) and returns a map of procedureName -> instructions.
//
// We assume that whenever we see an OpLabel whose label doesn't look like
// a "simple" block label (e.g. "L1", "L2"), it's a procedure label. Then
// we gather instructions until we see "ret" or the next procedure label.
func ExtractProcedures(instrs []IRInstruction) (map[string][]IRInstruction, []IRInstruction) {
	// We’ll also return a slice for any top-level “main” instructions that
	// aren't inside a labeled procedure.
	// E.g., you may have lines like "read a" "read b" "call pd" "halt"
	// after all the procedure definitions.
	procedures := make(map[string][]IRInstruction)
	var mainInstrs []IRInstruction

	var currentProcName string
	var currentProcInstrs []IRInstruction

	finalizeProc := func() {
		// Save currentProcInstrs in the map under currentProcName
		// as long as currentProcName is not empty.
		if currentProcName != "" && len(currentProcInstrs) > 0 {
			procedures[currentProcName] = currentProcInstrs
		}
		// Reset
		currentProcName = ""
		currentProcInstrs = nil
	}

	isLikelyProcedureLabel := func(label string) bool {
		// Very naive check:
		// If label starts with "L" followed by a number, we assume it's just a block label.
		// Otherwise, we assume it's a procedure name.
		// Example: "L1", "L2", etc. -> block label
		// "pa", "pb", "de" -> procedure name
		// Adjust as needed for your naming conventions.
		if len(label) == 0 {
			return false
		}
		if label[0] == 'L' {
			// Check if the rest is purely digits
			for i := 1; i < len(label); i++ {
				if label[i] < '0' || label[i] > '9' {
					return true // not purely "L\d+", so let's say it's a proc
				}
			}
			// If we get here, it’s "L123" form -> a block label
			return false
		}
		return true // not "L##", so probably a procedure
	}

	for _, ins := range instrs {
		if ins.Op == OpLabel && isLikelyProcedureLabel(ins.Label) {
			// We found a new procedure label
			// If we were collecting instructions for a previous procedure,
			// finalize it first
			finalizeProc()

			// Start a new procedure
			currentProcName = ins.Label
			currentProcInstrs = []IRInstruction{ins}
			continue
		}

		if currentProcName == "" {
			// We’re not in a procedure, so these lines belong to mainInstrs
			mainInstrs = append(mainInstrs, ins)
		} else {
			// We’re inside a procedure
			currentProcInstrs = append(currentProcInstrs, ins)

			// If we see 'ret', that ends this procedure
			if ins.Op == OpRet {
				// finalize
				finalizeProc()
			}
		}
	}

	// If we ended the file but still have a procedure open (no trailing 'ret'):
	finalizeProc()

	return procedures, mainInstrs
}

func BuildAllProceduresCFG(allInstrs []IRInstruction) (map[string][]*BasicBlock, []*BasicBlock) {
	procs, mainInstrs := ExtractProcedures(allInstrs)

	cfgs := make(map[string][]*BasicBlock)
	for procName, procInstrs := range procs {
		blocks := BuildBasicBlocks(procInstrs)
		cfgs[procName] = blocks
	}

	var mainBlocks []*BasicBlock
	if len(mainInstrs) > 0 {
		mainBlocks = BuildBasicBlocks(mainInstrs)
	}
	return cfgs, mainBlocks
}

func PrintAllProceduresCFG(cfgs map[string][]*BasicBlock, mainBlocks []*BasicBlock) {
	for procName, blocks := range cfgs {
		fmt.Printf("==== Procedure %s ====\n", procName)
		PrintBasicBlocks(blocks)
	}

	if len(mainBlocks) > 0 {
		fmt.Println("==== Main (top-level) ====")
		PrintBasicBlocks(mainBlocks)
	}
}

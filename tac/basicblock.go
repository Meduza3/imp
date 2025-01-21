package tac

type BasicBlock struct {
	ID           int
	Instructions []Instruction
	Predecessors []*BasicBlock
	Successors   []*BasicBlock
}

type CFG struct {
	Blocks []*BasicBlock
}

type DataFlowAnalysis struct {
	In   map[*BasicBlock]Set
	Out  map[*BasicBlock]Set
	Gen  map[*BasicBlock]Set
	Kill map[*BasicBlock]Set
}

type Set struct {
	elements map[string]struct{}
}

func NewSet() *Set {
	return &Set{
		elements: make(map[string]struct{}),
	}
}

func (s *Set) Add(element string) {
	s.elements[element] = struct{}{}
}

// Remove removes an element from the set
func (s *Set) Remove(element string) {
	delete(s.elements, element)
}

// Contains checks if the set contains an element
func (s *Set) Contains(element string) bool {
	_, exists := s.elements[element]
	return exists
}

// Union performs a union with another set
func (s *Set) Union(other *Set) *Set {
	result := NewSet()
	for key := range s.elements {
		result.Add(key)
	}
	for key := range other.elements {
		result.Add(key)
	}
	return result
}

// String returns the set as a string
func (s *Set) String() string {
	result := "{"
	first := true
	for key := range s.elements {
		if !first {
			result += ", "
		}
		result += key
		first = false
	}
	result += "}"
	return result
}


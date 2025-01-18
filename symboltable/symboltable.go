package symboltable

type SymbolTable map[string]Symbol

type SymbolType = string

type Symbol struct {
	Name         string
	Type         SymbolType // "PROCEDURE" "IDENTIFIER" "ARRAY"
	Arguments    SymbolTable
	Declarations SymbolTable
	Base         string // For array base address
	Value        string // The numeric value behind the name of type identifier
}

func New() SymbolTable {
	st := make(SymbolTable)
	return st
}

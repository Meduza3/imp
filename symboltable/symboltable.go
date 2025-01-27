package symboltable

import "fmt"

type SymbolTable struct {
	MainTable       map[string]Symbol
	ProcedureTables map[string]map[string]Symbol
}

type SymbolKind string

const (
	DECLARATION SymbolKind = "DECLARATION"
	ARGUMENT    SymbolKind = "ARGUMENT"
)

type Symbol struct {
	Name    string
	Kind    SymbolKind
	IsTable bool
	Address int
	From    int
	To      int
	Size    int
}

func New() *SymbolTable {
	return &SymbolTable{
		MainTable:       make(map[string]Symbol),
		ProcedureTables: make(map[string]map[string]Symbol),
	}
}

func (st *SymbolTable) DeclareMain(name string, symbol Symbol) error {
	got, ok := st.MainTable[name]
	if ok {
		return fmt.Errorf("failed to declare main for %q, already declared: %v", name, got)
	}
	st.MainTable[name] = symbol
	return nil
}

func (st *SymbolTable) DeclareProcedure(name, procedureName string, symbol Symbol) error {
	if st.ProcedureTables[procedureName] == nil {
		st.ProcedureTables[procedureName] = make(map[string]Symbol)
		st.ProcedureTables[procedureName][name] = symbol
	}
	got, ok := st.ProcedureTables[procedureName][name]
	if ok {
		return fmt.Errorf("failed to declare for procedure %s for %q, already declared: %v", procedureName, name, got)
	}
	st.ProcedureTables[procedureName][name] = symbol
	return nil
}

// Lookup searches for a symbol in the main table and specific procedure table (if given).
func (st *SymbolTable) Lookup(name, procedureName string) (*Symbol, error) {
	if procedureName != "" {
		if procedureSymbols, ok := st.ProcedureTables[procedureName]; ok {
			if symbol, ok := procedureSymbols[name]; ok {
				return &symbol, nil
			}
		}
		return nil, fmt.Errorf("symbol %q not found in procedure %q", name, procedureName)
	}

	if symbol, ok := st.MainTable[name]; ok {
		return &symbol, nil
	}
	return nil, fmt.Errorf("symbol %q not found in main table", name)
}

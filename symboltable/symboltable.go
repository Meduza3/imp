package symboltable

import "fmt"

type SymbolTable struct {
	Table         map[string]map[string]Symbol
	CurrentOffset int
}

type SymbolKind string

const (
	DECLARATION SymbolKind = "DECLARATION"
	ARGUMENT    SymbolKind = "ARGUMENT"
	PROCEDURE   SymbolKind = "PROCEDURE"
	CONSTANT    SymbolKind = "CONSTANT"
)

type Symbol struct {
	Name     string
	Kind     SymbolKind
	ArgCount int
	IsTable  bool
	Address  int
	From     int
	To       int
	Size     int
}

func New() *SymbolTable {
	pt := make(map[string]map[string]Symbol)
	pt["main"] = make(map[string]Symbol)
	return &SymbolTable{
		Table:         pt,
		CurrentOffset: 1,
	}
}
func (st *SymbolTable) Declare(name, procedureName string, symbol Symbol) error {
	if st.Table[procedureName] == nil {
		st.Table[procedureName] = make(map[string]Symbol)
		symbol.Address = st.CurrentOffset
		st.Table[procedureName][name] = symbol
		st.CurrentOffset++
	}
	got, ok := st.Table[procedureName][name]
	if ok {
		return fmt.Errorf("failed to declare for procedure %s for %q, already declared: %v", procedureName, name, got)
	}

	symbol.Address = st.CurrentOffset
	st.Table[procedureName][name] = symbol
	st.CurrentOffset++
	return nil
}

// Lookup searches for a symbol in the main table and specific procedure table (if given).
func (st *SymbolTable) Lookup(name, procedureName string) (*Symbol, error) {
	if procedureName != "" {
		if procedureSymbols, ok := st.Table[procedureName]; ok {
			if symbol, ok := procedureSymbols[name]; ok {
				return &symbol, nil
			}
		}
		return nil, fmt.Errorf("symbol %q not found in procedure %q", name, procedureName)
	}

	if symbol, ok := st.Table["main"][name]; ok {
		return &symbol, nil
	}
	return nil, fmt.Errorf("symbol %q not found in main table", name)
}

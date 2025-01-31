package symboltable

import (
	"fmt"
	"io"
)

type SymbolTable struct {
	Table         map[string]map[string]Symbol
	CurrentOffset int
}

func (st *SymbolTable) Display(w io.Writer, prefix string) {
	for function, table := range st.Table {
		io.WriteString(w, prefix+"=="+function+"==\n")
		for key, value := range table {
			io.WriteString(w, fmt.Sprintf("%s%s = %v\n", prefix, key, value))
		}
	}
}

type SymbolKind string

const (
	DECLARATION SymbolKind = "DECLARATION"
	ARGUMENT    SymbolKind = "ARGUMENT"
	PROCEDURE   SymbolKind = "PROCEDURE"
	CONSTANT    SymbolKind = "CONSTANT"
	RETURNADDR  SymbolKind = "RETURNADDR"
	TEMP        SymbolKind = "TEMP"
)

type Symbol struct {
	Name          string
	Kind          SymbolKind
	Address       int
	IsTable       bool
	From          int
	To            int
	Size          int
	ArgumentIndex int
	ArgCount      int
}

func New() *SymbolTable {
	pt := make(map[string]map[string]Symbol)
	pt["main"] = make(map[string]Symbol)
	return &SymbolTable{
		Table:         pt,
		CurrentOffset: 100,
	}
}
func (st *SymbolTable) Declare(name, procedureName string, symbol Symbol) error {
	if st.Table[procedureName] == nil {
		st.Table[procedureName] = make(map[string]Symbol)
		symbol.Address = st.CurrentOffset
		st.Table[procedureName][name] = symbol
		if symbol.Size > 0 {
			st.CurrentOffset += symbol.Size
		} else {
			st.CurrentOffset++
		}
	}
	got, ok := st.Table[procedureName][name]
	if ok {
		return fmt.Errorf("failed to declare for procedure %s for %q, already declared: %v", procedureName, name, got)
	}

	symbol.Address = st.CurrentOffset
	st.Table[procedureName][name] = symbol
	if symbol.Size > 0 {
		st.CurrentOffset += symbol.Size
	} else {
		st.CurrentOffset++
	}
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

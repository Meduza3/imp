package symboltable

import (
	"fmt"
	"io"
)

type SymbolTable struct {
	Table         map[string]map[string]*Symbol
	CurrentOffset int
}

func (st *SymbolTable) IncreaseOffset(by int) {
	st.CurrentOffset += by
}

func (st *SymbolTable) Display(w io.Writer, prefix string) {
	for function, table := range st.Table {
		io.WriteString(w, prefix+"=="+function+"==\n")
		for key, value := range table {
			io.WriteString(w, fmt.Sprintf("%s%s = %v\n", prefix, key, value))
		}
	}
}

func (st *SymbolTable) Initialize(sym *Symbol, currentProc string) {
	// Mark the symbol as initialized.
	sym.IsInitialized = true

	// First, try to update the symbol in the current procedure’s map.
	if procTable, ok := st.Table[currentProc]; ok {
		if _, exists := procTable[sym.Name]; exists {
			procTable[sym.Name] = sym
			return
		}
	}
	// If it isn’t found in the current procedure’s table, try the main table.
	if mainTable, ok := st.Table["main"]; ok {
		if _, exists := mainTable[sym.Name]; exists {
			mainTable[sym.Name] = sym
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
	ITERATOR    SymbolKind = "ITERATOR"
	TEMP        SymbolKind = "TEMP"
)

type Symbol struct {
	Name          string
	IsInitialized bool
	Kind          SymbolKind
	Address       int
	IsTable       bool
	From          int
	To            int
	Size          int
	Arguments     []*Symbol
	ArgumentsType []SymbolKind
	ArgumentIndex int

	ArgCount int
}

func New() *SymbolTable {
	pt := make(map[string]map[string]*Symbol)
	pt["main"] = make(map[string]*Symbol)
	return &SymbolTable{
		Table:         pt,
		CurrentOffset: 100,
	}
}
func (st *SymbolTable) Declare(name, procedureName string, symbol Symbol) (*Symbol, error) {
	// 1. Ensure the procedure map exists.
	if st.Table[procedureName] == nil {
		st.Table[procedureName] = make(map[string]*Symbol)
	}

	// 2. Check if it's already declared.
	if got, ok := st.Table[procedureName][name]; ok {
		return got, fmt.Errorf(
			"failed to declare symbol %q in procedure %q: already declared",
			name, procedureName,
		)
	}

	// 3. Assign address, store in the table, update offset.
	if symbol.IsTable {
		symbol.Address = st.CurrentOffset - symbol.From
	} else {
		symbol.Address = st.CurrentOffset
	}
	st.Table[procedureName][name] = &symbol

	if symbol.Size > 0 {
		st.CurrentOffset += symbol.Size
	} else {
		st.CurrentOffset++
	}
	return st.Table[procedureName][name], nil
}

// Lookup searches for a symbol in the main table and specific procedure table (if given).
func (st *SymbolTable) Lookup(name, procedureName string) (*Symbol, error) {
	// If a procedure name is provided, try to look up in that procedure's symbol table.
	if procedureName != "" {
		if procedureSymbols, ok := st.Table[procedureName]; ok {
			if symbol, ok := procedureSymbols[name]; ok {
				return symbol, nil
			}
		}
	}

	// Always try the main table as a fallback.
	if symbol, ok := st.Table["main"][name]; ok {
		return symbol, nil
	}

	// Symbol was not found in either place.
	if procedureName != "" {
		return nil, fmt.Errorf("symbol %q not found in procedure %q or in main", name, procedureName)
	}
	return nil, fmt.Errorf("symbol %q not found in main table", name)
}

package cache

import "fmt"

type table map[string]map[string]interface{}

func newTable() table {
	return make(map[string]map[string]interface{})
}

func (t table) newEntry(name string, symbolType string, kind Kind, index int) {
	t[name] = map[string]interface{}{"type": symbolType, "kind": kind, "index": index}
}

type SymbolTable struct {
	classTable      table
	subroutineTable table
	FieldIndex      int
	staticIndex     int
	varIndex        int
	argIndex        int
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		classTable:      newTable(),
		subroutineTable: newTable(),
		FieldIndex:      0,
		staticIndex:     0,
		varIndex:        0,
		argIndex:        0,
	}
}

func (s *SymbolTable) StartSubroutine() {
	s.subroutineTable = newTable()
	s.varIndex = 0
	s.argIndex = 0
}

type Kind uint8

const (
	Var Kind = iota
	Arg
	Static
	Field
	None
)

func (s Kind) String() string {
	switch s {
	case Var:
		return "var"
	case Arg:
		return "arg"
	case Static:
		return "static"
	case Field:
		return "field"
	default:
		return ""
	}
}

func ParseKind(kind string) Kind {
	switch kind {
	case "var":
		return Var
	case "arg":
		return Arg
	case "static":
		return Static
	case "field":
		return Field
	default:
		return None
	}
}

func (s *SymbolTable) Define(name string, symbolType string, kind Kind) {
	switch kind {
	case Var:
		s.subroutineTable.newEntry(name, symbolType, kind, s.varIndex)
		s.varIndex++
	case Arg:
		s.subroutineTable.newEntry(name, symbolType, kind, s.argIndex)
		s.argIndex++
	case Static:
		s.classTable.newEntry(name, symbolType, kind, s.staticIndex)
		s.staticIndex++
	case Field:
		s.classTable.newEntry(name, symbolType, kind, s.FieldIndex)
		s.FieldIndex++
	default:
		panic(fmt.Sprintf("invalid symbol kind: %s", kind))
	}
}

func (s *SymbolTable) VarCount(kind Kind) int {
	switch kind {
	case Var:
		return s.varIndex
	case Arg:
		return s.argIndex
	case Static:
		return s.staticIndex
	case Field:
		return s.FieldIndex
	}
	return -1
}

func (s *SymbolTable) KindOf(name string) Kind {
	if kind, ok := s.classTable[name]["kind"]; ok {
		return kind.(Kind)
	}

	if kind, ok := s.subroutineTable[name]["kind"]; ok {
		return kind.(Kind)
	}
	return None
}

func (s *SymbolTable) TypeOf(name string) string {
	if symbolType, ok := s.classTable[name]["type"]; ok {
		return symbolType.(string)
	}

	if symbolType, ok := s.subroutineTable[name]["type"]; ok {
		return symbolType.(string)
	}

	return ""
}

func (s *SymbolTable) IndexOf(name string) int {
	if index, ok := s.classTable[name]["index"]; ok {
		return index.(int)
	}

	if index, ok := s.subroutineTable[name]["index"]; ok {
		return index.(int)
	}

	return -1
}

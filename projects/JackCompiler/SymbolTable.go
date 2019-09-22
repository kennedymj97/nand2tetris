package JackCompiler

type table map[string]map[string]interface{}

func newTable() table {
	return make(map[string]map[string]interface{})
}

func (t table) newEntry(name string, symbolType string, kind symbolKind, index int) {
	t[name] = map[string]interface{}{"type": symbolType, "kind": kind, "index": index}
}

type symbolTable struct {
	classTable      table
	subroutineTable table
	fieldIndex      int
	staticIndex     int
	varIndex        int
	argIndex        int
}

func newSymbolTable() *symbolTable {
	return &symbolTable{
		classTable:      newTable(),
		subroutineTable: newTable(),
		fieldIndex:      0,
		staticIndex:     0,
		varIndex:        0,
		argIndex:        0,
	}
}

func (s *symbolTable) startSubroutine() {
	s.subroutineTable = newTable()
	s.varIndex = 0
	s.argIndex = 0
}

type symbolKind uint8

const (
	VAR symbolKind = iota
	ARG
	STATIC
	FIELD
	NONE
)

func (s *symbolTable) define(name string, symbolType string, kind symbolKind) {
	switch kind {
	case VAR:
		s.subroutineTable.newEntry(name, symbolType, kind, s.varIndex)
		s.varIndex++
	case ARG:
		s.subroutineTable.newEntry(name, symbolType, kind, s.argIndex)
		s.argIndex++
	case STATIC:
		s.classTable.newEntry(name, symbolType, kind, s.staticIndex)
		s.staticIndex++
	case FIELD:
		s.classTable.newEntry(name, symbolType, kind, s.fieldIndex)
		s.fieldIndex++
	}
}

func (s *symbolTable) varCount(kind symbolKind) int {
	switch kind {
	case VAR:
		return s.varIndex
	case ARG:
		return s.argIndex
	case STATIC:
		return s.staticIndex
	case FIELD:
		return s.fieldIndex
	}
	return -1
}

func (s *symbolTable) kindOf(name string) symbolKind {
	if kind, ok := s.classTable[name]["kind"]; ok {
		return kind.(symbolKind)
	}

	if kind, ok := s.subroutineTable[name]["kind"]; ok {
		return kind.(symbolKind)
	}
	return NONE
}

func (s *symbolTable) typeOf(name string) string {
	if symbolType, ok := s.classTable[name]["type"]; ok {
		return symbolType.(string)
	}

	if symbolType, ok := s.subroutineTable[name]["type"]; ok {
		return symbolType.(string)
	}

	return ""
}

func (s *symbolTable) indexOf(name string) int {
	if index, ok := s.classTable[name]["index"]; ok {
		return index.(int)
	}

	if index, ok := s.subroutineTable[name]["index"]; ok {
		return index.(int)
	}

	return -1
}

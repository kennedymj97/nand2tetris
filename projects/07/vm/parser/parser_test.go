package parser

import "testing"

func TestParser(t *testing.T) {
	commands := []struct {
		raw       string
		formatted string
		ctype     string
		arg1      string
		arg2      int
	}{
		{"", "", "", "", -1},
		{"// check comments are igonred", "", "", "", -1},
		{"add  ", "add", "C_ARITHMETIC", "add", -1},
		{"sub", "sub", "C_ARITHMETIC", "sub", -1},
		{"  neg", "neg", "C_ARITHMETIC", "neg", -1},
		{"eq //check if the stack is equal", "eq", "C_ARITHMETIC", "eq", -1},
		{"gt", "gt", "C_ARITHMETIC", "gt", -1},
		{"lt", "lt", "C_ARITHMETIC", "lt", -1},
		{"and", "and", "C_ARITHMETIC", "and", -1},
		{"or", "or", "C_ARITHMETIC", "or", -1},
		{"not", "not", "C_ARITHMETIC", "not", -1},
		{"push argument 5    ", "push argument 5", "C_PUSH", "argument", 5},
		{"   pop this 37", "pop this 37", "C_POP", "this", 37},
		{"label end //ignore this bit", "label end", "C_LABEL", "end", -1},
		{"goto loop", "goto loop", "C_GOTO", "loop", -1},
		{"if-goto test", "if-goto test", "C_IF", "test", -1},
		{"function sum 2", "function sum 2", "C_FUNCTION", "sum", 2},
		{"call mult 3", "call mult 3", "C_CALL", "mult", 3},
		{"return", "return", "C_RETURN", "", -1},
	}
	for _, command := range commands {
		var p VmParser
		p = &Command{Line: command.raw}
		formattedCommand := p.FormatLine()
		if formattedCommand != command.formatted {
			t.Errorf("Formatted command was incorrect, got: %s, wanted: %s", formattedCommand, command.formatted)
		}
		commandType, _ := p.CommandType()
		if commandType != command.ctype {
			t.Errorf("Command type was incorrect, got: %s, wanted: %s", commandType, command.ctype)
		}
		firstArg, _ := p.Arg1(commandType)
		if firstArg != command.arg1 {
			t.Errorf("Arg1 was incorrect, got: %s, wanted: %s", firstArg, command.arg1)
		}
		secondArg, _ := p.Arg2(commandType)
		if secondArg != command.arg2 {
			t.Errorf("Arg2 was incorrect, got: %d, wanted: %d", secondArg, command.arg2)
		}
	}
}

package parser

import (
	"fmt"
	"strconv"
	"strings"
)

type Parser interface {
	FormatLine() string
}

type VmParser interface {
	Parser
	CommandType() (string, error)
	Arg1(commandType string) (string, error)
	Arg2(commandType string) (int, error)
}

type Command struct {
	Line string
}

// FormatLine takes a command and removes all comments and whitespace
func (c *Command) FormatLine() string {
	c.Line = strings.ReplaceAll(c.Line, "\n", "")
	commentIndex := strings.Index(c.Line, "//")
	if commentIndex != -1 {
		c.Line = c.Line[0:commentIndex]
	}
	c.Line = strings.TrimSpace(c.Line)
	return c.Line
}

// CommandType returns a string detailing the type of the command, if the command type is invalid an error is returned.
func (c *Command) CommandType() (string, error) {
	firstSpaceIndex := strings.Index(c.Line, " ")
	var firstWord string
	if firstSpaceIndex == -1 {
		firstWord = c.Line
	} else {
		firstWord = c.Line[0:firstSpaceIndex]
	}
	arithmeticCommands := []string{
		"add",
		"sub",
		"neg",
		"eq",
		"gt",
		"lt",
		"and",
		"or",
		"not",
	}
	for _, command := range arithmeticCommands {
		if firstWord == command {
			return "C_ARITHMETIC", nil
		}
	}
	switch firstWord {
	case "push":
		return "C_PUSH", nil
	case "pop":
		return "C_POP", nil
	case "label":
		return "C_LABEL", nil
	case "goto":
		return "C_GOTO", nil
	case "if-goto":
		return "C_IF", nil
	case "function":
		return "C_FUNCTION", nil
	case "call":
		return "C_CALL", nil
	case "return":
		return "C_RETURN", nil
	default:
		return "", fmt.Errorf("%s is not a recognised command", firstWord)
	}
}

func (c *Command) Arg1(commandType string) (string, error) {
	switch commandType {
	case "C_ARITHMETIC":
		return c.Line, nil
	case "C_LABEL", "C_GOTO", "C_IF":
		spaceIndex := strings.Index(c.Line, " ")
		return c.Line[spaceIndex+1:], nil
	case "C_PUSH", "C_POP", "C_FUNCTION", "C_CALL":
		firstSpaceIndex := strings.Index(c.Line, " ")
		lastSpaceIndex := strings.LastIndex(c.Line, " ")
		return c.Line[firstSpaceIndex+1 : lastSpaceIndex], nil
	case "C_RETURN":
		return "", nil
	default:
		return "", fmt.Errorf("%s is not a valid command type", commandType)
	}
}

func (c *Command) Arg2(commandType string) (int, error) {
	switch commandType {
	case "C_PUSH", "C_POP", "C_FUNCTION", "C_CALL":
		lastSpaceIndex := strings.LastIndex(c.Line, " ")
		val, err := strconv.Atoi(c.Line[lastSpaceIndex+1:])
		if err != nil {
			return -1, err
		}
		return val, nil
	case "C_ARITHMETIC", "C_LABEL", "C_GOTO", "C_IF", "C_RETURN":
		return -1, nil
	default:
		return -1, fmt.Errorf("%s is not a valid command type", commandType)
	}
}

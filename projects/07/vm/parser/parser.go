package parser

type Parser interface {
	FormatLine(line string) string
}

type VmParser interface {
	CommandType(line string) string
	arg1(commandType string, line string) string
	arg2(commandType string, line string) int
}

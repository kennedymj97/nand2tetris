package translater

import "fmt"

type Translater interface {
	WriteArithmetic(equalityCheckCount int) (string, int)
	WritePushPop() string
}

type AssemblyWriter struct {
	CommandType string
	Arg1        string
	Arg2        int
}

func (aw *AssemblyWriter) WriteArithmetic(equalityCheckCount int) (string, int) {
	var assemblyCode string
	e := equalityCheckCount
	switch aw.Arg1 {
	case "add":
		assemblyCode = `@SP
M=M-1
A=M
D=M
A=A-1
M=M+D
`
		return assemblyCode, 0
	case "sub":
		assemblyCode = `@SP
M=M-1
A=M
D=M
A=A-1
M=M-D
`
		return assemblyCode, 0
	case "neg":
		assemblyCode = `@SP
A=M-1
M=-M
`
		return assemblyCode, 0
	case "eq":
		assemblyCode = fmt.Sprintf(`@SP
M=M-1
A=M
D=M
A=A-1
D=D-M
@EQUAL%d
D;JEQ
@SP
A=M-1
M=0
@END%d
0;JMP
(EQUAL%d)
@SP
A=M-1
M=-1
(END%d)
`, e, e, e, e)
		return assemblyCode, 1
	case "gt":
		assemblyCode = fmt.Sprintf(`@SP
M=M-1
A=M
D=M
A=A-1
D=M-D
@GREATER%d
D;JGT
@SP
A=M-1
M=0
@END%d
0;JMP
(GREATER%d)
@SP
A=M-1
M=-1
(END%d)	
`, e, e, e, e)
		return assemblyCode, 1
	case "lt":
		assemblyCode = fmt.Sprintf(`@SP
M=M-1
A=M
D=M
A=A-1
D=M-D
@LESS%d
D;JLT
@SP
A=M-1
M=0
@END%d
0;JMP
(LESS%d)
@SP
A=M-1
M=-1
(END%d)
`, e, e, e, e)
		return assemblyCode, 1
	case "and":
		assemblyCode = `@SP
M=M-1
A=M
D=M
A=A-1
M=D&M
`
		return assemblyCode, 0
	case "or":
		assemblyCode = `@SP
M=M-1
A=M
D=M
A=A-1
M=D|M
`
		return assemblyCode, 0
	case "not":
		assemblyCode = `@SP
A=M-1
M=!M	
`
		return assemblyCode, 0
	default:
		return "", 0
	}
}

func (aw *AssemblyWriter) WritePushPop() string {
	var assemblyCode string
	// 	if aw.CommandType == "C_PUSH" {
	// 		switch aw.Arg1 {
	// 		case "constant":
	// 			assemblyCode = fmt.Sprintf(`@%d
	// M=%d
	// `, aw.StackPointer, aw.Arg2)
	// 			return assemblyCode, aw.StackPointer + 1
	// 		}
	// 	}
	assemblyCode = fmt.Sprintf(`@%d
D=A
@SP
A=M
M=D
@SP
M=M+1
`, aw.Arg2)
	return assemblyCode
}

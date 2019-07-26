package translater

import "fmt"

type Translater interface {
	WriteArithmetic(equalityCheckCount int) (string, int)
	WritePushPop() string
	WriteLabel() string
	WriteGoto() string
	WriteIf() string
}

type AssemblyWriter struct {
	Filename    string
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
	if aw.CommandType == "C_PUSH" {
		switch aw.Arg1 {
		case "constant":
			assemblyCode = fmt.Sprintf(`@%d
D=A
@SP
A=M
M=D
@SP
M=M+1
`, aw.Arg2)
			return assemblyCode
		case "argument":
			assemblyCode = fmt.Sprintf(`@ARG
D=M
@%d
A=D+A
D=M
@SP
A=M
M=D
@SP
M=M+1
`, aw.Arg2)
			return assemblyCode
		case "local":
			assemblyCode = fmt.Sprintf(`@LCL
D=M
@%d
A=D+A
D=M
@SP
A=M
M=D
@SP
M=M+1
`, aw.Arg2)
			return assemblyCode
		case "this":
			assemblyCode = fmt.Sprintf(`@THIS
D=M
@%d
A=D+A
D=M
@SP
A=M
M=D
@SP
M=M+1
`, aw.Arg2)
			return assemblyCode
		case "that":
			assemblyCode = fmt.Sprintf(`@THAT
D=M
@%d
A=D+A
D=M
@SP
A=M
M=D
@SP
M=M+1
`, aw.Arg2)
			return assemblyCode
		case "temp":
			assemblyCode = fmt.Sprintf(`@5
D=A
@%d
A=D+A
D=M
@SP
A=M
M=D
@SP
M=M+1
`, aw.Arg2)
			return assemblyCode
		case "pointer":
			assemblyCode = fmt.Sprintf(`@3
D=A
@%d
A=D+A
D=M
@SP
A=M
M=D
@SP
M=M+1
`, aw.Arg2)
			return assemblyCode
		case "static":
			assemblyCode = fmt.Sprintf(`@%s.%d
D=M
@SP
A=M
M=D
@SP
M=M+1
`, aw.Filename, aw.Arg2)
			return assemblyCode
		}
	} else if aw.CommandType == "C_POP" {
		switch aw.Arg1 {
		case "argument":
			assemblyCode = fmt.Sprintf(`@SP
M=M-1
A=M
D=M
@ARG
A=M
D=D+A
@%d
D=D+A
@SP
A=M
A=D-M
D=D-A
M=D
`, aw.Arg2)
			return assemblyCode
		case "local":
			assemblyCode = fmt.Sprintf(`@SP
M=M-1
A=M
D=M
@LCL
A=M
D=D+A
@%d
D=D+A
@SP
A=M
A=D-M
D=D-A
M=D
`, aw.Arg2)
			return assemblyCode
		case "this":
			assemblyCode = fmt.Sprintf(`@SP
M=M-1
A=M
D=M
@THIS
A=M
D=D+A
@%d
D=D+A
@SP
A=M
A=D-M
D=D-A
M=D
`, aw.Arg2)
			return assemblyCode
		case "that":
			assemblyCode = fmt.Sprintf(`@SP
M=M-1
A=M
D=M
@THAT
A=M
D=D+A
@%d
D=D+A
@SP
A=M
A=D-M
D=D-A
M=D
`, aw.Arg2)
			return assemblyCode
		case "temp":
			assemblyCode = fmt.Sprintf(`@SP
M=M-1
A=M
D=M
@5
D=D+A
@%d
D=D+A
@SP
A=M
A=D-M
D=D-A
M=D
`, aw.Arg2)
			return assemblyCode
		case "pointer":
			assemblyCode = fmt.Sprintf(`@SP
M=M-1
A=M
D=M
@3
D=D+A
@%d
D=D+A
@SP
A=M
A=D-M
D=D-A
M=D
`, aw.Arg2)
			return assemblyCode
		case "static":
			assemblyCode = fmt.Sprintf(`@SP
M=M-1
A=M
D=M
@%s.%d
M=D
`, aw.Filename, aw.Arg2)
			return assemblyCode
		}
	}
	return ""
}

func (aw *AssemblyWriter) WriteLabel() string {
	assemblyCode := fmt.Sprintf(`(%s.tempfunction$%s)
`, aw.Filename, aw.Arg1)
	return assemblyCode
}

func (aw *AssemblyWriter) WriteGoto() string {
	assemblyCode := fmt.Sprintf(`@%s.tempfunction$%s
0;JMP	
`, aw.Filename, aw.Arg1)
	return assemblyCode
}

func (aw *AssemblyWriter) WriteIf() string {
	assemblyCode := fmt.Sprintf(`@SP
M=M-1
A=M
D=M
@%s.tempfunction$%s
D;JNE
`, aw.Filename, aw.Arg1)
	return assemblyCode
}

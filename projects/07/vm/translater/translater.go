package translater

import "fmt"

type Translater interface {
	WriteArithmetic() string
	WritePushPop() string
}

type AssemblyWriter struct {
	CommandType string
	Arg1        string
	Arg2        int
}

func (aw *AssemblyWriter) WriteArithmetic() string {
	var assemblyCode string
	switch aw.Arg1 {
	case "add":
		assemblyCode = `@SP
A=M-1
D=M
A=A-1
M=M+D
@SP
M=M-1
`
		return assemblyCode
		// 	case "sub":
		// 		assemblyCode = fmt.Sprintf(`@%d
		// D=M
		// @%d
		// D=D-M
		// @%d
		// M=D
		// `, x, y, x)
		// 		return assemblyCode
		// 	case "neg":
		// 		assemblyCode = fmt.Sprintf(`@%d
		// M=-M
		// `, y)
		// 		return assemblyCode
		// 	case "eq":

		// 		if x == y {
		// 			assemblyCode = fmt.Sprintf(`@%d
		// M=-1
		// `, x)
		// 		} else {
		// 			assemblyCode = fmt.Sprintf(`@%d
		// M=0
		// `, x)
		// 		}
		// 		return assemblyCode
		// 	case "gt":
		// 		if x > y {
		// 			assemblyCode = fmt.Sprintf(`@%d\n
		// M=-1\n
		// `, x)
		// 		} else {
		// 			assemblyCode = fmt.Sprintf(`
		// @%d\n
		// M=0\n
		// `, x)
		// 		}
		// 		return assemblyCode
		// 	case "lt":
		// 		if x < y {
		// 			assemblyCode = fmt.Sprintf(`@%d
		// M=-1
		// `, x)
		// 		} else {
		// 			assemblyCode = fmt.Sprintf(`@%d
		// M=0
		// `, x)
		// 		}
		// 		return assemblyCode
		// 	case "and":
		// 		assemblyCode = fmt.Sprintf(`@%d\n
		// D=M\n
		// @%d\n
		// D=D&M\n
		// @%d\n
		// M=D\n
		// `, x, y, x)
		// 		return assemblyCode
		// 	case "or":
		// 		assemblyCode = fmt.Sprintf(`@%d\n
		// D=M\n
		// @%d\n
		// D=D|M\n
		// @%d\n
		// M=D\n
		// `, x, y, x)
		// 		return assemblyCode
		// 	case "not":
		// 		assemblyCode = fmt.Sprintf(`@%d\n
		// M=!M\n
		// `, y)
		// 		return assemblyCode
	default:
		return ""
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

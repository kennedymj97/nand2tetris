package translater

import (
	"fmt"
	"strconv"
)

type Translater interface {
	WriteArithmetic(equalityCheckCount int) (string, int)
	WritePushPop() string
	WriteLabel() string
	WriteGoto() string
	WriteIf() string
	WriteFunction() string
	WriteReturn() string
	WriteCall(equalityCheckCount int) (string, int)
	WriteInit() string
}

type AssemblyWriter struct {
	Filename     string
	FunctionName string
	CommandType  string
	Arg1         string
	Arg2         int
}

func (aw *AssemblyWriter) WriteArithmetic(equalityCheckCount int) (string, int) {
	var assemblyCode string
	e := "$" + aw.FunctionName + strconv.Itoa(equalityCheckCount)
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
@EQUAL%s
D;JEQ
@SP
A=M-1
M=0
@END%s
0;JMP
(EQUAL%s)
@SP
A=M-1
M=-1
(END%s)
`, e, e, e, e)
		return assemblyCode, 1
	case "gt":
		assemblyCode = fmt.Sprintf(`@SP
M=M-1
A=M
D=M
A=A-1
D=M-D
@GREATER%s
D;JGT
@SP
A=M-1
M=0
@END%s
0;JMP
(GREATER%s)
@SP
A=M-1
M=-1
(END%s)	
`, e, e, e, e)
		return assemblyCode, 1
	case "lt":
		assemblyCode = fmt.Sprintf(`@SP
M=M-1
A=M
D=M
A=A-1
D=M-D
@LESS%s
D;JLT
@SP
A=M-1
M=0
@END%s
0;JMP
(LESS%s)
@SP
A=M-1
M=-1
(END%s)
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
	assemblyCode := fmt.Sprintf(`(%s$%s)
`, aw.FunctionName, aw.Arg1)
	return assemblyCode
}

func (aw *AssemblyWriter) WriteGoto() string {
	assemblyCode := fmt.Sprintf(`@%s$%s
0;JMP	
`, aw.FunctionName, aw.Arg1)
	return assemblyCode
}

func (aw *AssemblyWriter) WriteIf() string {
	assemblyCode := fmt.Sprintf(`@SP
M=M-1
A=M
D=M
@%s$%s
D;JNE
`, aw.FunctionName, aw.Arg1)
	return assemblyCode
}

func (aw *AssemblyWriter) WriteFunction() string {
	assemblyCode := fmt.Sprintf(`(%s)
`, aw.Arg1)

	for val := 0; val < aw.Arg2; val++ {
		if val == 0 {
			assemblyCode += `@LCL
A=M
M=0
@SP
A=M
M=0
@SP
M=M+1
`
		} else {
			assemblyCode += fmt.Sprintf(`@LCL
D=M
@%d
A=D+A
M=0
@SP
A=M
M=0
@SP
M=M+1
`, aw.Arg2)
		}
	}

	return assemblyCode
}

func (aw *AssemblyWriter) WriteReturn() string {

	restoreMemSegment := func(memorySegment string) string {
		var decrementValue string
		switch memorySegment {
		case "RET":
			decrementValue = "5"
		case "LCL":
			decrementValue = "4"
		case "ARG":
			decrementValue = "3"
		case "THIS":
			decrementValue = "2"
		case "THAT":
			decrementValue = "1"
		}

		var assemblyCode string
		if memorySegment == "RET" {
			assemblyCode = fmt.Sprintf(`@R14
D=M
@%s
A=D-A
D=M
@R15
M=D		
`, decrementValue)
		} else {
			assemblyCode = fmt.Sprintf(`@R14
D=M
@%s
A=D-A
D=M
@%s
M=D		
`, decrementValue, memorySegment)
		}

		return assemblyCode
	}

	setFrame := `@LCL
D=M
@R14
M=D
`
	setRet := restoreMemSegment("RET")
	popToArg := `@SP
A=M-1
D=M
@ARG
A=M
M=D
`
	restoreSp := `@ARG
D=M+1
@SP
M=D	
`
	restoreThat := restoreMemSegment("THAT")
	restoreThis := restoreMemSegment("THIS")
	restoreArg := restoreMemSegment("ARG")
	restoreLcl := restoreMemSegment("LCL")
	jumpToRet := `@R15
A=M
0;JMP
`

	return setFrame + setRet + popToArg + restoreSp + restoreThat + restoreThis + restoreArg + restoreLcl + jumpToRet
}

func (aw *AssemblyWriter) WriteCall(equalityCheckCount int) (string, int) {
	pushVariable := func(variable string) string {
		return fmt.Sprintf(`@%s
D=M
@SP
A=M
M=D
@SP
M=M+1
`, variable)
	}

	pushReturn := fmt.Sprintf(`@RETURN.%s.%d
D=A
@SP
A=M
M=D
@SP
M=M+1
`, aw.Arg1, equalityCheckCount)
	pushLcl := pushVariable("LCL")
	pushArg := pushVariable("ARG")
	pushThis := pushVariable("THIS")
	pushThat := pushVariable("THAT")

	repositionArgVal := fmt.Sprintf("%d", aw.Arg2+5)
	repositionArg := fmt.Sprintf(`@SP
D=M
@%s
D=D-A
@ARG
M=D
`, repositionArgVal)

	repositionLcl := `@SP
D=M
@LCL
M=D
`

	gotoFunc := fmt.Sprintf(`@%s
0;JMP	
`, aw.Arg1)

	returnLabel := fmt.Sprintf("(RETURN.%s.%d)\n", aw.Arg1, equalityCheckCount)

	return pushReturn + pushLcl + pushArg + pushThis + pushThat + repositionArg + repositionLcl + gotoFunc + returnLabel, 1
}

func (aw *AssemblyWriter) WriteInit() string {
	setSp := `@256
D=A
@0
M=D
`
	callInit, _ := aw.WriteCall(0)

	return setSp + callInit
}

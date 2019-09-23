package writer

import (
	"bufio"
	"fmt"
	"strconv"
)

type VMWriter struct {
	*bufio.Writer
}

func NewVMWriter(w *bufio.Writer) *VMWriter {
	return &VMWriter{w}
}

type Segment uint8

const (
	Const Segment = iota
	Arg
	Local
	Static
	This
	That
	Pointer
	Temp
)

func (s Segment) String() string {
	switch s {
	case Const:
		return "constant"
	case Arg:
		return "argument"
	case Local:
		return "local"
	case Static:
		return "static"
	case This:
		return "this"
	case That:
		return "that"
	case Pointer:
		return "pointer"
	case Temp:
		return "temp"
	default:
		return ""
	}
}

func (v *VMWriter) WritePush(seg Segment, index string) {
	idx, err := strconv.Atoi(index)
	if err != nil {
		panic(err)
	}
	v.WriteString(fmt.Sprintf("push %s %d\n", seg, idx))
}

func (v *VMWriter) WritePop(seg Segment, index int) {
	v.WriteString(fmt.Sprintf("pop %s %d\n", seg, index))
}

type Command string

const (
	Add Command = "add"
	Sub Command = "sub"
	Neg Command = "neg"
	Eq  Command = "eq"
	Gt  Command = "gt"
	Lt  Command = "lt"
	And Command = "and"
	Or  Command = "or"
	Not Command = "not"
)

func (v *VMWriter) WriteArithmetic(command Command) {
	v.WriteString(fmt.Sprintf("%s\n", command))
}

func (v *VMWriter) WriteLabel(label string) {
	v.WriteString(fmt.Sprintf("label %s\n", label))
}

func (v *VMWriter) WriteGoto(label string) {
	v.WriteString(fmt.Sprintf("goto %s\n", label))
}

func (v *VMWriter) WriteIf(label string) {
	v.WriteString(fmt.Sprintf("if-goto %s\n", label))
}

func (v *VMWriter) WriteCall(name string, nArgs int) {
	v.WriteString(fmt.Sprintf("call %s %d\n", name, nArgs))
}

func (v *VMWriter) WriteFunction(name string, nLocals int) {
	v.WriteString(fmt.Sprintf("function %s %d\n", name, nLocals))
}

func (v *VMWriter) WriteReturn() {
	v.WriteString("return\n")
}

package engine

import (
	"bufio"
	"fmt"
	"io"
	"strconv"

	"example.com/cache"
	"example.com/tokenizer"
	"example.com/writer"
)

type count struct {
	ifIdx     int
	whileIdx  int
	className string
}

type compilationEngine struct {
	scanner     *tokenizer.Scanner
	output      *writer.VMWriter
	symbolTable *cache.SymbolTable
	count       *count
}

func NewCompilationEngine(reader io.Reader, w *bufio.Writer) *compilationEngine {
	return &compilationEngine{
		tokenizer.NewScanner(reader),
		writer.NewVMWriter(w),
		cache.NewSymbolTable(),
		&count{0, 0, ""},
	}
}

func (c *compilationEngine) advance() {
	c.scanner.Advance()
}

func (c *compilationEngine) tokenValue() string {
	return c.scanner.Token.Value
}

func (c *compilationEngine) tokenCategory() tokenizer.Category {
	return c.scanner.Token.Category
}

func (c *compilationEngine) tokenIsType() bool {
	return c.scanner.Token.IsType()
}

func (c *compilationEngine) tokenIsOp() bool {
	return c.scanner.Token.IsOp()
}

func (c *compilationEngine) tokenIsUnaryOp() bool {
	return c.scanner.Token.IsUnaryOp()
}

func (c *compilationEngine) writeToken() {
	c.output.WriteString(fmt.Sprintf("<%s>%s</%s>\n", c.tokenCategory(), c.tokenValue(), c.tokenCategory()))
}

func (c *compilationEngine) compileIdentifier(defining bool, kind string, symbolType string) {
	if c.tokenCategory() != tokenizer.Identifier {
		panic(fmt.Sprintf("expected an identifer term, got %s", c.tokenValue()))
	}
	c.writeIdentifier(c.tokenValue(), defining, kind, symbolType)
	c.advance()
}

// TODO: need to split this up into many functions for better clarity
func (c *compilationEngine) writeIdentifier(name string, defining bool, kindName string, symbolType string) {
	var status string
	if defining {
		status = "define"
	} else {
		status = "use"
	}

	kind := cache.ParseKind(kindName)

	if kind != cache.None && defining {
		c.symbolTable.Define(name, symbolType, kind)
		index := c.symbolTable.IndexOf(name)
		// c.output.WriteString(fmt.Sprintf("<%s%s%d>%s</%s%s%d>\n", status, kind, index, name, status, kind, index))
		switch kind {
		case cache.Var:
			c.output.WritePush(writer.Const, "0")
			c.output.WritePop(writer.Local, index)
		}
		return
	}

	symbolKind := c.symbolTable.KindOf(name)
	if symbolKind != cache.None && kindName != "subroutine" && kindName != "class" {
		index := c.symbolTable.IndexOf(name)
		// c.output.WriteString(fmt.Sprintf("<%s%s%d>%s</%s%s%d>\n", status, symbolKind, index, name, status, symbolKind, index))
		switch symbolKind {
		case cache.Var:
			c.output.WritePush(writer.Local, strconv.Itoa(index))
		}
		return
	}

	c.output.WriteString(fmt.Sprintf("<%s%s>%s</%s%s>\n", status, kindName, name, status, kindName))
}

func (c *compilationEngine) writeTokenAndAdvance() {
	c.writeToken()
	c.advance()
}

func (c *compilationEngine) writeString(s string) {
	c.output.WriteString(s)
}

func (c *compilationEngine) compileTokenValue(tokenValue string) {
	if c.tokenValue() == tokenValue {
		// c.writeTokenAndAdvance()
		c.advance()
	} else {
		panic(fmt.Errorf(`expected token to have value "%s", got "%s"`, tokenValue, c.tokenValue()))
	}
}

func (c *compilationEngine) compileExpression() {
	// c.writeString("<expression>\n")
	operationStore := make([]string, 0)
	c.compileTerm()
	operationStore = c.handleMultipleTerms(operationStore)
	for i := len(operationStore) - 1; i >= 0; i-- {
		operation := operationStore[i]
		// need to convert the operation from the symbol to appropriate vm command
		c.writeOperation(operation)
	}
	// c.writeString("</expression>\n")
}

func (c *compilationEngine) writeOperation(operation string) {
	switch operation {
	case "+":
		c.output.WriteArithmetic(writer.Add)
	case "-":
		c.output.WriteArithmetic(writer.Sub)
	case "*":
		c.output.WriteCall("Math.multiply", 2)
	case "/":
		c.output.WriteCall("Math.divide", 2)
	case "&amp;":
		c.output.WriteArithmetic(writer.And)
	case "|":
		c.output.WriteArithmetic(writer.Or)
	case "&lt;":
		c.output.WriteArithmetic(writer.Lt)
	case "&gt;":
		c.output.WriteArithmetic(writer.Gt)
	case "=":
		c.output.WriteArithmetic(writer.Eq)
	}
}

func (c *compilationEngine) handleMultipleExpressions(nArgs int) int {
	if c.tokenValue() != "," {
		return nArgs
	}
	// c.writeTokenAndAdvance()
	c.advance()
	c.compileExpression()
	nArgs++
	nArgs = c.handleMultipleExpressions(nArgs)
	return nArgs
}

func (c *compilationEngine) compileFunctionCall() int {
	// if c.tokenValue() != "(" {
	// 	return 0
	// }
	nArgs := 0
	// c.compileTokenValue("(")
	c.advance()
	// c.writeString("<expressionList>\n")
	// handle empty expression list
	if c.tokenValue() == ")" {
		// c.writeString("</expressionList>\n")
		// c.compileTokenValue(")")
		c.advance()
		return nArgs
	}
	nArgs++
	c.compileExpression()
	nArgs = c.handleMultipleExpressions(nArgs)
	// c.writeString("</expressionList>\n")
	// c.compileTokenValue(")")
	c.advance()
	return nArgs
}

func (c *compilationEngine) compileMethodCall(kind cache.Kind, idx int) int {
	nArgs := 1
	c.advance() //(
	seg := convertKindToSegment(kind)
	c.output.WritePush(seg, strconv.Itoa(idx))
	if c.tokenValue() == ")" {
		// c.writeString("</expressionList>\n")
		// c.compileTokenValue(")")
		c.advance()
		return nArgs
	}
	nArgs++
	c.compileExpression()
	nArgs = c.handleMultipleExpressions(nArgs)
	c.advance()
	return nArgs
}

func (c *compilationEngine) handleArrayIndex() {
	if c.tokenValue() != "[" {
		return
	}
	c.compileTokenValue("[")
	c.compileExpression()
	c.compileTokenValue("]")
}

func (c *compilationEngine) handleIdentifierTerm() {
	// this identifier could be a class, subroutine, var, arg, static, field
	// can check if var, arg, static, field using symbol table
	// if none of above dependant on next token:
	// 		- if next token is . then it must be a class (could also be a var/arg/static/field but we already checked)
	//		- if next token is ( then it must be a subroutine call
	identifierName := c.tokenValue()
	c.advance()
	switch c.tokenValue() {
	case "(":
		// THIS IS A METHOD OF THE CURRENT OBJ
		// c.writeIdentifier(identifierName, false, "subroutine", "")
		objName := c.symbolTable.TypeOf("this")
		if objName == "" {
			objName = c.count.className
		}
		nArgs := c.compileMethodCall(cache.None, 0)
		c.output.WriteCall(fmt.Sprintf("%s.%s", objName, identifierName), nArgs)
	case ".":
		// if kind := c.symbolTable.KindOf(identifierName); kind != cache.None {
		// 	c.writeIdentifier(identifierName, false, kind.String(), "")
		// } else {
		// 	c.writeIdentifier(identifierName, false, "class", "")
		// }
		c.advance()
		functionName := c.tokenValue()
		var nArgs int
		objType := c.symbolTable.TypeOf(identifierName)
		c.advance()
		if objType == "" {
			nArgs = c.compileFunctionCall()
			c.output.WriteCall(fmt.Sprintf("%s.%s", identifierName, functionName), nArgs)
		} else {
			objKind := c.symbolTable.KindOf(identifierName)
			objIdx := c.symbolTable.IndexOf(identifierName)
			nArgs = c.compileMethodCall(objKind, objIdx)
			c.output.WriteCall(fmt.Sprintf("%s.%s", objType, functionName), nArgs)
		}
		// functionName, nArgs := c.compileObjectUse()
	case "[":
		kind := c.symbolTable.KindOf(identifierName)
		idx := c.symbolTable.IndexOf(identifierName)
		seg := convertKindToSegment(kind)
		c.output.WritePush(seg, strconv.Itoa(idx))
		c.advance() // advance to the index
		c.compileExpression()
		c.output.WriteArithmetic(writer.Add)
		c.output.WritePop(writer.Pointer, 1)
		c.output.WritePush(writer.That, strconv.Itoa(0))
		c.advance()
	default:
		// get the kind and index
		kind := c.symbolTable.KindOf(identifierName)
		index := c.symbolTable.IndexOf(identifierName)
		segment := convertKindToSegment(kind)
		c.output.WritePush(segment, strconv.Itoa(index))
	}
	/*
		OLD CODE

	*/
	// identifierName := c.tokenValue()
	// c.advance()
	// switch c.tokenValue() {
	// case "[":
	// 	c.writeIdentifier(identifierName, false, "", "")
	// 	c.handleArrayIndex()
	// case "(":
	// 	c.writeIdentifier(identifierName, false, "subroutine", "")
	// 	c.compileFunctionCall()
	// case ".":
	// 	// PROBLEM: this will always be labelled as a class when it could be a var
	// 	if symbolType := c.symbolTable.KindOf(identifierName); symbolType != cache.None {
	// 		c.writeIdentifier(identifierName, false, symbolType.String(), "")
	// 	} else {
	// 		c.writeIdentifier(identifierName, false, "class", "")
	// 	}
	// 	c.compileObjectUse()
	// default:
	// 	c.writeIdentifier(identifierName, false, "", "")
	// }
}

func (c *compilationEngine) handleExpressionBrackets() {
	c.compileTokenValue("(")
	c.compileExpression()
	c.compileTokenValue(")")
}

func (c *compilationEngine) compileTerm() {
	// c.writeString("<term>\n")
	if c.tokenCategory() == tokenizer.IntConst {
		c.output.WritePush(writer.Const, c.tokenValue())
		c.advance()
	} else if c.tokenCategory() == tokenizer.StringConst {
		// c.writeTokenAndAdvance()
		str := c.tokenValue()
		length := len(c.tokenValue())
		c.output.WritePush(writer.Const, strconv.Itoa(length))
		c.output.WriteCall("String.new", 1)
		for _, char := range str {
			c.output.WritePush(writer.Const, strconv.Itoa(int(char)))
			c.output.WriteCall("String.appendChar", 2)
		}
		c.advance()
	} else if c.tokenCategory() == tokenizer.Keyword {
		if c.tokenValue() == "true" {
			c.output.WritePush(writer.Const, strconv.Itoa(1))
			c.output.WriteArithmetic(writer.Neg)
		} else if c.tokenValue() == "false" {
			c.output.WritePush(writer.Const, strconv.Itoa(0))
		} else if c.tokenValue() == "this" {
			c.output.WritePush(writer.Pointer, strconv.Itoa(0))
		} else if c.tokenValue() == "null" {
			c.output.WritePush(writer.Const, strconv.Itoa(0))
		}
		c.advance()
	} else if c.tokenCategory() == tokenizer.Identifier {
		c.handleIdentifierTerm()
	} else if c.tokenValue() == "(" {
		c.handleExpressionBrackets()
	} else if c.tokenIsUnaryOp() {
		// c.writeTokenAndAdvance()
		op := c.tokenValue()
		c.advance()
		c.compileTerm()
		switch op {
		case "-":
			c.output.WriteArithmetic(writer.Neg)
		case "~":
			c.output.WriteArithmetic(writer.Not)
		}
	} else {
		panic(fmt.Errorf("invalid term grammar %s is not valid for a term", c.tokenValue()))
	}
	// c.writeString("</term>\n")
}

func (c *compilationEngine) handleMultipleTerms(operationStore []string) []string {
	if !c.tokenIsOp() {
		return operationStore
	}
	operationStore = append(operationStore, c.tokenValue())
	// c.writeTokenAndAdvance()
	c.advance()
	c.compileTerm()
	c.handleMultipleTerms(operationStore)
	return operationStore
}

func (c *compilationEngine) compileLet() {
	if c.tokenValue() != "let" {
		return
	}
	c.advance()
	// get index and kind of the variable we are assigning to
	kind := c.symbolTable.KindOf(c.tokenValue())
	index := c.symbolTable.IndexOf(c.tokenValue())
	c.advance()
	// need to handle arrays here
	isArrayAssignment := c.tokenValue() == "["
	if isArrayAssignment {
		seg := convertKindToSegment(kind)
		c.output.WritePush(seg, strconv.Itoa(index))
		c.advance() // advance to the index
		c.compileExpression()
		c.output.WriteArithmetic(writer.Add)
		c.advance()
	}
	c.advance()
	// complete operation after the equals
	c.compileExpression()
	// pop the result back to the index/variable found
	if isArrayAssignment {
		c.output.WritePop(writer.Temp, 0)
		c.output.WritePop(writer.Pointer, 1)
		c.output.WritePush(writer.Temp, strconv.Itoa(0))
		c.output.WritePop(writer.That, 0)
	} else {
		segment := convertKindToSegment(kind)
		c.output.WritePop(segment, index)
	}
	c.advance()

	// c.writeString("<letStatement>\n")
	// c.writeTokenAndAdvance()
	// c.advance()
	// c.compileIdentifier(false, "", "")
	// c.handleArrayIndex()
	// c.compileTokenValue("=")
	// c.compileExpression()
	// c.compileTokenValue(";")
	// c.writeString("</letStatement>\n")
}

func convertKindToSegment(kind cache.Kind) writer.Segment {
	switch kind {
	case cache.Var:
		return writer.Local
	case cache.Arg:
		return writer.Arg
	case cache.Static:
		return writer.Static
	case cache.Field:
		return writer.This
	default:
		return writer.Pointer
	}
}

func (c *compilationEngine) compileElse() {
	if c.tokenValue() != "else" {
		return
	}
	// c.writeTokenAndAdvance()
	c.advance()

	c.handleStatements()
}

func (c *compilationEngine) compileIf() {
	if c.tokenValue() != "if" {
		return
	}
	// c.writeString("<ifStatement>\n")
	c.count.ifIdx++
	ifIdxStr := strconv.Itoa(c.count.ifIdx)
	c.advance()
	// compute not condition
	c.handleExpressionBrackets()
	c.output.WriteArithmetic(writer.Not)
	// if not condition jump to else label
	c.output.WriteIf("else" + ifIdxStr)
	// compute code inside if
	c.handleStatements()
	// jump to end of statement label
	c.output.WriteGoto("end" + ifIdxStr)
	// write else label
	c.output.WriteLabel("else" + ifIdxStr)
	if c.tokenValue() == "else" {
		c.advance()
		c.handleStatements()
	}
	// compute code inside else
	// end of statement label
	c.output.WriteLabel("end" + ifIdxStr)

	// c.writeTokenAndAdvance()
	// c.advance()
	// c.handleExpressionStatements()
	// c.compileElse()
	// c.writeString("</ifStatement>\n")
}

func (c *compilationEngine) compileWhile() {
	if c.tokenValue() != "while" {
		return
	}
	// c.writeString("<whileStatement>\n")
	c.count.whileIdx++
	whileIdxStr := strconv.Itoa(c.count.whileIdx)
	// write label here
	c.output.WriteLabel("while" + whileIdxStr)
	c.advance()
	// check if conditions met (if not jump to end label)
	c.handleExpressionBrackets()
	c.output.WriteArithmetic(writer.Not)
	c.output.WriteIf("endWhile" + whileIdxStr)
	// code inside while loop
	c.handleStatements()
	// goto start label
	c.output.WriteGoto("while" + whileIdxStr)
	// label for end of loop
	c.output.WriteLabel("endWhile" + whileIdxStr)

	// c.writeTokenAndAdvance()
	// c.advance()
	// c.handleExpressionStatements()
	// c.writeString("</whileStatement>\n")
}

func (c *compilationEngine) compileDo() {
	if c.tokenValue() != "do" {
		return
	}
	// c.writeString("<doStatement>\n")
	// c.writeTokenAndAdvance()
	c.advance()
	c.compileTerm()
	c.output.WritePop(writer.Temp, 0)
	c.advance()
	// identifierName := c.tokenValue()
	// c.advance()
	// switch c.tokenValue() {
	// case "(":
	// 	// c.writeIdentifier(identifierName, false, "subroutine", "")
	// 	nArgs := c.compileFunctionCall()
	// 	c.output.WriteCall(identifierName, nArgs)
	// case ".":
	// 	// if kind := c.symbolTable.KindOf(identifierName); kind != cache.None {
	// 	// 	c.writeIdentifier(identifierName, false, kind.String(), "")
	// 	// } else {
	// 	// 	c.writeIdentifier(identifierName, false, "class", "")
	// 	// }
	// 	functionName, nArgs := c.compileObjectUse()
	// 	c.output.WriteCall(fmt.Sprintf("%s.%s", identifierName, functionName), nArgs)
	// }
	// c.compileTokenValue(";")
	// c.advance()
	// c.writeString("</doStatement>\n")
}

func (c *compilationEngine) compileReturn() {
	if c.tokenValue() != "return" {
		return
	}
	// c.writeString("<returnStatement>\n")
	// c.writeTokenAndAdvance()
	c.advance()
	// handle empty return
	if c.tokenValue() == ";" {
		// c.writeTokenAndAdvance()
		c.output.WritePush(writer.Const, strconv.Itoa(0))
		c.output.WriteReturn()
		c.advance()
		// c.writeString("</returnStatement>\n")
		return
	}
	c.compileExpression()
	c.output.WriteReturn()
	c.compileTokenValue(";")
	// c.writeString("</returnStatement>\n")
}

func (c *compilationEngine) isTokenStatement() bool {
	statementTokens := []string{
		"let",
		"if",
		"while",
		"do",
		"return",
	}
	for _, token := range statementTokens {
		if c.tokenValue() == token {
			return true
		}
	}
	return false
}

func (c *compilationEngine) compileStatement() {
	if !c.isTokenStatement() {
		return
	}
	c.compileLet()
	c.compileIf()
	c.compileWhile()
	c.compileDo()
	c.compileReturn()
	c.compileStatement()
}

func (c *compilationEngine) compileStatements() {
	// c.writeString("<statements>\n")
	c.compileStatement()
	// c.writeString("</statements>\n")
}

func (c *compilationEngine) handleStatements() {
	c.compileTokenValue("{")
	c.compileStatements()
	c.compileTokenValue("}")
}

func (c *compilationEngine) handleExpressionStatements() {
	c.handleExpressionBrackets()
	c.handleStatements()
}

func (c *compilationEngine) compileTokenIsTypeOrVoid() {
	if c.tokenIsType() || c.tokenValue() == "void" {
		// c.writeTokenAndAdvance()
		c.advance()
	} else {
		panic(fmt.Errorf(`expected either a type according to jack grammar or void, got: "%s"`, c.tokenValue()))
	}
}

func (c *compilationEngine) compileParameterList() {
	// c.writeString("<parameterList>\n")
	// handle empty parameter list
	if c.tokenValue() == ")" {
		// c.writeString("</parameterList>\n")
		return
	}
	symbolType := c.tokenValue()
	c.compileTokenIsType()
	// need to add the variable to the argument symbol table
	c.symbolTable.Define(c.tokenValue(), symbolType, cache.Arg)
	// c.compileIdentifier(true, "arg", symbolType)
	c.advance()
	c.handleMultipleParameters()
	// c.writeString("</parameterList>\n")
}

func (c *compilationEngine) handleMultipleParameters() {
	if c.tokenValue() != "," {
		return
	}
	// c.writeTokenAndAdvance()
	c.advance()
	symbolType := c.tokenValue()
	c.compileTokenIsType()
	// c.compileIdentifier(true, "arg", symbolType)
	c.symbolTable.Define(c.tokenValue(), symbolType, cache.Arg)
	c.advance()
	c.handleMultipleParameters()
	return
}

func (c *compilationEngine) compileTokenIsType() {
	if c.tokenIsType() {
		// c.writeTokenAndAdvance()
		c.advance()
	} else {
		panic(fmt.Errorf(`expected a token that is a type according to jack grammar, got "%s"`, c.tokenValue()))
	}
}

func (c *compilationEngine) compileMultipleVarDecs(kind string, symbolType string) {
	if c.tokenValue() != "," {
		return
	}
	// c.writeTokenAndAdvance()
	c.advance()
	c.compileIdentifier(true, kind, symbolType)
	c.compileMultipleVarDecs(kind, symbolType)
}

func (c *compilationEngine) handleMultipleSubroutineVarDecs(symbolType string, buffer string, nLocals int) (string, int) {
	if c.tokenValue() != "," {
		return buffer, nLocals
	}
	nLocals++
	c.advance()
	varName := c.tokenValue()
	c.symbolTable.Define(varName, symbolType, cache.Var)
	buffer += "push constant 0\n"
	idx := c.symbolTable.IndexOf(varName)
	buffer += "pop local " + strconv.Itoa(idx) + "\n"
	c.advance()
	buffer, nLocals = c.handleMultipleSubroutineVarDecs(symbolType, buffer, nLocals)
	return buffer, nLocals
}

func (c *compilationEngine) compileVarDec(buffer string, nLocals int) (string, int) {
	if c.tokenValue() != "var" {
		return buffer, nLocals
	}
	// c.writeString("<varDec>\n")
	// c.writeTokenAndAdvance()
	nLocals++
	c.advance()
	symbolType := c.tokenValue()
	// c.compileTokenIsType()
	c.advance()
	varName := c.tokenValue()
	c.symbolTable.Define(varName, symbolType, cache.Var)
	buffer += "push constant 0\n"
	idx := c.symbolTable.IndexOf(varName)
	buffer += "pop local " + strconv.Itoa(idx) + "\n"
	c.advance()
	// c.compileIdentifier(true, "var", symbolType)
	buffer, nLocals = c.handleMultipleSubroutineVarDecs(symbolType, buffer, nLocals)
	c.compileTokenValue(";")
	// c.writeString("</varDec>\n")
	buffer, nLocals = c.compileVarDec(buffer, nLocals)
	return buffer, nLocals
}

func (c *compilationEngine) compileSubroutine(className string) {
	if c.tokenValue() != "constructor" && c.tokenValue() != "function" && c.tokenValue() != "method" {
		return
	}
	c.symbolTable.StartSubroutine()
	if c.tokenValue() == "function" {
		// c.writeString("<subroutineDec>\n")
		// c.writeTokenAndAdvance()
		nLocals := 0
		c.advance()
		// returnType := c.tokenValue()
		c.compileTokenIsTypeOrVoid()
		// c.compileIdentifier(true, "subroutine", "")
		subroutineName := c.tokenValue()
		c.advance()
		c.compileTokenValue("(")
		c.compileParameterList()
		c.compileTokenValue(")")
		c.compileTokenValue("{")
		varDecBuffer := ""
		varDecBuffer, nLocals = c.compileVarDec(varDecBuffer, nLocals)
		c.output.WriteFunction(fmt.Sprintf("%s.%s", className, subroutineName), nLocals)
		c.output.WriteString(varDecBuffer)
		c.compileStatements()
		c.compileTokenValue("}")
		// if returnType == "void" {
		// 	c.output.WritePush(writer.Const, "0")
		// 	c.output.WriteReturn()
		// }
		// c.writeString("</subroutineDec>\n")
	} else if c.tokenValue() == "constructor" {
		// c.symbolTable.Define("this", className, cache.Arg)
		nLocals := 0
		c.advance()
		// returnType := c.tokenValue()
		c.compileTokenIsTypeOrVoid()
		// c.compileIdentifier(true, "subroutine", "")
		subroutineName := c.tokenValue()
		c.advance()
		c.compileTokenValue("(")
		c.compileParameterList()
		c.compileTokenValue(")")
		c.compileTokenValue("{")
		varDecBuffer := ""
		varDecBuffer, nLocals = c.compileVarDec(varDecBuffer, nLocals)
		c.output.WriteFunction(fmt.Sprintf("%s.%s", className, subroutineName), nLocals)
		c.output.WriteString(varDecBuffer)
		numOfFieldVars := c.symbolTable.FieldIndex
		c.output.WritePush(writer.Const, strconv.Itoa(numOfFieldVars))
		c.output.WriteCall("Memory.alloc", 1)
		c.output.WritePop(writer.Pointer, 0)
		c.compileStatements()
		c.compileTokenValue("}")
	} else if c.tokenValue() == "method" {
		c.symbolTable.Define("this", className, cache.Arg)
		nLocals := 0
		c.advance()
		// returnType := c.tokenValue()
		c.compileTokenIsTypeOrVoid()
		// c.compileIdentifier(true, "subroutine", "")
		subroutineName := c.tokenValue()
		c.advance()
		c.compileTokenValue("(")
		c.compileParameterList()
		c.compileTokenValue(")")
		c.compileTokenValue("{")
		varDecBuffer := ""
		varDecBuffer, nLocals = c.compileVarDec(varDecBuffer, nLocals)
		c.output.WriteFunction(fmt.Sprintf("%s.%s", className, subroutineName), nLocals)
		c.output.WritePush(writer.Arg, strconv.Itoa(0))
		c.output.WritePop(writer.Pointer, 0)
		c.output.WriteString(varDecBuffer)
		c.compileStatements()
		c.compileTokenValue("}")
	}
	c.compileSubroutine(className)
}

func (c *compilationEngine) compileClassVarDec() {
	if c.tokenValue() != "static" && c.tokenValue() != "field" {
		return
	}
	// c.writeString("<classVarDec>\n")
	kind := c.tokenValue()
	// c.writeTokenAndAdvance()
	c.advance()
	symbolType := c.tokenValue()
	c.compileTokenIsType()
	c.compileIdentifier(true, kind, symbolType)
	c.compileMultipleVarDecs(kind, symbolType)
	c.compileTokenValue(";")
	// c.writeString("</classVarDec>\n")
	c.compileClassVarDec()
}

func (c *compilationEngine) compileFinalToken() {
	if c.tokenValue() == "}" {
		// c.writeToken()
	} else {
		panic(fmt.Errorf(`expected token "}" as the final token, got "%s"`, c.tokenValue()))
	}
}

func (c *compilationEngine) CompileClass() {
	// c.writeString("<class>\n")
	c.advance()
	c.compileTokenValue("class")
	// c.compileIdentifier(true, "class", "")
	c.count.className = c.tokenValue()
	c.advance()
	c.compileTokenValue("{")
	// c.advance()
	c.compileClassVarDec()
	c.compileSubroutine(c.count.className)
	c.compileFinalToken()
	// c.writeString("</class>\n")
}

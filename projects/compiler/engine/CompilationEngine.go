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

type compilationEngine struct {
	scanner     *tokenizer.Scanner
	output      *writer.VMWriter
	symbolTable *cache.SymbolTable
}

func NewCompilationEngine(reader io.Reader, w *bufio.Writer) *compilationEngine {
	return &compilationEngine{
		tokenizer.NewScanner(reader),
		writer.NewVMWriter(w),
		cache.NewSymbolTable(),
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

func (c *compilationEngine) compileObjectUse() (string, int) {
	// if c.tokenValue() != "." {
	// 	return
	// }
	// c.compileTokenValue(".")
	// c.compileIdentifier(false, "subroutine", "")
	c.advance()
	functionName := c.tokenValue()
	c.advance()
	nArgs := c.compileFunctionCall()
	return functionName, nArgs
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
		// c.writeIdentifier(identifierName, false, "subroutine", "")
		nArgs := c.compileFunctionCall()
		c.output.WriteCall(identifierName, nArgs)
	case ".":
		// if kind := c.symbolTable.KindOf(identifierName); kind != cache.None {
		// 	c.writeIdentifier(identifierName, false, kind.String(), "")
		// } else {
		// 	c.writeIdentifier(identifierName, false, "class", "")
		// }
		functionName, nArgs := c.compileObjectUse()
		c.output.WriteCall(fmt.Sprintf("%s.%s", identifierName, functionName), nArgs)
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
		c.advance()
	} else if c.tokenCategory() == tokenizer.Keyword {
		if c.tokenValue() == "true" {
			c.output.WritePush(writer.Const, strconv.Itoa(1))
			c.output.WriteArithmetic(writer.Neg)
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
	c.advance()
	// complete operation after the equals
	c.compileExpression()
	// pop the result back to the index/variable found
	segment := convertKindToSegment(kind)
	c.output.WritePop(segment, index)
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
	default:
		return writer.None
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
	// c.writeTokenAndAdvance()
	c.advance()
	c.handleExpressionStatements()
	c.compileElse()
	// c.writeString("</ifStatement>\n")
}

func (c *compilationEngine) compileWhile() {
	if c.tokenValue() != "while" {
		return
	}
	// c.writeString("<whileStatement>\n")
	// c.writeTokenAndAdvance()
	c.advance()
	c.handleExpressionStatements()
	// c.writeString("</whileStatement>\n")
}

func (c *compilationEngine) compileDo() {
	if c.tokenValue() != "do" {
		return
	}
	// c.writeString("<doStatement>\n")
	// c.writeTokenAndAdvance()
	c.advance()
	identifierName := c.tokenValue()
	c.advance()
	switch c.tokenValue() {
	case "(":
		// c.writeIdentifier(identifierName, false, "subroutine", "")
		nArgs := c.compileFunctionCall()
		c.output.WriteCall(identifierName, nArgs)
	case ".":
		// if kind := c.symbolTable.KindOf(identifierName); kind != cache.None {
		// 	c.writeIdentifier(identifierName, false, kind.String(), "")
		// } else {
		// 	c.writeIdentifier(identifierName, false, "class", "")
		// }
		functionName, nArgs := c.compileObjectUse()
		c.output.WriteCall(fmt.Sprintf("%s.%s", identifierName, functionName), nArgs)
	}
	// c.compileTokenValue(";")
	c.advance()
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
		c.advance()
		// c.writeString("</returnStatement>\n")
		return
	}
	c.compileExpression()
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

func (c *compilationEngine) compileParameterList(nLocals int) int {
	// c.writeString("<parameterList>\n")
	// handle empty parameter list
	if c.tokenValue() == ")" {
		// c.writeString("</parameterList>\n")
		return nLocals
	}
	nLocals++
	symbolType := c.tokenValue()
	c.compileTokenIsType()
	// c.compileIdentifier(true, "arg", symbolType)
	c.advance()
	nLocals = c.handleMultipleParameters(symbolType, nLocals)
	// c.writeString("</parameterList>\n")
	return nLocals
}

func (c *compilationEngine) handleMultipleParameters(symbolType string, nLocals int) int {
	if c.tokenValue() != "," {
		return nLocals
	}
	nLocals++
	// c.writeTokenAndAdvance()
	c.advance()
	c.compileTokenIsType()
	c.compileIdentifier(true, "arg", symbolType)
	nLocals = c.handleMultipleParameters(symbolType, nLocals)
	return nLocals
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

func (c *compilationEngine) compileVarDec() {
	if c.tokenValue() != "var" {
		return
	}
	// c.writeString("<varDec>\n")
	// c.writeTokenAndAdvance()
	c.advance()
	symbolType := c.tokenValue()
	c.compileTokenIsType()
	c.compileIdentifier(true, "var", symbolType)
	c.compileMultipleVarDecs("var", symbolType)
	c.compileTokenValue(";")
	// c.writeString("</varDec>\n")
	c.compileVarDec()
}

func (c *compilationEngine) compileSubroutineBody() {
	// c.writeString("<subroutineBody>\n")
	c.compileTokenValue("{")
	c.compileVarDec()
	c.compileStatements()
	c.compileTokenValue("}")
	// c.writeString("</subroutineBody>\n")
}

func (c *compilationEngine) compileSubroutine(className string) {
	if c.tokenValue() != "constructor" && c.tokenValue() != "function" && c.tokenValue() != "method" {
		return
	}
	c.symbolTable.StartSubroutine()
	// c.writeString("<subroutineDec>\n")
	// c.writeTokenAndAdvance()
	var nLocals int
	if c.tokenValue() == "method" {
		nLocals = 1
	} else {
		nLocals = 0
	}
	c.advance()
	returnType := c.tokenValue()
	c.compileTokenIsTypeOrVoid()
	// c.compileIdentifier(true, "subroutine", "")
	subroutineName := c.tokenValue()
	c.advance()
	c.compileTokenValue("(")
	nLocals = c.compileParameterList(nLocals)
	c.compileTokenValue(")")
	c.output.WriteFunction(fmt.Sprintf("%s.%s", className, subroutineName), nLocals)
	c.compileSubroutineBody()
	if returnType == "void" {
		c.output.WritePush(writer.Const, "0")
		c.output.WriteReturn()
	}
	// c.writeString("</subroutineDec>\n")
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
	className := c.tokenValue()
	c.advance()
	c.compileTokenValue("{")
	// c.advance()
	c.compileClassVarDec()
	c.compileSubroutine(className)
	c.compileFinalToken()
	// c.writeString("</class>\n")
}

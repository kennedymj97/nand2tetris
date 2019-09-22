package JackCompiler

import (
	"bufio"
	"fmt"
	"io"
)

type compilationEngine struct {
	scanner     *scanner
	output      *bufio.Writer
	symbolTable *symbolTable
}

func newCompilationEngine(reader io.Reader, writer *bufio.Writer) *compilationEngine {
	return &compilationEngine{
		newScanner(reader),
		writer,
		newSymbolTable(),
	}
}

func (c *compilationEngine) advance() {
	c.scanner.advance()
}

func (c *compilationEngine) tokenValue() string {
	return c.scanner.token.value
}

func (c *compilationEngine) tokenCategory() category {
	return c.scanner.token.category
}

func (c *compilationEngine) tokenIsType() bool {
	return c.scanner.token.isType()
}

func (c *compilationEngine) tokenIsOp() bool {
	return c.scanner.token.isOp()
}

func (c *compilationEngine) tokenIsUnaryOp() bool {
	return c.scanner.token.isUnaryOp()
}

func (c *compilationEngine) writeToken() {
	var label string
	switch c.tokenCategory() {
	case keyword:
		label = "keyword"
	case symbol:
		label = "symbol"
	case stringConst:
		label = "stringConstant"
	case intConst:
		label = "integerConstant"
	case identifier:
		label = "identifier"
	}
	c.output.WriteString(fmt.Sprintf("<%s>%s</%s>\n", label, c.tokenValue(), label))
}

func (c *compilationEngine) compileIdentifier(defining bool, kind string, symbolType string) {
	if c.tokenCategory() != identifier {
		panic(fmt.Sprintf("expected an identifer term, got %s", c.tokenValue()))
	}
	c.writeIdentifier(c.tokenValue(), defining, kind, symbolType)
	c.advance()
}

// TODO: need to split this up into many functions for better clarity
func (c *compilationEngine) writeIdentifier(name string, defining bool, kind string, symbolType string) {
	var status string
	if defining {
		status = "define"
	} else {
		status = "use"
	}

	if isTableVar(kind) && defining {
		c.symbolTable.define(name, symbolType, kind)
		index := c.symbolTable.indexOf(name)
		c.output.WriteString(fmt.Sprintf("<%s%s%d>%s</%s%s%d>\n", status, kind, index, name, status, kind, index))
		return
	}

	symbolKind := c.symbolTable.kindOf(name)
	if isTableVar(kind) && symbolKind != NONE {
		index := c.symbolTable.indexOf(name)
		c.output.WriteString(fmt.Sprintf("<%s%s%d>%s</%s%s%d>\n", status, symbolKind, index, name, status, symbolKind, index))
		return
	}

	c.output.WriteString(fmt.Sprintf("<%s%s>%s</%s%s>\n", status, kind, name, status, kind))
}

func isTableVar(kind string) bool {
	if kind != "class" && kind != "method" && kind != "function" && kind != "subroutine" && kind != "constructor" {
		return true
	}
	return false
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
		c.writeTokenAndAdvance()
	} else {
		panic(fmt.Errorf(`expected token to have value "%s", got "%s"`, tokenValue, c.tokenValue()))
	}
}

func (c *compilationEngine) compileExpression() {
	c.writeString("<expression>\n")
	c.compileTerm()
	c.handleMultipleTerms()
	c.writeString("</expression>\n")
}

func (c *compilationEngine) handleMultipleExpressions() {
	if c.tokenValue() != "," {
		return
	}
	c.writeTokenAndAdvance()
	c.compileExpression()
	c.handleMultipleExpressions()
}

func (c *compilationEngine) compileFunctionCall() {
	if c.tokenValue() != "(" {
		return
	}
	c.compileTokenValue("(")
	c.writeString("<expressionList>\n")
	// handle empty expression list
	if c.tokenValue() == ")" {
		c.writeString("</expressionList>\n")
		c.compileTokenValue(")")
		return
	}
	c.compileExpression()
	c.handleMultipleExpressions()
	c.writeString("</expressionList>\n")
	c.compileTokenValue(")")
}

func (c *compilationEngine) compileObjectUse() {
	if c.tokenValue() != "." {
		return
	}
	c.compileTokenValue(".")
	c.compileIdentifier(false, "subroutine", "")
	c.compileFunctionCall()
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
	case "[":
		c.writeIdentifier(identifierName, false, "", "")
		c.handleArrayIndex()
	case "(":
		c.writeIdentifier(identifierName, false, "subroutine", "")
		c.compileFunctionCall()
	case ".":
		// PROBLEM: this will always be labelled as a class when it could be a var
		if symbolType := c.symbolTable.kindOf(identifierName); symbolType != NONE {
			c.writeIdentifier(identifierName, false, symbolType.String(), "")
		} else {
			c.writeIdentifier(identifierName, false, "class", "")
		}
		c.compileObjectUse()
	default:
		c.writeIdentifier(identifierName, false, "", "")
	}
}

func (c *compilationEngine) handleExpressionBrackets() {
	c.compileTokenValue("(")
	c.compileExpression()
	c.compileTokenValue(")")
}

func (c *compilationEngine) compileTerm() {
	c.writeString("<term>\n")
	if c.tokenCategory() == intConst || c.tokenCategory() == stringConst || c.tokenCategory() == keyword {
		c.writeTokenAndAdvance()
	} else if c.tokenCategory() == identifier {
		c.handleIdentifierTerm()
	} else if c.tokenValue() == "(" {
		c.handleExpressionBrackets()
	} else if c.tokenIsUnaryOp() {
		c.writeTokenAndAdvance()
		c.compileTerm()
	} else {
		panic(fmt.Errorf("invalid term grammar %s is not valid for a term", c.tokenValue()))
	}
	c.writeString("</term>\n")
}

func (c *compilationEngine) handleMultipleTerms() {
	if !c.tokenIsOp() {
		return
	}
	c.writeTokenAndAdvance()
	c.compileTerm()
	c.handleMultipleTerms()
}

func (c *compilationEngine) compileLet() {
	if c.tokenValue() != "let" {
		return
	}
	c.writeString("<letStatement>\n")
	c.writeTokenAndAdvance()
	c.compileIdentifier(false, "", "")
	c.handleArrayIndex()
	c.compileTokenValue("=")
	c.compileExpression()
	c.compileTokenValue(";")
	c.writeString("</letStatement>\n")
}

func (c *compilationEngine) compileElse() {
	if c.tokenValue() != "else" {
		return
	}
	c.writeTokenAndAdvance()
	c.handleStatements()
}

func (c *compilationEngine) compileIf() {
	if c.tokenValue() != "if" {
		return
	}
	c.writeString("<ifStatement>\n")
	c.writeTokenAndAdvance()
	c.handleExpressionStatements()
	c.compileElse()
	c.writeString("</ifStatement>\n")
}

func (c *compilationEngine) compileWhile() {
	if c.tokenValue() != "while" {
		return
	}
	c.writeString("<whileStatement>\n")
	c.writeTokenAndAdvance()
	c.handleExpressionStatements()
	c.writeString("</whileStatement>\n")
}

func (c *compilationEngine) compileDo() {
	if c.tokenValue() != "do" {
		return
	}
	c.writeString("<doStatement>\n")
	c.writeTokenAndAdvance()
	identifierName := c.tokenValue()
	c.advance()
	switch c.tokenValue() {
	case "(":
		c.writeIdentifier(identifierName, false, "subroutine", "")
		c.compileFunctionCall()
	case ".":
		if symbolType := c.symbolTable.kindOf(identifierName); symbolType != NONE {
			c.writeIdentifier(identifierName, false, symbolType.String(), "")
		} else {
			c.writeIdentifier(identifierName, false, "class", "")
		}
		c.compileObjectUse()
	}
	c.compileTokenValue(";")
	c.writeString("</doStatement>\n")
}

func (c *compilationEngine) compileReturn() {
	if c.tokenValue() != "return" {
		return
	}
	c.writeString("<returnStatement>\n")
	c.writeTokenAndAdvance()
	// handle empty return
	if c.tokenValue() == ";" {
		c.writeTokenAndAdvance()
		c.writeString("</returnStatement>\n")
		return
	}
	c.compileExpression()
	c.compileTokenValue(";")
	c.writeString("</returnStatement>\n")
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
	c.writeString("<statements>\n")
	c.compileStatement()
	c.writeString("</statements>\n")
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
		c.writeTokenAndAdvance()
	} else {
		panic(fmt.Errorf(`expected either a type according to jack grammar or void, got: "%s"`, c.tokenValue()))
	}
}

func (c *compilationEngine) compileParameterList() {
	c.writeString("<parameterList>\n")
	// handle empty parameter list
	if c.tokenValue() == ")" {
		c.writeString("</parameterList>\n")
		return
	}
	symbolType := c.tokenValue()
	c.compileTokenIsType()
	c.compileIdentifier(true, "arg", symbolType)
	c.handleMultipleParameters(symbolType)
	c.writeString("</parameterList>\n")
}

func (c *compilationEngine) handleMultipleParameters(symbolType string) {
	if c.tokenValue() != "," {
		return
	}
	c.writeTokenAndAdvance()
	c.compileTokenIsType()
	c.compileIdentifier(true, "arg", symbolType)
	c.handleMultipleParameters(symbolType)
}

func (c *compilationEngine) compileTokenIsType() {
	if c.tokenIsType() {
		c.writeTokenAndAdvance()
	} else {
		panic(fmt.Errorf(`expected a token that is a type according to jack grammar, got "%s"`, c.tokenValue()))
	}
}

func (c *compilationEngine) compileMultipleVarDecs(kind string, symbolType string) {
	if c.tokenValue() != "," {
		return
	}
	c.writeTokenAndAdvance()
	c.compileIdentifier(true, kind, symbolType)
	c.compileMultipleVarDecs(kind, symbolType)
}

func (c *compilationEngine) compileVarDec() {
	if c.tokenValue() != "var" {
		return
	}
	c.writeString("<varDec>\n")
	c.writeTokenAndAdvance()
	symbolType := c.tokenValue()
	c.compileTokenIsType()
	c.compileIdentifier(true, "var", symbolType)
	c.compileMultipleVarDecs("var", symbolType)
	c.compileTokenValue(";")
	c.writeString("</varDec>\n")
	c.compileVarDec()
}

func (c *compilationEngine) compileSubroutineBody() {
	c.writeString("<subroutineBody>\n")
	c.compileTokenValue("{")
	c.compileVarDec()
	c.compileStatements()
	c.compileTokenValue("}")
	c.writeString("</subroutineBody>\n")
}

func (c *compilationEngine) compileSubroutine() {
	if c.tokenValue() != "constructor" && c.tokenValue() != "function" && c.tokenValue() != "method" {
		return
	}
	c.symbolTable.startSubroutine()
	c.writeString("<subroutineDec>\n")
	c.writeTokenAndAdvance()
	c.compileTokenIsTypeOrVoid()
	c.compileIdentifier(true, "subroutine", "")
	c.compileTokenValue("(")
	c.compileParameterList()
	c.compileTokenValue(")")
	c.compileSubroutineBody()
	c.writeString("</subroutineDec>\n")
	c.compileSubroutine()
}

func (c *compilationEngine) compileClassVarDec() {
	if c.tokenValue() != "static" && c.tokenValue() != "field" {
		return
	}
	c.writeString("<classVarDec>\n")
	kind := c.tokenValue()
	c.writeTokenAndAdvance()
	symbolType := c.tokenValue()
	c.compileTokenIsType()
	c.compileIdentifier(true, kind, symbolType)
	c.compileMultipleVarDecs(kind, symbolType)
	c.compileTokenValue(";")
	c.writeString("</classVarDec>\n")
	c.compileClassVarDec()
}

func (c *compilationEngine) compileFinalToken() {
	if c.tokenValue() == "}" {
		c.writeToken()
	} else {
		panic(fmt.Errorf(`expected token "}" as the final token, got "%s"`, c.tokenValue()))
	}
}

func (c *compilationEngine) compileClass() {
	c.writeString("<class>\n")
	c.advance()
	c.compileTokenValue("class")
	c.compileIdentifier(true, "class", "")
	c.compileTokenValue("{")
	c.compileClassVarDec()
	c.compileSubroutine()
	c.compileFinalToken()
	c.writeString("</class>\n")
}

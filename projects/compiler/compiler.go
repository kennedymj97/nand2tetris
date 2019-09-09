package compiler

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Compiler consists of a scanner that steps through the tokens in a jack file,
// a bufio writer to produce the parsed file and a token to handle tokens.
type Compiler struct {
	scanner *scanner
	output  *bufio.Writer
}

// NewCompiler creates an Compiler
func NewCompiler(reader io.Reader, writer *bufio.Writer) *Compiler {
	scanner := newScanner(reader)
	return &Compiler{
		scanner,
		writer,
	}
}

func (c *Compiler) advance() {
	c.scanner.advance()
}

func (c *Compiler) tokenValue() string {
	return c.scanner.token.value
}

func (c *Compiler) tokenCategory() category {
	return c.scanner.token.category
}

func (c *Compiler) tokenIsType() bool {
	return c.scanner.token.isType()
}

func (c *Compiler) tokenIsOp() bool {
	return c.scanner.token.isOp()
}

func (c *Compiler) tokenIsUnaryOp() bool {
	return c.scanner.token.isUnaryOp()
}

func (c *Compiler) writeToken() {
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

func (c *Compiler) writeTokenAndAdvance() {
	c.writeToken()
	c.advance()
}

func (c *Compiler) writeString(s string) {
	c.output.WriteString(s)
}

func (c *Compiler) checkTokenValue(tokenValue string, message string) {
	if c.tokenValue() == tokenValue {
		c.writeTokenAndAdvance()
	} else {
		log.Fatal(message)
	}
}

func (c *Compiler) checkTokenIdentifier(message string) {
	if c.tokenCategory() == identifier {
		c.writeTokenAndAdvance()
	} else {
		log.Fatal(message)
	}
}

// CompileClass will translate the jack file on the Compiler to an xml
func (c *Compiler) compileClass() {
	errorMessage := "invalid class grammar"
	c.writeString("<class>\n")
	c.advance()
	c.checkTokenValue("class", errorMessage)
	c.checkTokenIdentifier(errorMessage)
	c.checkTokenValue("{", errorMessage)
	c.compileClassVarDec()
	c.compileSubroutine()
	if c.tokenValue() == "}" {
		c.writeToken()
	} else {
		log.Fatal(errorMessage)
	}
	c.writeString("</class>\n")
}

func (c *Compiler) checkTokenIsType(message string) {
	if c.tokenIsType() {
		c.writeTokenAndAdvance()
	} else {
		log.Fatal(message)
	}
}

func (c *Compiler) compileMultipleVarDecs(message string) {
	if c.tokenValue() != "," {
		return
	}
	c.writeTokenAndAdvance()
	c.checkTokenIdentifier(message)
	c.compileMultipleVarDecs(message)
}

func (c *Compiler) compileClassVarDec() {
	errorMessage := "invalid classVarDec grammar"
	if c.tokenValue() != "static" && c.tokenValue() != "field" {
		return
	}
	c.writeString("<classVarDec>\n")
	c.writeTokenAndAdvance()
	c.checkTokenIsType(errorMessage)
	c.checkTokenIdentifier(errorMessage)
	c.compileMultipleVarDecs(errorMessage)
	c.checkTokenValue(";", errorMessage)
	c.writeString("</classVarDec>\n")
	c.compileClassVarDec()
}

func (c *Compiler) compileSubroutine() {
	if c.tokenValue() != "constructor" && c.tokenValue() != "function" && c.tokenValue() != "method" {
		return
	}
	errorMessage := "invalid subroutineDec grammar"
	c.writeString("<subroutineDec>\n")
	c.writeTokenAndAdvance()
	if c.tokenIsType() || c.tokenValue() == "void" {
		c.writeTokenAndAdvance()
	} else {
		log.Fatal(errorMessage)
	}
	c.checkTokenIdentifier(errorMessage)
	c.checkTokenValue("(", errorMessage)
	c.compileParameterList()
	c.checkTokenValue(")", errorMessage)
	c.compileSubroutineBody()
	c.writeString("</subroutineDec>\n")
	c.compileSubroutine()
}

func (c *Compiler) handleMultipleParameters(message string) {
	if c.tokenValue() != "," {
		return
	}
	c.writeTokenAndAdvance()
	c.checkTokenIsType(message)
	c.checkTokenIdentifier(message)
	c.handleMultipleParameters(message)
}

func (c *Compiler) compileParameterList() {
	errorMessage := "invalid parameterList grammar"
	c.writeString("<parameterList>\n")
	if c.tokenValue() == ")" {
		c.writeString("</parameterList>\n")
		return
	}
	c.checkTokenIsType(errorMessage)
	c.checkTokenIdentifier(errorMessage)
	c.handleMultipleParameters(errorMessage)
	c.writeString("</parameterList>\n")
}

func (c *Compiler) compileSubroutineBody() {
	errorMessage := "invalid subroutineBody grammar"
	c.writeString("<subroutineBody>\n")
	c.checkTokenValue("{", errorMessage)
	c.compileVarDec()
	c.compileStatements()
	c.checkTokenValue("}", errorMessage)
	c.writeString("</subroutineBody>\n")
}

func (c *Compiler) compileVarDec() {
	if c.tokenValue() != "var" {
		return
	}
	errorMessage := "invalid varDec grammar"
	c.writeString("<varDec>\n")
	c.writeTokenAndAdvance()
	c.checkTokenIsType(errorMessage)
	c.checkTokenIdentifier(errorMessage)
	c.compileMultipleVarDecs(errorMessage)
	c.checkTokenValue(";", errorMessage)
	c.writeString("</varDec>\n")
	c.compileVarDec()
}

func (c *Compiler) compileStatements() {
	c.writeString("<statements>\n")
	c.compileStatement()
	c.writeString("</statements>\n")
}

func (c *Compiler) compileStatement() {
	switch c.tokenValue() {
	case "let":
		c.compileLet()
		c.compileStatement()
	case "if":
		c.compileIf()
		c.compileStatement()
	case "while":
		c.compileWhile()
		c.compileStatement()
	case "do":
		c.compileDo()
		c.compileStatement()
	case "return":
		c.compileReturn()
		c.compileStatement()
	default:
		return
	}
}

func (c *Compiler) compileLet() {
	if c.tokenValue() != "let" {
		return
	}
	errorMessage := "invalid letStatement grammar"
	c.writeString("<letStatement>\n")
	c.writeTokenAndAdvance()
	c.checkTokenIdentifier(errorMessage)
	switch c.tokenValue() {
	case "[":
		c.writeTokenAndAdvance()
		c.compileExpression()
		c.checkTokenValue("]", errorMessage)
		c.checkTokenValue("=", errorMessage)
		c.compileExpression()
		c.checkTokenValue(";", errorMessage)
		c.writeString("</letStatement>\n")
	case "=":
		c.writeTokenAndAdvance()
		c.compileExpression()
		c.checkTokenValue(";", errorMessage)
		c.writeString("</letStatement>\n")
	default:
		log.Fatal(errorMessage)
	}
}

func (c *Compiler) handleStatements(message string) {
	c.checkTokenValue("{", message)
	c.compileStatements()
	c.checkTokenValue("}", message)
}

func (c *Compiler) handleExpressionStatements(message string) {
	c.checkTokenValue("(", message)
	c.compileExpression()
	c.checkTokenValue(")", message)
	c.handleStatements(message)
}

func (c *Compiler) compileIf() {
	if c.tokenValue() != "if" {
		return
	}
	errorMessage := "invalid if grammar"
	c.writeString("<ifStatement>\n")
	c.writeTokenAndAdvance()
	c.handleExpressionStatements(errorMessage)
	if c.tokenValue() == "else" {
		c.writeTokenAndAdvance()
		c.handleStatements(errorMessage)
	}
	c.writeString("</ifStatement>\n")
}

func (c *Compiler) compileWhile() {
	if c.tokenValue() != "while" {
		return
	}
	errorMessage := "invalid while grammar"
	c.writeString("<whileStatement>\n")
	c.writeTokenAndAdvance()
	c.handleExpressionStatements(errorMessage)
	c.writeString("</whileStatement>\n")
}

func (c *Compiler) compileDo() {
	if c.tokenValue() != "do" {
		return
	}
	errorMessage := "invalid do grammar"
	c.writeString("<doStatement>\n")
	c.writeTokenAndAdvance()
	c.compileSubroutineCall()
	c.checkTokenValue(";", errorMessage)
	c.writeString("</doStatement>\n")
}

func (c *Compiler) compileReturn() {
	if c.tokenValue() != "return" {
		return
	}
	errorMessage := "invalid return grammar"
	c.writeString("<returnStatement>\n")
	c.writeTokenAndAdvance()
	if c.tokenValue() == ";" {
		c.writeTokenAndAdvance()
		c.writeString("</returnStatement>\n")
		return
	}
	c.compileExpression()
	c.checkTokenValue(";", errorMessage)
	c.writeString("</returnStatement>\n")
}

func (c *Compiler) handleMultipleTerms() {
	if c.tokenIsOp() {
		c.writeTokenAndAdvance()
	} else {
		return
	}
	c.compileTerm()
	c.handleMultipleTerms()
}

func (c *Compiler) compileExpression() {
	c.writeString("<expression>\n")
	c.compileTerm()
	c.handleMultipleTerms()
	c.writeString("</expression>\n")
}

func (c *Compiler) handleIdentifierTerm(errorMessage string) {
	if c.tokenValue() == "[" {
		c.writeTokenAndAdvance()
		c.compileExpression()
		c.checkTokenValue("]", errorMessage)
		c.writeString("</term>\n")
	} else if c.tokenValue() == "(" {
		c.writeTokenAndAdvance()
		c.compileExpressionList()
		c.checkTokenValue(")", errorMessage)
		c.writeString("</term>\n")
	} else if c.tokenValue() == "." {
		c.writeTokenAndAdvance()
		c.checkTokenIdentifier(errorMessage)
		c.checkTokenValue("(", errorMessage)
		c.compileExpressionList()
		c.checkTokenValue(")", errorMessage)
		c.writeString("</term>\n")
	} else {
		c.writeString("</term>\n")
		return
	}
}

func (c *Compiler) compileTerm() {
	errorMessage := "invalid term grammar"
	c.writeString("<term>\n")
	if c.tokenCategory() == intConst || c.tokenCategory() == stringConst || c.tokenCategory() == keyword {
		c.writeTokenAndAdvance()
		c.writeString("</term>\n")
	} else if c.tokenCategory() == identifier {
		c.writeTokenAndAdvance()
		c.handleIdentifierTerm(errorMessage)
	} else if c.tokenValue() == "(" {
		c.writeTokenAndAdvance()
		c.compileExpression()
		c.checkTokenValue(")", errorMessage)
		c.writeString("</term>\n")
	} else if c.tokenIsUnaryOp() {
		c.writeTokenAndAdvance()
		c.compileTerm()
		c.writeString("</term>\n")
	}
}

func (c *Compiler) compileSubroutineCall() {
	errorMessage := "invalid subroutineCall grammar"
	c.checkTokenIdentifier(errorMessage)
	if c.tokenValue() == "(" {
		c.writeTokenAndAdvance()
		c.compileExpressionList()
		c.checkTokenValue(")", errorMessage)
	} else if c.tokenValue() == "." {
		c.writeTokenAndAdvance()
		c.checkTokenIdentifier(errorMessage)
		c.checkTokenValue("(", errorMessage)
		c.compileExpressionList()
		c.checkTokenValue(")", errorMessage)
	} else {
		log.Fatal(errorMessage)
	}
}

func (c *Compiler) handleMultipleExpressions() {
	if c.tokenValue() != "," {
		return
	}
	c.writeTokenAndAdvance()
	c.compileExpression()
	c.handleMultipleExpressions()
}

func (c *Compiler) compileExpressionList() {
	c.writeString("<expressionList>\n")
	if c.tokenValue() == ")" {
		c.writeString("</expressionList>\n")
		return
	}
	c.compileExpression()
	c.handleMultipleExpressions()
	c.writeString("</expressionList>\n")
}

func compileFile(path string) {
	outPath := strings.Replace(path, ".jack", "(gen).xml", 1)
	outFile, err := os.Create(outPath)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := outFile.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	writer := bufio.NewWriter(outFile)
	defer func() {
		if err := writer.Flush(); err != nil {
			log.Fatal(err)
		}
	}()

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	compiler := NewCompiler(file, writer)
	compiler.compileClass()
}

func getJackFiles(jackFiles *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".jack" {
			*jackFiles = append(*jackFiles, path)
		}
		return nil
	}
}

func Compile(path string) {
	path = filepath.Clean(path)
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}
	if fileInfo.IsDir() {
		var jackFiles []string
		err := filepath.Walk(path, getJackFiles(&jackFiles))
		if err != nil {
			log.Fatal(err)
		}

		for _, path = range jackFiles {
			compileFile(path)
		}
	} else {
		compileFile(path)
	}
}

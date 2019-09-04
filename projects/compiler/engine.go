package compiler

import (
	"bufio"
	"fmt"
	"io"
	"log"
)

// Engine consists of a scanner that steps through the tokens in a jack file,
// a bufio writer to produce the parsed file and a token to handle tokens.
type Engine struct {
	scanner *scanner
	output  *bufio.Writer
	token   *token
}

// NewEngine creates an engine to process a file from the given path
func NewEngine(r io.Reader, w *bufio.Writer) *Engine {
	// Create output xml file
	scanner := newScanner(r)
	return &Engine{
		scanner: scanner,
		output:  w,
	}
}

// advance retrieves each token string from the scanner, creates a new token
// and sets the engines token
func (e *Engine) advance() {
	e.scanner.Scan()
	tokenValue := e.scanner.Text()
	token, err := newToken(tokenValue)
	if err != nil {
		log.Fatalf("Invalid token: %s", token.value)
	}
	e.token = token
}

func (e *Engine) writeString(s string) {
	e.output.WriteString(s)
}

// writeToken writes the current token to the output file
func (e *Engine) writeToken() {
	var label string
	switch e.tokenType() {
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
	e.output.WriteString(fmt.Sprintf("<%s>%s</%s>\n", label, e.tokenValue(), label))
}

func (e *Engine) tokenValue() string {
	return e.token.value
}

func (e *Engine) tokenType() tokenType {
	return e.token.tokenType
}

func (e *Engine) tokenIsType() bool {
	return e.token.isType()
}

func (e *Engine) tokenIsOp() bool {
	return e.token.isOp()
}

// CompileClass will translate the jack file on the engine to an xml
func (e *Engine) CompileClass() {
	e.writeString("<class>\n")

	// Scan through tokens ensuring the grammar specification is met, if not throw an error
	e.advance()
	if e.tokenValue() == "class" {
		e.writeToken()
	} else {
		log.Fatal("invalid class grammar")
	}

	e.advance()
	if e.tokenType() == identifier {
		e.writeToken()
	} else {
		log.Fatal("invalid class grammar")
	}

	e.advance()
	if e.tokenValue() == "{" {
		e.writeToken()
	} else {
		log.Fatal("invalid class grammar")
	}

	e.compileClassVarDec()
	e.compileSubroutine()

	if e.tokenValue() == "}" {
		e.writeToken()
	} else {
		log.Fatal("incorrect class grammar")
	}
	e.writeString("</class>\n")
}

func (e *Engine) handleMultipleVarDecs() {
	e.advance()
	if e.tokenValue() == "," {
		e.writeToken()
	} else {
		return
	}

	e.advance()
	if e.tokenType() == identifier {
		e.writeToken()
		e.handleMultipleVarDecs()
	} else {
		log.Fatal("invalid grammar for declaring multiple variables on the same line")
	}
}

func (e *Engine) compileClassVarDec() {
	e.advance()
	switch e.tokenValue() {
	case "static", "field":
		e.writeString("<classVarDec>\n")
		e.writeToken()

		e.advance()
		if e.tokenIsType() {
			e.writeToken()
		} else {
			log.Fatal("invalid classVarDec grammar")
		}

		e.advance()
		if e.tokenType() == identifier {
			e.writeToken()
		} else {
			log.Fatal("invalid classVarDec grammer")
		}

		e.handleMultipleVarDecs()

		if e.tokenValue() == ";" {
			e.writeToken()
			e.writeString("</classVarDec>\n")
			e.compileClassVarDec()
		} else {
			log.Fatal("invalid classVarDec grammar")
		}
	default:
		return
	}
}

func (e *Engine) compileSubroutine() {
	switch e.tokenValue() {
	case "constructor", "function", "method":
		e.writeString("<subroutineDec>\n")
		e.writeToken()

		e.advance()
		if e.tokenIsType() || e.tokenValue() == "void" {
			e.writeToken()
		} else {
			log.Fatal("invalid subroutineDec grammar")
		}

		e.advance()
		if e.tokenType() == identifier {
			e.writeToken()
		} else {
			log.Fatal("invalid subroutineDec grammar")
		}

		e.advance()
		if e.tokenValue() == "(" {
			e.writeToken()
		} else {
			log.Fatal("invalid subroutineDec grammar")
		}

		e.compileParameterList()

		if e.tokenValue() == ")" {
			e.writeToken()
		} else {
			log.Fatal("invalid subroutineDec grammar")
		}

		e.compileSubroutineBody()
		e.writeString("</subroutineDec>\n")

		e.compileSubroutine()
	default:
		return
	}
}

func (e *Engine) handleMultipleParameters() {
	e.advance()
	if e.tokenValue() == "," {
		e.writeToken()
	} else {
		return
	}

	e.advance()
	if e.tokenIsType() {
		e.writeToken()
	} else {
		log.Fatal("invalid parameterList grammer")
	}

	e.advance()
	if e.tokenType() == identifier {
		e.writeToken()
		e.handleMultipleParameters()
	} else {
		log.Fatal("invalid parameterList grammar")
	}
}

func (e *Engine) compileParameterList() {
	e.advance()
	e.writeString("<parameterList>\n")
	// handle case of an empty parameter list
	if e.tokenValue() == ")" {
		e.writeString("</parameterList>\n")
		return
	}

	if e.tokenIsType() {
		e.writeToken()
	} else {
		log.Fatal("invalid parameterList grammer")
	}

	e.advance()
	if e.tokenType() == identifier {
		e.writeToken()
	} else {
		log.Fatal("invalid parameterList grammar")
	}

	e.handleMultipleParameters()
	e.writeString("</parameterList>\n")
}

func (e *Engine) compileSubroutineBody() {
	e.advance()
	e.writeString("<subroutineBody>\n")
	if e.tokenValue() == "{" {
		e.writeToken()
	} else {
		log.Fatal("invalid subroutineBody grammar")
	}

	e.compileVarDec()
	e.compileStatements()

	if e.tokenValue() == "}" {
		e.writeToken()
	} else {
		log.Fatal("invalid subroutineBodyGrammar")
	}

	e.writeString("</subroutineBody>\n")
	e.advance()
}

func (e *Engine) compileVarDec() {
	e.advance()
	if e.tokenValue() == "var" {
		e.writeString("<varDec>\n")
		e.writeToken()
	} else {
		return
	}

	e.advance()
	if e.tokenIsType() {
		e.writeToken()
	} else {
		log.Fatal("invalid varDec grammar")
	}

	e.advance()
	if e.tokenType() == identifier {
		e.writeToken()
	} else {
		log.Fatal("invalid varDec grammar")
	}

	e.handleMultipleVarDecs()

	if e.tokenValue() == ";" {
		e.writeToken()
		e.writeString("</varDec>\n")
		e.compileVarDec()
	} else {
		log.Fatal("invalid varDec grammar")
	}
}

func (e *Engine) compileStatements() {
	e.writeString("<statements>\n")
	e.compileStatement()
	e.writeString("</statements>\n")
}

func (e *Engine) compileStatement() {
	switch e.tokenValue() {
	case "let":
		e.compileLet()
		e.compileStatement()
	case "if":
		e.compileIf()
		e.compileStatement()
	case "while":
		e.compileWhile()
		e.compileStatement()
	case "do":
		e.compileDo()
		e.compileStatement()
	case "return":
		e.compileReturn()
		e.compileStatement()
	default:
		return
	}
}

func (e *Engine) compileLet() {
	e.writeString("<letStatement>\n")
	e.writeToken()

	e.advance()
	if e.tokenType() == identifier {
		e.writeToken()
	} else {
		log.Fatal("invalid letStatement grammar")
	}

	e.advance()
	switch e.tokenValue() {
	case "[":
		e.writeToken()

		e.advance()
		e.compileExpression()

		if e.tokenValue() == "]" {
			e.writeToken()
		} else {
			log.Fatal("invalid letStatement grammar")
		}

		e.advance()
		if e.tokenValue() == "=" {
			e.writeToken()
		} else {
			log.Fatal("invalid letStatement grammar")
		}

		e.advance()
		e.compileExpression()

		if e.tokenValue() == ";" {
			e.writeToken()
			e.writeString("</letStatement>\n")
			e.advance()
		} else {
			log.Fatal("invalid letStatement grammar")
		}
	case "=":
		e.writeToken()

		e.advance()
		e.compileExpression()

		if e.tokenValue() == ";" {
			e.writeToken()
			e.writeString("</letStatement>\n")
			e.advance()
		} else {
			log.Fatal("invalid letStatement grammar")
		}
	default:
		log.Fatal("invalid letStatement grammar")
	}
}

func (e *Engine) handleStatements() {
	e.advance()
	if e.tokenValue() == "{" {
		e.writeToken()
	} else {
		log.Fatal("invalid handleStatements grammar")
	}

	e.advance()
	e.compileStatements()

	if e.tokenValue() == "}" {
		e.writeToken()
		e.advance()
	} else {
		log.Fatal("invalid handleStatements grammar")
	}
}

func (e *Engine) handleExpressionStatements() {
	e.advance()
	if e.tokenValue() == "(" {
		e.writeToken()
	} else {
		log.Fatal("invalid handleExpressionStatements grammar")
	}

	e.advance()
	e.compileExpression()

	if e.tokenValue() == ")" {
		e.writeToken()
	} else {
		log.Fatal("invalid handleExpressionStatements grammar")
	}

	e.handleStatements()
}

func (e *Engine) compileIf() {
	e.writeString("<ifStatement>\n")
	e.writeToken()

	e.handleExpressionStatements()

	if e.tokenValue() == "else" {
		e.writeToken()

		e.handleStatements()
	}

	e.writeString("</ifStatement>\n")

}

func (e *Engine) compileWhile() {
	e.writeString("<whileStatement>\n")
	e.writeToken()

	e.handleExpressionStatements()

	e.writeString("</whileStatement>\n")
}

func (e *Engine) compileDo() {
	e.writeString("<doStatement>\n")
	e.writeToken()

	e.compileSubroutineCall()

	e.advance()
	if e.tokenValue() == ";" {
		e.writeToken()
		e.writeString("</doStatement>\n")
		e.advance()
	} else {
		log.Fatal("invalid doStatement grammar")
	}
}

func (e *Engine) compileReturn() {
	e.writeString("<returnStatement>\n")
	e.writeToken()

	e.advance()
	if e.tokenValue() == ";" {
		e.writeToken()
		e.writeString("</returnStatement>\n")
		e.advance()
		return
	}
	e.compileExpression()

	if e.tokenValue() == ";" {
		e.writeToken()
		e.writeString("</returnStatement>\n")
		e.advance()
	} else {
		log.Fatal("invalid returnStatement grammar")
	}
}

func (e *Engine) handleMultipleTerms() {
	e.advance()
	if e.tokenIsOp() {
		e.writeToken()
	} else {
		return
	}

	e.advance()
	e.compileTerm()

	e.handleMultipleTerms()
}

func (e *Engine) compileExpression() {
	e.writeString("<expression>\n")
	e.compileTerm()
	e.handleMultipleTerms()
	e.writeString("</expression>\n")
}

func (e *Engine) compileTerm() {
	e.writeString("<term>\n")
	e.writeToken()
	e.writeString("</term>\n")
}

func (e *Engine) compileSubroutineCall() {
	e.advance()
	if e.tokenType() == identifier {
		e.writeToken()
	} else {
		log.Fatal("invalid subroutineCall grammar")
	}

	e.advance()
	if e.tokenValue() == "(" {
		e.writeToken()

		e.compileExpressionList()

		if e.tokenValue() == ")" {
			e.writeToken()
		} else {
			log.Fatal("invalid subroutineCall grammar")
		}
	} else if e.tokenValue() == "." {
		e.writeToken()

		e.advance()
		if e.tokenType() == identifier {
			e.writeToken()
		} else {
			log.Fatal("invalid subroutineCall grammar")
		}

		e.advance()
		if e.tokenValue() == "(" {
			e.writeToken()
		} else {
			log.Fatal("invalid subroutineCall grammar")
		}

		e.compileExpressionList()

		if e.tokenValue() == ")" {
			e.writeToken()
		} else {
			log.Fatal("invalid subroutineCall grammar")
		}
	} else {
		log.Fatal("invalid subroutineCall grammar")
	}
}

func (e *Engine) handleMultipleExpressions() {
	if e.tokenValue() == "," {
		e.writeToken()
	} else {
		return
	}

	e.advance()
	e.compileExpression()

	e.handleMultipleExpressions()
}

func (e *Engine) compileExpressionList() {
	e.advance()
	e.writeString("<expressionList>\n")
	if e.tokenValue() == ")" {
		e.writeString("</expressionList>\n")
		return
	}

	e.compileExpression()

	e.handleMultipleExpressions()

	e.writeString("</expressionList>\n")
}

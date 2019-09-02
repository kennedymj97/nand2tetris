package compiler

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

// Engine consists of a scanner that steps through the tokens in a jack file,
// a bufio writer to produce the parsed file and a token to handle tokens.
type Engine struct {
	scanner *scanner
	output  *bufio.Writer
	token   *token
}

// NewEngine creates an engine to process a file from the given path
func NewEngine(path string) *Engine {
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

	w := bufio.NewWriter(outFile)
	defer func() {
		if err := w.Flush(); err != nil {
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

	// Create output xml file

	scanner := newScanner(file)
	return &Engine{
		scanner: scanner,
		output:  w,
	}
}

// advance retrieves each token string from the scanner, creates a new token
// and sets the engines token
func (e *Engine) advance() {
	e.scanner.Scan()
	token, err := newToken(e.scanner.Text())
	if err != nil {
		log.Fatalf("Invalid token: %s", token.value)
	}
	e.token = token
}

// writeToken writes the current token to the output file
func (e *Engine) writeToken() {
	var label string
	switch e.token.tokenType {
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
	e.output.WriteString(fmt.Sprintf("<%s>%s</%s>\n", label, e.token.value, label))
}

// CompileClass will translate the jack file on the engine to an xml
func (e *Engine) CompileClass() {
	e.output.WriteString("<class>\n")

	// Scan through tokens ensuring the grammar specification is met, if not throw an error
	e.advance()
	if e.token.value == "class" {
		e.writeToken()
	} else {
		log.Fatal("invalid class grammar")
	}

	e.advance()
	if e.token.tokenType == identifier {
		e.writeToken()
	} else {
		log.Fatal("invalid class grammar")
	}

	e.advance()
	if e.token.value == "{" {
		e.writeToken()
	} else {
		log.Fatal("invalid class grammar")
	}

	e.advance()
	e.compileClassVarDec()

	e.compileSubroutine()

	if e.token.value == "}" {
		e.writeToken()
	} else {
		log.Fatal("incorrect class grammar")
	}
	e.output.WriteString("</class>\n")
}

func (e *Engine) compileClassVarDec() {
	for {
		switch e.token.value {
		case "static", "field":
			e.output.WriteString("<classVarDec>\n")
			e.writeToken()
		default:
			break
		}

		e.advance()
		if e.token.isType() {
			e.writeToken()
		} else {
			log.Fatal("invalid classVarDec grammar")
		}

		e.advance()
		if e.token.tokenType == identifier {
			e.writeToken()
		} else {
			log.Fatal("invalid classVarDec grammer")
		}

		for {
			e.advance()
			if e.token.value == ";" {
				e.writeToken()
				e.output.WriteString("</classVarDec>\n")
				e.advance()
				break
			} else if e.token.value == "," {
				e.writeToken()
			} else {
				log.Fatal("invalid classVarDec grammar")
			}

			e.advance()
			if e.token.tokenType == identifier {
				e.writeToken()
			} else {
				log.Fatal("invalid classVarDec grammar")
			}
		}
	}
}

func (e *Engine) compileSubroutine() {
	for {
		switch e.token.value {
		case "constructor", "function", "method":
			e.output.WriteString("<subroutineDec>\n")
			e.writeToken()
		default:
			break
		}

		e.advance()
		if e.token.isType() || e.token.value == "void" {
			e.writeToken()
		} else {
			log.Fatal("invalid subroutineDec grammar")
		}

		e.advance()
		if e.token.tokenType == identifier {
			e.writeToken()
		} else {
			log.Fatal("invalid subroutineDec grammar")
		}

		e.advance()
		if e.token.value == "(" {
			e.writeToken()
		} else {
			log.Fatal("invalid subroutineDec grammar")
		}

		e.advance()
		e.compileParameterList()

		if e.token.value == ")" {
			e.writeToken()
		} else {
			log.Fatal("invalid subroutineDec grammar")
		}

		e.advance()
		e.compileSubroutineBody()

		e.advance()
	}
}

func (e *Engine) compileParameterList() {
	e.output.WriteString("<parameterList>\n")
	// handle case of an empty parameter list
	if e.token.value == ")" {
		e.output.WriteString("</parameterList>\n")
		return
	}

	if e.token.isType() {
		e.writeToken()
	} else {
		log.Fatal("invalid parameterList grammer")
	}

	e.advance()
	if e.token.tokenType == identifier {
		e.writeToken()
	} else {
		log.Fatal("invalid parameterList grammar")
	}

	for {
		e.advance()
		if e.token.value == "," {
			e.writeToken()
		} else if e.token.value == ")" {
			e.output.WriteString("</parameterList>\n")
			break
		} else {
			log.Fatal("invalid parameterList grammar")
		}

		e.advance()
		if e.token.isType() {
			e.writeToken()
		} else {
			log.Fatal("invalid parameterList grammer")
		}

		e.advance()
		if e.token.tokenType == identifier {
			e.writeToken()
		} else {
			log.Fatal("invalid parameterList grammar")
		}
	}
}

func (e *Engine) compileSubroutineBody() {
	e.output.WriteString("<subroutineBody>\n")
	if e.token.value == "{" {
		e.writeToken()
	} else {
		log.Fatal("invalid subroutineBody grammar")
	}

	e.advance()
	e.compileVarDec()

	e.compileStatements()

	if e.token.value == "}" {
		e.writeToken()
	} else {
		log.Fatal("invalid subroutineBodyGrammar")
	}
	e.output.WriteString("</subroutineBody>\n")
}

func (e *Engine) compileVarDec() {
	for {
		if e.token.value == "var" {
			e.output.WriteString("<varDec>\n")
			e.writeToken()
		} else {
			break
		}

		e.advance()
		if e.token.isType() {
			e.writeToken()
		} else {
			log.Fatal("invalid varDec grammar")
		}

		e.advance()
		if e.token.tokenType == identifier {
			e.writeToken()
		} else {
			log.Fatal("invalid varDec grammar")
		}

		for {
			e.advance()
			if e.token.value == ";" {
				e.writeToken()
				e.output.WriteString("</varDec>\n")
				e.advance()
				break
			} else if e.token.value == "," {
				e.writeToken()
			} else {
				log.Fatal("invalid varDec grammar")
			}

			e.advance()
			if e.token.tokenType == identifier {
				e.writeToken()
			} else {
				log.Fatal("invalid varDec grammar")
			}
		}
	}
}

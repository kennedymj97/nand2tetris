package engine

import (
	"bufio"
	"compiler/tokenizer"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type writer struct {
	*bufio.Writer
}

func newWriter(file io.Writer) *writer {
	return &writer{bufio.NewWriter(file)}
}

func writeTerminal(token string, label string) string {
	return fmt.Sprintf("<%s>%s</%s>\n", label, token, label)
}

func (out *writer) writeToken(token string) {
	token, tokenType := tokenizer.ParseToken(token)
	var label string
	switch tokenType {
	case "KEYWORD":
		label = "keyword"
	case "SYMBOL":
		label = "symbol"
	case "STRING_CONST":
		label = "stringConstant"
	case "INTEGER_CONST":
		label = "integerConstant"
	case "IDENTIFIER":
		label = "identifier"
	}
	out.WriteString(writeTerminal(token, label))
}

type scanner struct {
	*bufio.Scanner
	token     string
	tokenType string
}

func newScanner(file io.Reader) *scanner {
	bufioScanner := bufio.NewScanner(file)
	bufioScanner.Split(tokenizer.ScanTokens)
	return &scanner{
		bufioScanner,
		"",
		"",
	}
}

func (s *scanner) advance() {
	s.Scan()
	token := s.Text()
	s.token, s.tokenType = tokenizer.ParseToken(token)
}

type Engine struct {
	scanner *scanner
	output  *writer
}

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

	w := newWriter(outFile)
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

func (e *Engine) CompileClass() {
	e.output.WriteString("<class>\n")

	// Scan through tokens ensuring the grammar specification is met, if not throw an error
	e.scanner.advance()
	if e.scanner.token == "class" {
		e.output.writeToken(e.scanner.token)
	} else {
		log.Fatal("invalid class grammar")
	}

	e.scanner.advance()
	if e.scanner.tokenType == "IDENTIFIER" {
		e.output.writeToken(e.scanner.token)
	} else {
		log.Fatal("invalid class grammar")
	}

	e.scanner.advance()
	if e.scanner.token == "{" {
		e.output.writeToken(e.scanner.token)
	} else {
		log.Fatal("invalid class grammar")
	}

	e.scanner.advance()
	e.compileClassVarDec()

	e.compileSubroutine()

	if e.scanner.token == "}" {
		e.output.writeToken(e.scanner.token)
	} else {
		log.Fatal("incorrect class grammar")
	}
	e.output.WriteString("</class>\n")
}

func isType(token string, tokenType string) bool {
	if tokenType == "IDENTIFIER" {
		return true
	}

	types := []string{
		"int",
		"char",
		"boolean",
	}
	for _, jackType := range types {
		if token == jackType {
			return true
		}
	}

	return false
}

func (e *Engine) compileClassVarDec() {
	for {
		switch e.scanner.token {
		case "static", "field":
			e.output.WriteString("<classVarDec>\n")
			e.output.writeToken(e.scanner.token)
		default:
			break
		}

		e.scanner.advance()
		if isType(e.scanner.token, e.scanner.tokenType) {
			e.output.writeToken(e.scanner.token)
		} else {
			log.Fatal("invalid classVarDec grammar")
		}

		e.scanner.advance()
		if e.scanner.tokenType == "IDENTIFIER" {
			e.output.writeToken(e.scanner.token)
		} else {
			log.Fatal("invalid classVarDec grammer")
		}

		for {
			e.scanner.advance()
			if e.scanner.token == ";" {
				e.output.writeToken(e.scanner.token)
				e.output.WriteString("</classVarDec>\n")
				e.scanner.advance()
				break
			} else if e.scanner.token == "," {
				e.output.writeToken(e.scanner.token)
			} else {
				log.Fatal("invalid classVarDec grammar")
			}

			e.scanner.advance()
			if e.scanner.tokenType == "IDENTIFIER" {
				e.output.writeToken(e.scanner.token)
			} else {
				log.Fatal("invalid classVarDec grammar")
			}
		}
	}
}

func (e *Engine) compileSubroutine() {
	for {
		switch e.scanner.token {
		case "constructor", "function", "method":
			e.output.WriteString("<subroutineDec>\n")
			e.output.writeToken(e.scanner.token)
		default:
			break
		}

		e.scanner.advance()
		if isType(e.scanner.token, e.scanner.tokenType) || e.scanner.token == "void" {
			e.output.writeToken(e.scanner.token)
		} else {
			log.Fatal("invalid subroutineDec grammar")
		}

		e.scanner.advance()
		if e.scanner.tokenType == "IDENTIFIER" {
			e.output.writeToken(e.scanner.token)
		} else {
			log.Fatal("invalid subroutineDec grammar")
		}

		e.scanner.advance()
		if e.scanner.token == "(" {
			e.output.writeToken(e.scanner.token)
		} else {
			log.Fatal("invalid subroutineDec grammar")
		}

		e.scanner.advance()
		e.compileParameterList()

		if e.scanner.token == ")" {
			e.output.writeToken(e.scanner.token)
		} else {
			log.Fatal("invalid subroutineDec grammar")
		}

		e.scanner.advance()
		e.compileSubroutineBody()

		e.scanner.advance()
	}
}

func (e *Engine) compileParameterList() {
	e.output.WriteString("<parameterList>\n")
	// handle case of an empty parameter list
	if e.scanner.token == ")" {
		e.output.WriteString("</parameterList>\n")
		return
	}

	if isType(e.scanner.token, e.scanner.tokenType) {
		e.output.writeToken(e.scanner.token)
	} else {
		log.Fatal("invalid parameterList grammer")
	}

	e.scanner.advance()
	if e.scanner.tokenType == "IDENTIFER" {
		e.output.writeToken(e.scanner.token)
	} else {
		log.Fatal("invalid parameterList grammar")
	}

	for {
		e.scanner.advance()
		if e.scanner.token == "," {
			e.output.writeToken(e.scanner.token)
		} else if e.scanner.token == ")" {
			e.output.WriteString("</parameterList>\n")
			break
		} else {
			log.Fatal("invalid parameterList grammar")
		}

		e.scanner.advance()
		if isType(e.scanner.token, e.scanner.tokenType) {
			e.output.writeToken(e.scanner.token)
		} else {
			log.Fatal("invalid parameterList grammer")
		}

		e.scanner.advance()
		if e.scanner.tokenType == "IDENTIFIER" {
			e.output.writeToken(e.scanner.token)
		} else {
			log.Fatal("invalid parameterList grammar")
		}
	}
}

func (e *Engine) compileSubroutineBody() {
	e.output.WriteString("<subroutineBody>\n")
	if e.scanner.token == "{" {
		e.output.writeToken(e.scanner.token)
	} else {
		log.Fatal("invalid subroutineBody grammar")
	}

	e.scanner.advance()
	e.compileVarDec()

	e.compileStatements()

	if e.scanner.token == "}" {
		e.output.writeToken(e.scanner.token)
	} else {
		log.Fatal("invalid subroutineBodyGrammar")
	}
	e.output.WriteString("</subroutineBody>\n")
}

func (e *Engine) compileVarDec() {
	for {
		if e.scanner.token == "var" {
			e.output.WriteString("<varDec>\n")
			e.output.writeToken(e.scanner.token)
		} else {
			break
		}

		e.scanner.advance()
		if isType(e.scanner.token, e.scanner.tokenType) {
			e.output.writeToken(e.scanner.token)
		} else {
			log.Fatal("invalid varDec grammar")
		}

		e.scanner.advance()
		if e.scanner.tokenType == "IDENTIFIER" {
			e.output.writeToken(e.scanner.token)
		} else {
			log.Fatal("invalid varDec grammar")
		}

		for {
			e.scanner.advance()
			if e.scanner.token == ";" {
				e.output.writeToken(e.scanner.token)
				e.output.WriteString("</varDec>\n")
				e.scanner.advance()
				break
			} else if e.scanner.token == "," {
				e.output.writeToken(e.scanner.token)
			} else {
				log.Fatal("invalid varDec grammar")
			}

			e.scanner.advance()
			if e.scanner.tokenType == "IDENTIFIER" {
				e.output.writeToken(e.scanner.token)
			} else {
				log.Fatal("invalid varDec grammar")
			}
		}
	}
}

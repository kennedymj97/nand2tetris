package engine

import (
	"bufio"
	"compiler/tokenizer"
	"fmt"
	"log"
	"os"
	"strings"
)

type Engine struct {
	scanner *bufio.Scanner
	output  *bufio.Writer
}

func New(path string) *Engine {
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

	scanner := bufio.NewScanner(file)
	scanner.Split(tokenizer.ScanTokens)
	return &Engine{
		scanner: scanner,
		output:  w,
	}
}

func writeTerminal(token string, label string) string {
	return fmt.Sprintf("<%s>%s</%s>\n", label, token, label)
}

func (e *Engine) CompileClass() {
	e.output.WriteString("<class>\n")
	for e.scanner.Scan() {
		token := e.scanner.Text()
		tokenType := tokenizer.TokenType(token)
		switch tokenType {
		case "KEYWORD":
			token, label := tokenizer.Keyword(token)
			switch token {
			case "class":
				e.output.WriteString(writeTerminal(token, label))
			case "static", "field":
				e.classVarDec()
			case "constructor", "function", "method":
				e.subroutineDec()
			}
		case "SYMBOL":
			token, label := tokenizer.Symbol(token)
			e.output.WriteString(writeTerminal(token, label))
		case "IDENTIFIER":
			token, label := tokenizer.Identifier(token)
			e.output.WriteString(writeTerminal(token, label))
		default:
			token, label := "", ""
		}
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

func (e *Engine) classVarDec() {
	e.output.WriteString("<classVarDec>\n")
	// Get current token, has to be either static or field and write to file
	token := e.scanner.Text()
	token, label := tokenizer.Keyword(token)
	e.output.WriteString(writeTerminal(token, label))

	// Scan the next token and write it if is a type if not return
	e.scanner.Scan()
	token = e.scanner.Text()
	tokenType := tokenizer.TokenType(token)
	if isType(token, tokenType) {
		e.output.WriteString(writeTerminal(token, label))
	} else {
		return
	}

	// Scan the rest of the tokens must be varNames or commas
	// Break when comma token is reached
	for e.scanner.Scan() {
		token := e.scanner.Text()
		tokenType := tokenizer.TokenType(token)
		if tokenType == "IDENTIFIER" {
			token, label := tokenizer.Identifier(token)
			e.output.WriteString(writeTerminal(token, label))
		} else if token == "," {
			token, label := tokenizer.Symbol(token)
			e.output.WriteString(writeTerminal(token, label))
		} else if token == ";" {
			token, label := tokenizer.Symbol(token)
			e.output.WriteString(writeTerminal(token, label))
			break
		} else {
			break
		}
	}
	e.output.WriteString("</classVarDec>\n")
}

func (e *Engine) subroutineDec() {

}

package main

import (
	"bufio"
	"compiler/tokenizer"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	args := os.Args[1:]

	path := filepath.Clean(args[0])

	outPath := strings.Replace(path, ".jack", "T(gen).xml", 1)
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
	w.WriteString(`<tokens>
`)
	for scanner.Scan() {
		token := scanner.Text()
		tokenType := tokenizer.TokenType(token)
		if tokenType == "STRING_CONST" {
			token = token[1 : len(token)-1]
		}
		switch token {
		case "<":
			token = "&lt;"
		case ">":
			token = "&gt;"
		case "&":
			token = "&amp;"
		case `"`:
			token = "&quot;"
		}
		var tokenLabel string
		switch tokenType {
		case "KEYWORD":
			tokenLabel = "keyword"
		case "SYMBOL":
			tokenLabel = "symbol"
		case "STRING_CONST":
			tokenLabel = "stringConstant"
		case "INT_CONST":
			tokenLabel = "integerConstant"
		case "IDENTIFIER":
			tokenLabel = "identifier"
		default:
			tokenLabel = ""
		}
		w.WriteString(fmt.Sprintf(`<%s>%s</%s>
`, tokenLabel, token, tokenLabel))
	}
	w.WriteString(`</tokens>`)
}

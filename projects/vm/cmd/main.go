package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"

	"vm/parser"
	"vm/translater"
)

func main() {
	args := os.Args[1:]
	file, err := os.Open(args[0])
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	dirname, filename := filepath.Split(args[0])
	fnameNoExt := strings.Replace(filename, ".vm", "", 1)
	outFilename := strings.Replace(filename, ".vm", ".asm", 1)
	outFilePath := dirname + outFilename
	outFile, err := os.Create(outFilePath)
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

	var p parser.VmParser
	var aw translater.Translater
	equalityCheckCount := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		p = &parser.Command{Line: scanner.Text()}
		formattedCommand := p.FormatLine()
		if formattedCommand == "" {
			continue
		}
		commandType, err := p.CommandType()
		if err != nil {
			log.Fatal(err)
		}

		firstArg, err := p.Arg1(commandType)
		if err != nil {
			log.Fatal(err)
		}

		secondArg, err := p.Arg2(commandType)
		if err != nil {
			log.Fatal(err)
		}

		aw = &translater.AssemblyWriter{
			Filename:    fnameNoExt,
			CommandType: commandType,
			Arg1:        firstArg,
			Arg2:        secondArg,
		}

		var assemblyCode string
		var equalityInc int
		switch commandType {
		case "C_ARITHMETIC":
			assemblyCode, equalityInc = aw.WriteArithmetic(equalityCheckCount)
		case "C_PUSH", "C_POP":
			assemblyCode = aw.WritePushPop()
		case "C_LABEL":
			assemblyCode = aw.WriteLabel()
		case "C_GOTO":
			assemblyCode = aw.WriteGoto()
		case "C_IF":
			assemblyCode = aw.WriteIf()
		}

		equalityCheckCount += equalityInc

		w.WriteString(assemblyCode)
	}
}
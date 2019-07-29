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
	// set args to all the args after filename
	args := os.Args[1:]

	// check if the first arg is a directory
	path := filepath.Clean(args[0])
	isPathDir, err := isDirectory(path)

	// create output filestream
	outFilePath := createOutputPath(path, isPathDir)
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
	
	// write init for output file
	var aw translater.Translater
	aw = &translater.AssemblyWriter{
		Filename: "",
		FunctionName: "",
		CommandType: "",
		Arg1: "Sys.init",
		Arg2: 0,
	}
	w.WriteString(aw.WriteInit())

	// if a directory process all files in directory else just process file
	if isPathDir {
		var files []string
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			files = append(files, path)
			return nil
		})
		if err != nil {
			panic(err)
		}

		for _, file := range files {
			if filepath.Ext(file) == ".vm" {
				processFile(file, w)
			}
		}
	} else {
		processFile(path, w)
	}
}

func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil{
		return false, err
	}
	return fileInfo.IsDir(), err
}

func createOutputPath(path string, isDir bool) string {
	dirname, filename := filepath.Split(path)
	outFilename := strings.Replace(filename, ".vm", "", 1)
	var outPath string
	if isDir {
		outPath = path + "/" + filename + ".asm"
	} else {
		outPath = dirname + outFilename + ".asm"
	}
	return outPath
}

func processFile(path string, w *bufio.Writer) {
	fname := filepath.Base(path)
	fnameNoExt := strings.Replace(fname, ".vm", "", 1)
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	var p parser.VmParser
	var aw translater.Translater
	equalityCheckCount := 0
	functionName := ""

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

		if commandType == "C_FUNCTION" {
			functionName = firstArg
		}

		aw = &translater.AssemblyWriter{
			Filename:    fnameNoExt,
			FunctionName: functionName,
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
		case "C_FUNCTION":
			assemblyCode = aw.WriteFunction()
		case "C_RETURN":
			assemblyCode = aw.WriteReturn()
		case "C_CALL":
			assemblyCode, equalityInc = aw.WriteCall(equalityCheckCount)
		}

		equalityCheckCount += equalityInc

		w.WriteString(assemblyCode)
	}
}
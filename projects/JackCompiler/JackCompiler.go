package JackCompiler

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
)

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

	compilationEngine := newCompilationEngine(file, writer)
	compilationEngine.compileClass()
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

// Compile takes a path to a folder or a file and compiles the .jack files/file
// into an xml document defining the grammar and structure of the jack code.
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

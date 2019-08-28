package main

import (
	"bufio"
	"compiler/tokenizer"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	args := os.Args[1:]

	path := filepath.Clean(args[0])

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	scanner := bufio.NewScanner(file)
	scanner.Split(tokenizer.ScanTokens)

	for scanner.Scan() {
		text := scanner.Text()
		fmt.Println(text)
	}
}

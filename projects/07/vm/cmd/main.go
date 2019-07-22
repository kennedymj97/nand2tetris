package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func main() {
	args := os.Args[1:]
	file, err := os.Open(args[0])
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}

package main

import (
	"os"

	"example.com/compiler"
)

func main() {
	args := os.Args[1:]
	compiler.Compile(args[0])
}

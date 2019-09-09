package main

import (
	"compiler"
	"os"
)

func main() {
	args := os.Args[1:]
	compiler.Compile(args[0])
}

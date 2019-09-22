package main

import (
	"JackCompiler"
	"os"
)

func main() {
	args := os.Args[1:]
	JackCompiler.Compile(args[0])
}

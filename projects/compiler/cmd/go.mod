module main

go 1.12

require (
	compiler/engine v0.0.0
	compiler/tokenizer v0.0.0
)

replace (
	compiler/engine => ../engine
	compiler/tokenizer => ../tokenizer
)

module main

go 1.12

require (
	vm/parser v0.0.0
	vm/translater v0.0.0
)

replace (
	vm/parser => ../parser
	vm/translater => ../translater
)

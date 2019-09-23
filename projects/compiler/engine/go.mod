module engine

go 1.13

require (
	example.com/cache v0.0.0
	example.com/tokenizer v0.0.0
	example.com/writer v0.0.0
)

replace (
	example.com/cache => ../cache
	example.com/tokenizer => ../tokenizer
	example.com/writer => ../writer
)

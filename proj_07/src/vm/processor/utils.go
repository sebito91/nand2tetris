package processor

// stack is 256  - 2047
var stack [1791]int

const (
	// ARGUMENT ...
	ARGUMENT = "argument"

	// TEMP ...
	TEMP = "temp"

	// THIS ...
	THIS = "this"

	// THAT ...
	THAT = "that"

	// LOCAL ...
	LOCAL = "local"

	// STATIC ...
	STATIC = "static"

	// POINTER ...
	POINTER = "pointer"
)

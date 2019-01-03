package tokenizer

import (
	"regexp"
)

// Token is the default struct for each token discovered
// within our .jack files
type Token struct {
	Type  string
	Value string
}

var re = regexp.MustCompile("(^[a-zA-Z_]{1}[a-zA-Z0-9_]*)")

const (
	// token types
	KEYWORD = iota
	SYMBOL
	IDENTIFIER
	INTCONST
	STRINGCONST
	UNKNOWN
)

// keywords is the global set of keywords within the Jack language
var keywords = map[string]int{
	"class":       KEYWORD,
	"method":      KEYWORD,
	"function":    KEYWORD,
	"constructor": KEYWORD,
	"int":         KEYWORD,
	"boolean":     KEYWORD,
	"char":        KEYWORD,
	"void":        KEYWORD,
	"var":         KEYWORD,
	"static":      KEYWORD,
	"field":       KEYWORD,
	"let":         KEYWORD,
	"do":          KEYWORD,
	"if":          KEYWORD,
	"else":        KEYWORD,
	"while":       KEYWORD,
	"return":      KEYWORD,
	"true":        KEYWORD,
	"false":       KEYWORD,
	"null":        KEYWORD,
	"this":        KEYWORD,
}

// symbols is the global set of symbols within the Jack language
var symbols = map[string]int{
	"{": SYMBOL,
	"}": SYMBOL,
	"(": SYMBOL,
	")": SYMBOL,
	"[": SYMBOL,
	"]": SYMBOL,
	".": SYMBOL,
	",": SYMBOL,
	";": SYMBOL,
	"+": SYMBOL,
	"-": SYMBOL,
	"*": SYMBOL,
	"/": SYMBOL,
	"&": SYMBOL,
	"|": SYMBOL,
	"<": SYMBOL,
	">": SYMBOL,
	"=": SYMBOL,
	"~": SYMBOL,
}

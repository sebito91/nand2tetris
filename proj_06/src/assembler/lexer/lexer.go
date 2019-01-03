// Package lexer handles the data manipulation to read and process each line of the provided code
package lexer

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"unicode"
)

// symbols are the mapping of symbols in our assembly code
var symbols map[string]map[string]int

// commonSymbols are commonly found, reserved symbol names
var commonSymbols map[string]int

// addr is a local variable reserved for incrementing the label counter
var addr int

func init() {
	symbols = make(map[string]map[string]int)
	addr = 16 // start at register 15

	commonSymbols = map[string]int{
		"R0":     0,
		"R1":     1,
		"R2":     2,
		"R3":     3,
		"R4":     4,
		"R5":     5,
		"R6":     6,
		"R7":     7,
		"R8":     8,
		"R9":     9,
		"R10":    10,
		"R11":    11,
		"R12":    12,
		"R13":    13,
		"R14":    14,
		"R15":    15,
		"SCREEN": 16384,
		"KBD":    24576,
		"SP":     0,
		"LCL":    1,
		"ARG":    2,
		"THIS":   3,
		"THAT":   4,
	}
}

// stripSpaces maps each character (rune) of the incoming string against a function
// that checks for whitespace. Since we've already removed any blank lines, this function
// simply checks for anything that's considered a utf-8 'space'
func stripSpaces(in string) []byte {
	return bytes.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, []byte(in))
}

// ProcessFiles takes in a list of filename and strips the file of spaces, writing to filename.out
func ProcessFiles(filenames ...string) (map[string][][]byte, error) {
	m := make(map[string][][]byte)

	for _, f := range filenames {
		var count int
		sym := make(map[string]int)

		read, err := ioutil.ReadFile(f)
		if err != nil {
			return nil, err
		}

		var out [][]byte
		for id, x := range strings.Split(fmt.Sprintf("%s", read), "\n") {
			if strings.HasPrefix(x, "//") {
				count++
				continue
			}

			val := stripSpaces(x)
			if len(val) == 0 {
				count++
				continue
			}

			// remove inline comments
			lines := bytes.Split(val, []byte("//"))
			line := lines[0]

			// create initial symbols for loops, jumps
			if bytes.HasPrefix(line, []byte("(")) {
				sym[fmt.Sprintf("%s", bytes.Trim(line, "()"))] = id - count
				count++
				continue
			}

			out = append(out, line)
		}

		// second pass to add @... to symbol table
		for _, x := range out {
			b := bytes.TrimPrefix(x, []byte("@"))
			if bytes.EqualFold(x, b) || len(b) == 0 {
				continue
			}

			if unicode.IsDigit(rune(b[0])) {
				continue
			}

			updateSymbol(sym, fmt.Sprintf("%s", b))
		}

		bits := strings.Split(filepath.Base(f), ".")
		m[bits[0]] = out
		symbols[bits[0]] = sym
	}

	return m, nil
}

// FetchSymbols returns the mapping of symbols in our assembly code
func FetchSymbols() map[string]map[string]int {
	return symbols
}

// updateSymbol takes in the map and symbol to check for existence
func updateSymbol(sym map[string]int, s string) {
	if _, ok := commonSymbols[s]; ok {
		return
	}

	if _, ok := sym[s]; ok {
		return
	}

	sym[s] = addr
	addr++
}

// CheckSymbol returns the corresponding location value for the provided symbol
func CheckSymbol(f, s string) (int, bool) {
	var val int
	var ok bool

	val, ok = commonSymbols[s]
	if ok {
		return val, ok
	}

	if _, ok := symbols[f]; !ok {
		return -1, false
	}

	val, ok = symbols[f][s]
	return val, ok
}

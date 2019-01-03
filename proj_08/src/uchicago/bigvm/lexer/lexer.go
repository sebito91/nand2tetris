// Package lexer handles the data manipulation to read and process each line of the provided code
package lexer

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

// symbols are the mapping of symbols in our assembly code
var symbols map[string]map[string]int

// commonSymbols are commonly found, reserved symbol names
var commonSymbols map[string]int

// addr is a local variable reserved for incrementing the label counter
var addr int

var validLabel = regexp.MustCompile("label ([a-zA-Z:_.]{1}[a-zA-Z0-9_.:]+)")
var funcLabel = regexp.MustCompile("function ([a-zA-Z:_.]{1}[a-zA-Z0-9_.:]+)")

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

// ProcessFiles takes in a list of filename and strips the file of spaces, writing to filename.out
func ProcessFiles(filenames ...string) (map[string][][]byte, error) {
	m := make(map[string][][]byte)

	for _, f := range filenames {
		sym := make(map[string]int)

		read, err := ioutil.ReadFile(f)
		if err != nil {
			return nil, err
		}

		var out [][]byte
		for _, x := range strings.Split(fmt.Sprintf("%s", read), "\n") {
			x = strings.TrimSpace(x)
			if strings.HasPrefix(x, "//") {
				continue
			}

			// remove inline comments
			lines := strings.Split(x, "//")
			line := strings.TrimSpace(lines[0])

			if line == "" {
				continue
			}

			if strings.Contains(line, "label") {
				labels := strings.Split(line, " ")
				_, ok := sym[labels[1]]

				if !validLabel.MatchString(line) || ok {
					return nil, fmt.Errorf("received invalid or duplicate label: %s", labels[1])
				}
				sym[labels[1]] = 0
			}

			if strings.Contains(line, "function") {
				labels := strings.Split(line, " ")
				_, ok := sym[labels[1]]

				if !funcLabel.MatchString(line) || ok {
					return nil, fmt.Errorf("received invalid or duplicate label: %s", labels[1])
				}
				sym[labels[1]] = 0
			}

			out = append(out, []byte(line))
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

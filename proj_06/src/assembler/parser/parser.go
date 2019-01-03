// Package parser splits each line as needed
package parser

import (
	"assembler/lexer"
	"bytes"
	"fmt"
	"strconv"
	"unicode"
)

// ParseFile takes the filename and corresponding raw assembly language and processes the
// lines into valid Hack (aka binary)
func ParseFile(k string, v [][]byte) ([]byte, error) {
	// iterate each line
	var buf bytes.Buffer
	for _, x := range v {
		switch {
		case bytes.HasPrefix(x, []byte("@")):
			str, err := processAInstruction(k, x)
			if err != nil {
				continue
			}

			buf.WriteString(fmt.Sprintf("%s\n", str))
		default:
			str, err := processCInstruction(x)
			if err != nil {
				continue
			}
			buf.WriteString(fmt.Sprintf("%s\n", str))
		}
	}

	return buf.Bytes(), nil
}

// processAInstruction handles an "@..." instruction from our assembly language
func processAInstruction(file string, in []byte) (string, error) {
	b := bytes.TrimPrefix(in, []byte("@"))
	if bytes.EqualFold(in, b) || len(b) == 0 {
		return "", fmt.Errorf("invalid A instruction: %s", in)
	}

	if !unicode.IsDigit(rune(b[0])) {
		val, ok := lexer.CheckSymbol(file, fmt.Sprintf("%s", b))
		if !ok {
			return "", fmt.Errorf("invalid label for A instruction: %s", in)
		}

		return fmt.Sprintf("%016b", val), nil
	}

	// add a lookup to symbol table here!
	val, err := strconv.Atoi(fmt.Sprintf("%s", b))
	if err != nil {
		return "", fmt.Errorf("could not format A instruction (%s): %s", in, err.Error())
	}

	return fmt.Sprintf("%016b", val), nil
}

// processCInstruction handles the C-instructions from the assembly language
func processCInstruction(in []byte) (string, error) {
	if bytes.Contains(in, []byte(";J")) {
		return processJump(in)
	}

	val := bytes.Split(in, []byte("="))
	if len(val) < 2 {
		return "", fmt.Errorf("invalid C instruction: %s", in)
	}

	d := val[0] // destination
	f := val[1] // function
	a := "1"

	if !bytes.Contains(f, []byte("M")) {
		a = "0"
	}

	return fmt.Sprintf("111%s%s%s000", a, checkFunc(f), checkDest(d)), nil
}

// processJump handles the jump portion of the C instruction
func processJump(in []byte) (string, error) {
	val := bytes.Split(in, []byte(";"))
	if len(val) < 2 {
		return "", fmt.Errorf("invalid C instruction: %s", in)
	}

	f := val[0] // destination
	j := val[1] // function
	a := "0"

	if bytes.Contains(f, []byte("M")) {
		a = "1"
	}

	return fmt.Sprintf("111%s%s000%s", a, checkFunc(f), checkJump(j)), nil
}

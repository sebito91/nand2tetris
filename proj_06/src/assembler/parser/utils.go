package parser

// dests is the list of matching Destination mnemonics from our assembly language
var dests map[string][]byte

// funcs is the list of the matching ALU functions from our assembly language
var funcs map[string][]byte

// jumps is the list of the jump options from our assembly language
var jumps map[string][]byte

func init() {
	dests = initDestinations()
	funcs = initFunctions()
	jumps = initJumps()
}

// initDestinations returns a new instance of the map of destinations
func initDestinations() map[string][]byte {
	return map[string][]byte{
		"null": []byte("000"),
		"M":    []byte("001"),
		"D":    []byte("010"),
		"MD":   []byte("011"),
		"A":    []byte("100"),
		"AM":   []byte("101"),
		"AD":   []byte("110"),
		"AMD":  []byte("111"),
	}
}

// initFunctions returns a new instance of the map of ALU functions
func initFunctions() map[string][]byte {
	return map[string][]byte{
		"0":   []byte("101010"),
		"1":   []byte("111111"),
		"-1":  []byte("111010"),
		"D":   []byte("001100"),
		"A":   []byte("110000"),
		"M":   []byte("110000"),
		"!D":  []byte("001101"),
		"!A":  []byte("110001"),
		"!M":  []byte("110001"),
		"-D":  []byte("001111"),
		"-A":  []byte("110011"),
		"-M":  []byte("110011"),
		"D+1": []byte("011111"),
		"A+1": []byte("110111"),
		"M+1": []byte("110111"),
		"D-1": []byte("001110"),
		"A-1": []byte("110010"),
		"M-1": []byte("110010"),
		"D+A": []byte("000010"),
		"A+D": []byte("000010"),
		"D+M": []byte("000010"),
		"D-A": []byte("010011"),
		"D-M": []byte("010011"),
		"A-D": []byte("000111"),
		"M-D": []byte("000111"),
		"D&A": []byte("000000"),
		"D&M": []byte("000000"),
		"D|A": []byte("010101"),
		"D|M": []byte("010101"),
	}
}

// initJumps returns a new instance of the map of jump options
func initJumps() map[string][]byte {
	return map[string][]byte{
		"null": []byte("000"),
		"JGT":  []byte("001"),
		"JEQ":  []byte("010"),
		"JGE":  []byte("011"),
		"JLT":  []byte("100"),
		"JNE":  []byte("101"),
		"JLE":  []byte("110"),
		"JMP":  []byte("111"),
	}
}

// checkDest returns the destination of the given instructions
func checkDest(d []byte) []byte {
	out, ok := dests[string(d)]
	if !ok {
		return dests["null"]
	}
	return out
}

// checkFunc returns the function of the given instructions
func checkFunc(f []byte) []byte {
	out, ok := funcs[string(f)]
	if !ok {
		return funcs["null"]
	}

	return out
}

// checkJump returns the jump of the given instructions
func checkJump(j []byte) []byte {
	out, ok := jumps[string(j)]
	if !ok {
		return jumps["null"]
	}

	return out
}

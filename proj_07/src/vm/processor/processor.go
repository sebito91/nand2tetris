// Package processor splits each line as needed
package processor

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// jumpCounter is a horrible hack to iterate the number of jumps we encounter throughout
var jumpCounter int

// ParseFile takes the filename and corresponding raw assembly language and processes the
// lines into valid Hack (aka binary)
//
// add    (x + y), integer add
// sub    (x - y), integer sub
// neg    (-y), negate
// eq     true if x == y, else false
// gt     true if x > y, else false
// lt     true if x < y, else false
// and    x && y, bitwise
// or     x || y, bitwise
// not    !x, bitwise
func ParseFile(k string, v [][]byte) ([]byte, error) {
	var sp int // stack pointer

	// iterate each line
	var buf bytes.Buffer
	for _, x := range v {
		var out string
		var err error

		bits := strings.Split(fmt.Sprintf("%s", x), " ")
		if len(bits) == 0 {
			return buf.Bytes(), fmt.Errorf("received invalid instruction: %s", x)
		}

		switch bits[0] {
		case "push":
			sp, out, err = handlePush(sp, bits)
			if err != nil {
				return buf.Bytes(), fmt.Errorf("received invalid push command: %s", x)
			}
		case "pop":
			sp, out, err = handlePop(sp, bits)
			if err != nil {
				return buf.Bytes(), fmt.Errorf("received invalid pop command: %s", x)
			}
		case "add", "sub", "and", "or":
			sp, out, err = handleBinary(sp, bits)
			if err != nil {
				return buf.Bytes(), fmt.Errorf("received invalid binary command: %s", x)
			}
		case "not", "neg":
			sp, out, err = handleUnary(sp, bits)
			if err != nil {
				return buf.Bytes(), fmt.Errorf("received invalid unary command: %s", x)
			}
		case "gt", "lt", "eq":
			sp, out, err = handleJump(sp, bits)
			if err != nil {
				return buf.Bytes(), fmt.Errorf("received invalid jump command: %s", x)
			}
		default:
			return []byte{}, fmt.Errorf("received invalid command: %s", x)
		}
		buf.WriteString(fmt.Sprintf("%s\n", out))
	}

	// add in the endless loop for the end
	buf.WriteString(strings.Join([]string{
		"(END)", // create (END) label
		"@END",  // load the addr for the END label
		"0;JMP", // jump automagically back up, endless loop
		"",
	}, "\n"))
	return buf.Bytes(), nil
}

func handlePop(sp int, bits []string) (int, string, error) {
	var out []string
	var post = []string{
		"@R13",
		"M=D",
		"@SP",
		"AM=M-1",
		"D=M",
		"@R13",
		"A=M",
		"M=D",
	}

	switch bits[1] {
	case ARGUMENT, LOCAL, THIS, THAT:
		if bits[1] == ARGUMENT {
			out = append(out, "@ARG")
		} else if bits[1] == LOCAL {
			out = append(out, "@LCL")
		} else {
			out = append(out, fmt.Sprintf("@%s", strings.ToUpper(bits[1])))
		}

		out = append(out, "D=M")
		out = append(out, fmt.Sprintf("@%s", bits[2]))
		out = append(out, "D=D+A")
		out = append(out, post...)
	case TEMP:
		out = append(out, "@R5")
		val, err := strconv.Atoi(bits[2])
		if err != nil {
			return sp, "", fmt.Errorf("incorrect value for pop temp statement: %s", bits[2])
		}
		val += 5
		out = append(out, "D=M")
		out = append(out, fmt.Sprintf("@%d", val))
		out = append(out, "D=D+A")
		out = append(out, post...)
	case STATIC:
		val, err := strconv.Atoi(bits[2])
		if err != nil {
			return sp, "", fmt.Errorf("invalid constant passed to static pop: %s", bits[2])
		}
		val += 16
		out = append(out, fmt.Sprintf("@STATIC.%d", val))
		out = append(out, "D=A")
		out = append(out, post...)
	case POINTER:
		val, err := strconv.Atoi(bits[2])
		if err != nil {
			return sp, "", fmt.Errorf("invalid constant passed to pointer pop: %s", bits[2])
		}
		val += 3
		out = append(out, fmt.Sprintf("@%d", val))
		out = append(out, "D=A")
		out = append(out, post...)
	default:
		return sp, "", fmt.Errorf("incorrect pop statement: %s", bits[1])
	}
	sp--

	return sp, strings.Join(out, "\n"), nil
}

func handlePush(sp int, bits []string) (int, string, error) {
	var out []string
	var post = []string{
		"@SP",
		"A=M",
		"M=D",
		"@SP",
		"M=M+1",
	}

	switch bits[1] {
	case ARGUMENT, LOCAL, THIS, THAT, POINTER:
		if bits[1] == ARGUMENT {
			out = append(out, "@ARG")
		} else if bits[1] == LOCAL {
			out = append(out, "@LCL")
		} else {
			out = append(out, fmt.Sprintf("@%s", strings.ToUpper(bits[1])))
		}

		out = append(out, "D=M")
		if bits[1] == POINTER {
			val, err := strconv.Atoi(bits[2])
			if err != nil {
				return sp, "", fmt.Errorf("invalid param passed to pointer push: %s", bits[2])
			}
			val += 3
			out = append(out, fmt.Sprintf("@%d", val))
		} else {
			out = append(out, fmt.Sprintf("@%s", bits[2]))
		}
		out = append(out, "A=D+A")
		out = append(out, "D=M")
		out = append(out, post...)
	case TEMP:
		out = append(out, "@R5")
		val, err := strconv.Atoi(bits[2])
		if err != nil {
			return sp, "", fmt.Errorf("incorrect value for pop temp statement: %s", bits[2])
		}
		val += 5
		out = append(out, "D=M")
		out = append(out, fmt.Sprintf("@%d", val))
		out = append(out, "A=D+A")
		out = append(out, "D=M")
		out = append(out, post...)
	case "constant":
		v, err := strconv.Atoi(bits[2])
		if err != nil {
			return sp, "", fmt.Errorf("invalid constant provided: %s", bits[2])
		}

		stack[sp] = v
		out = []string{
			fmt.Sprintf("@%s", bits[2]),
			"D=A", // load the const into the D-reg
			"@SP", // load the SP into the A-reg
			"A=M",
			"M=D", // load the D-reg into the mem location currently pointed to by SP
			"@SP", // reload the SP and increment
			"M=M+1",
		}
	case STATIC:
		val, err := strconv.Atoi(bits[2])
		if err != nil {
			return sp, "", fmt.Errorf("incorrect constant provided to static push: %s", bits[2])
		}
		val += 16 // get it beyond the R0-R15 registers
		out = append(out, fmt.Sprintf("@STATIC.%d", val))
		out = append(out, "D=M")
		out = append(out, post...)
	default:
		return sp, "", fmt.Errorf("incorrect push statement: %s", bits[1])
	}
	sp++

	return sp, strings.Join(out, "\n"), nil
}

func handleBinary(sp int, bits []string) (int, string, error) {
	out := []string{
		"@SP",    // load the SP into both A-reg and mem
		"AM=M-1", // decrement the SP by one and load the value there into M, SP goes to A-reg (y-value)
		"D=M",    // assign the current M to the D-reg
		"A=A-1",  // decrement the SP once more and consume the value at that value into the A-reg (x-value)
	}

	switch bits[0] {
	case "add":
		out = append(out, "M=M+D")
	case "sub":
		out = append(out, "M=M-D")
	case "and":
		out = append(out, "M=M&D")
	case "or":
		out = append(out, "M=M|D")
	default:
		return sp, "", fmt.Errorf("received invalid command: %s", bits[0])
	}
	sp++
	return sp, strings.Join(out, "\n"), nil
}

func handleUnary(sp int, bits []string) (int, string, error) {
	var out = []string{
		"@SP",
		"A=M-1",
	}

	switch bits[0] {
	case "not":
		out = append(out, "M=!M")
	case "neg":
		out = append(out, "D=0")
		out = append(out, "M=D-M")
	default:
		return sp, "", fmt.Errorf("received invalid unary command: %s", bits[0])
	}

	sp++
	return sp, strings.Join(out, "\n"), nil
}

func handleJump(sp int, bits []string) (int, string, error) {
	// NOTE: I'm not sure this works properly, but I can't think of any other solution than to create multiple
	//       versions of the same "if-block" logic when these jumps are repeated. I'm sure there's a better way
	//       but since we're under the gun here no time to debate.
	//
	//       If we had a bit more time I'd like to investigate the `END_GT`, `END_EQ`, etc parts from Pong.asm
	//
	// NEWNOTE: i wracked my brain on this one and for some reason only the inverted logic was working. I'm not
	//          100% sure but will need to revisit for next week

	var jmp string
	switch bits[0] {
	case "gt":
		jmp = fmt.Sprintf("D;JLE")
	case "eq":
		jmp = fmt.Sprintf("D;JNE")
	case "lt":
		jmp = fmt.Sprintf("D;JGE")
	default:
		return sp, "", fmt.Errorf("received invalid jump command: %s", bits[0])
	}

	out := []string{
		"@SP",    // load the SP into both A-reg and mem
		"AM=M-1", // decrement the SP by one and load the value there into M, SP goes to A-reg (y-value)
		"D=M",    // assign the current value at memory location M to the D-reg
		"A=A-1",  // decrement the SP once more and consume the value at that memory location into the A-reg (x-value)
		"D=M-D",  // calculate the difference between
		fmt.Sprintf("@JUMP.CHECK.%d", jumpCounter), // horrible hack to add a jump for this specific condition, adds a lot of repetition sadly
		jmp, // the jump we need to make, aka `gt`, `eq`, `lt`
		"@SP",
		"A=M-1", // decrement the stack pointer and drop -1 if not matching the condition,
		"M=-1",  // drop the -1 for false condition
		fmt.Sprintf("@JUMP.CONTINUE.%d", jumpCounter), // create the continuation point
		"0;JMP", // jump unconditionally here to continue
		fmt.Sprintf("(JUMP.CHECK.%d)", jumpCounter), // if we did make the condition (aka true), then mark the SP location as 0
		"@SP",
		"A=M-1",
		"M=0",
		fmt.Sprintf("(JUMP.CONTINUE.%d)", jumpCounter), // continuation point
	}

	jumpCounter++
	sp++

	return sp, strings.Join(out, "\n"), nil
}

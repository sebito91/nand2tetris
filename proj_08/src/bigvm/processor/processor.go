// Package processor splits each line as needed
package processor

import (
	"bigvm/lexer"
	"bytes"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"
)

// jumpCounter is a horrible hack to iterate the number of jumps we encounter throughout
var jumpCounter int

// retCounter is a horrible hack to iterate the number of returns we encounter throughout
var retCounter int

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
//
// label
// if-goto
// goto
// function
// call
// return
func ParseFile(fp, k string, v [][]byte) ([]byte, error) {
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
			sp, out, err = handlePush(sp, fp, bits)
			if err != nil {
				return buf.Bytes(), fmt.Errorf("received invalid push command (%s): %s", x, err.Error())
			}
		case "pop":
			sp, out, err = handlePop(sp, fp, bits)
			if err != nil {
				return buf.Bytes(), fmt.Errorf("received invalid pop command (%s): %s", x, err.Error())
			}
		case "add", "sub", "and", "or":
			sp, out, err = handleBinary(sp, bits)
			if err != nil {
				return buf.Bytes(), fmt.Errorf("received invalid binary command (%s): %s", x, err.Error())
			}
		case "not", "neg":
			sp, out, err = handleUnary(sp, bits)
			if err != nil {
				return buf.Bytes(), fmt.Errorf("received invalid unary command (%s): %s", x, err.Error())
			}
		case "gt", "lt", "eq":
			sp, out, err = handleJump(sp, bits)
			if err != nil {
				return buf.Bytes(), fmt.Errorf("received invalid jump command (%s): %s", x, err.Error())
			}
		case "label":
			sp, out, err = handleLabel(sp, k, bits)
			if err != nil {
				return buf.Bytes(), fmt.Errorf("received invalid label command (%s): %s", x, err.Error())
			}
		case "if-goto":
			sp, out, err = handleIfGoTo(sp, k, bits)
			if err != nil {
				return buf.Bytes(), fmt.Errorf("received invalid if-goto command (%s): %s", x, err.Error())
			}
		case "goto":
			sp, out, err = handleGoTo(sp, k, bits)
			if err != nil {
				return buf.Bytes(), fmt.Errorf("received invalid goto command (%s): %s", x, err.Error())
			}
		case "function":
			sp, out, err = handleFunction(sp, k, bits)
			if err != nil {
				return buf.Bytes(), fmt.Errorf("received invalid function command (%s): %s", x, err.Error())
			}
		case "return":
			sp, out, err = handleReturn(sp, k, bits)
			if err != nil {
				return buf.Bytes(), fmt.Errorf("received invalid return command (%s): %s", x, err.Error())
			}
		case "call":
			sp, out, err = handleCall(sp, k, bits)
			if err != nil {
				return buf.Bytes(), fmt.Errorf("received invalid call command (%s): %s", x, err.Error())
			}
		default:
			return []byte{}, fmt.Errorf("received invalid command bro: %s", x)
		}
		buf.WriteString(fmt.Sprintf("%s\n", out))
	}

	return buf.Bytes(), nil
}

// Init ...
func Init(f string) string {
	var out = []string{
		"@256    // START INIT",
		"D=A",
		"@SP",
		"M=D",
	}

	_, s, err := handleCall(0, f, []string{"call", "Sys.init", fmt.Sprintf("%d", 0)})
	if err != nil {
		log.Fatal(err)
	}

	out = append(out, fmt.Sprintf("%s    // END INIT", s))
	return strings.Join(out, "\n")
}

func handleCall(sp int, k string, bits []string) (int, string, error) {
	var err error

	ret := fmt.Sprintf("RETURN_%d", retCounter) // set up anew return address label and location
	var out = []string{
		fmt.Sprintf("@%s    // START CALL -- %s", ret, strings.Join(bits, " ")),
		"D=A",
		"@SP",
		"A=M",
		"M=D",
		"@SP",
		"M=M+1",
	}

	// push each of the LCL, ARG, THIS and THAT addresses
	for _, x := range []string{"LCL", "ARG", "THIS", "THAT"} {
		p := []string{
			fmt.Sprintf("@%s", x),
			"D=M",
			"@SP",
			"A=M",
			"M=D",
			"@SP",
			"M=M+1",
		}
		out = append(out, p...)
	}

	val, err := strconv.Atoi(bits[2])
	if err != nil {
		return sp, "", err
	}
	val += 5

	var post = []string{
		"@SP",
		"D=M",
		fmt.Sprintf("@%d", val),
		"D=D-A",
		"@ARG",
		"M=D",
		"@SP",
		"D=M",
		"@LCL",
		"M=D",
		fmt.Sprintf("@%s", bits[1]),
		"0;JMP",
		fmt.Sprintf("(%s)    // END CALL -- %s", ret, strings.Join(bits, " ")),
	}

	out = append(out, post...)
	retCounter++

	return sp, strings.Join(out, "\n"), nil
}

// NOTE: lifted the logic for the return straight from the text, pg 163
//       the logic for the FRAME uses the LCL
func handleReturn(sp int, k string, bits []string) (int, string, error) {
	_, s, err := handlePop(sp, "", []string{"pop", ARGUMENT, fmt.Sprintf("%d", 0)})
	if err != nil {
		return sp, "", err
	}

	var out = []string{
		fmt.Sprintf("@LCL    // START RETURN -- %s", strings.Join(bits, " ")), // frame is a temp var
		"D=M",
		"@FRAME", // put the return-address in atemp var
		"M=D",
		"@5", // *(FRAME-5)
		"D=A",
		"@FRAME",
		"D=M-D",
		"A=D",
		"D=M",
		"@RET",
		"M=D",
		s,
		"@ARG",
		"D=M",
		"@SP",
		"M=D+1",
	}

	// Restore THAT, THIS, ARG, LCL of the called
	for _, x := range []string{"THAT", "THIS", "ARG", "LCL"} {
		out = append(out, []string{
			"@FRAME",
			"M=M-1",
			"A=M",
			"D=M",
			fmt.Sprintf("@%s", x),
			"M=D",
		}...)
	}

	// goto return-address
	out = append(out, []string{
		"@RET",
		"A=M",
		fmt.Sprintf("0;JMP    // END RETURN -- %s", strings.Join(bits, " ")),
	}...)

	return sp, strings.Join(out, "\n"), nil
}

func handleFunction(sp int, k string, bits []string) (int, string, error) {
	if _, ok := lexer.CheckSymbol(k, bits[1]); !ok {
		return sp, "", fmt.Errorf("label %s not found", bits[1])
	}

	if len(bits) <= 2 {
		return sp, "", fmt.Errorf("could not process number of local vars for function: %s", bits[1])
	}

	d, err := strconv.Atoi(bits[2])
	if err != nil {
		return sp, "", fmt.Errorf("invalid number (%s) for variables in function: %s", bits[1], err.Error())
	}

	var out = []string{fmt.Sprintf("(%s)    // START FUNCTION -- %s", bits[1], strings.Join(bits, " "))}
	//	var s string
	for i := 0; i < d; i++ {
		out = append(out, []string{
			"@SP",
			"A=M",
			"M=0",
			"@SP",
			"M=M+1",
		}...)
	}

	return sp, strings.Join(out, "\n"), nil
}

func handleGoTo(sp int, k string, bits []string) (int, string, error) {
	if _, ok := lexer.CheckSymbol(k, bits[1]); !ok {
		return sp, "", fmt.Errorf("label %s not found", bits[1])
	}

	out := []string{
		fmt.Sprintf("@%s", bits[1]),
		"0;JMP",
	}

	return sp, strings.Join(out, "\n"), nil
}

func handleIfGoTo(sp int, k string, bits []string) (int, string, error) {
	if _, ok := lexer.CheckSymbol(k, bits[1]); !ok {
		return sp, "", fmt.Errorf("label %s not found", bits[1])
	}

	// decrement the stack and compare if the value != 0
	out := []string{
		fmt.Sprintf("@SP    // START IF-GOTO -- %s", strings.Join(bits, " ")),
		"AM=M-1",
		"D=M",
		fmt.Sprintf("@%s", bits[1]),
		fmt.Sprintf("D;JNE    // END IF-GOTO -- %s", strings.Join(bits, " ")),
	}

	return sp, strings.Join(out, "\n"), nil
}

func handleLabel(sp int, k string, bits []string) (int, string, error) {
	if _, ok := lexer.CheckSymbol(k, bits[1]); !ok {
		return sp, "", fmt.Errorf("label %s not found", bits[1])
	}

	return sp, fmt.Sprintf("(%s)    // HANDLE LABEL -- %s", bits[1], strings.Join(bits, " ")), nil
}

func handlePop(sp int, fp string, bits []string) (int, string, error) {
	var out []string
	var post = []string{
		"@R13",
		"M=D",
		"@SP",
		"AM=M-1",
		"D=M",
		"@R13",
		"A=M",
		fmt.Sprintf("M=D    // END POP -- %s", strings.Join(bits, " ")),
	}

	val, err := strconv.Atoi(bits[2])
	if err != nil {
		return sp, "", fmt.Errorf("incorrect value for pop %s statement: %s", bits[1], bits[2])
	}

	switch bits[1] {
	case ARGUMENT, LOCAL, THIS, THAT:
		out = []string{
			fmt.Sprintf("@%d", val),
			"D=A",
		}

		if bits[1] == ARGUMENT {
			out = append(out, fmt.Sprintf("@ARG    // START POP -- %s", strings.Join(bits, " ")))
		} else if bits[1] == LOCAL {
			out = append(out, fmt.Sprintf("@LCL    // START POP -- %s", strings.Join(bits, " ")))
		} else {
			out = append(out, fmt.Sprintf("@%s    // START POP -- %s", strings.ToUpper(bits[1]), strings.Join(bits, " ")))
		}

		out = append(out, "A=M")
		out = append(out, "D=A+D")
		out = append(out, post...)
	case TEMP:
		out = []string{
			fmt.Sprintf("@%d", val),
			"D=A",
			fmt.Sprintf("@R5    // START POP -- %s", strings.Join(bits, " ")),
			"D=A+D",
		}
		out = append(out, post...)
	case STATIC:
		f := strings.Split(filepath.Base(fp), ".")

		out = []string{
			fmt.Sprintf("@%s.%d    // START POP -- %s", f[0], val, strings.Join(bits, " ")),
			"D=A",
			"@R13",
			"M=D",
			"@SP",
			"AM=M-1",
			"D=M",
			"@R13",
			"A=M",
			"M=D",
		}
	case POINTER:
		out = []string{
			fmt.Sprintf("@%d", val),
			"D=A",
			"@3",
			"D=D+A",
		}
		out = append(out, post...)
	default:
		return sp, "", fmt.Errorf("incorrect pop statement: %s", bits[1])
	}
	sp--

	return sp, strings.Join(out, "\n"), nil
}

func handlePush(sp int, fp string, bits []string) (int, string, error) {
	var out []string
	var post = []string{
		"@SP",
		"A=M",
		"M=D",
		"@SP",
		fmt.Sprintf("M=M+1    // END PUSH -- %s", strings.Join(bits, " ")),
	}

	val, err := strconv.Atoi(strings.TrimSpace(bits[2]))
	if err != nil {
		return sp, "", fmt.Errorf("invalid param passed to %s push: %s", bits[1], bits[2])
	}

	switch bits[1] {
	case ARGUMENT, LOCAL, THIS, THAT:
		if bits[1] == ARGUMENT {
			out = append(out, fmt.Sprintf("@ARG    // START PUSH -- %s", strings.Join(bits, " ")))
		} else if bits[1] == LOCAL {
			out = append(out, fmt.Sprintf("@LCL    // START PUSH -- %s", strings.Join(bits, " ")))
		} else {
			out = append(out, fmt.Sprintf("@%s    // START PUSH -- %s", strings.ToUpper(bits[1]), strings.Join(bits, " ")))
		}

		out = append(out, []string{
			"D=M",
			fmt.Sprintf("@%d", val),
			"A=D+A",
			"D=M",
		}...)

		out = append(out, post...)
	case POINTER:
		out = []string{
			fmt.Sprintf("@%d    // START PUSH -- %s", val, strings.Join(bits, " ")),
			"D=A",
			"@3",
			"A=A+D",
			"D=M",
		}
		out = append(out, post...)
	case TEMP:
		out = []string{
			fmt.Sprintf("@%d", val),
			"D=A",
			fmt.Sprintf("@R5    // START PUSH -- %s", strings.Join(bits, " ")),
			"A=A+D",
			"D=M",
		}
		out = append(out, post...)
	case "constant":
		out = []string{
			fmt.Sprintf("@%d    // START PUSH -- %s", val, strings.Join(bits, " ")),
			"D=A", // load the const into the D-reg
		}
		out = append(out, post...)
	case STATIC:
		f := strings.Split(filepath.Base(fp), ".")

		out = []string{
			fmt.Sprintf("@%s.%d    // START PUSH -- %s", f[0], val, strings.Join(bits, " ")),
			"D=M",
			"@SP",
			"A=M",
			"M=D",
			"@SP",
			"M=M+1",
		}
	default:
		return sp, "", fmt.Errorf("incorrect push statement: %s", bits[1])
	}
	sp++

	return sp, strings.Join(out, "\n"), nil
}

func handleBinary(sp int, bits []string) (int, string, error) {
	out := []string{
		"@SP",
		"AM=M-1",
		"D=M",
		"A=A-1",
	}

	switch bits[0] {
	case "add":
		out = append(out, fmt.Sprintf("M=M+D    // END BINARY -- %s", strings.Join(bits, " ")))
	case "sub":
		out = append(out, fmt.Sprintf("M=M-D    // END BINARY -- %s", strings.Join(bits, " ")))
	case "and":
		out = append(out, fmt.Sprintf("M=M&D    // END BINARY -- %s", strings.Join(bits, " ")))
	case "or":
		out = append(out, fmt.Sprintf("M=M|D    // END BINARY -- %s", strings.Join(bits, " ")))
	default:
		return sp, "", fmt.Errorf("received invalid binary command: %s", bits[0])
	}
	sp++
	return sp, strings.Join(out, "\n"), nil
}

func handleUnary(sp int, bits []string) (int, string, error) {
	var out = []string{
		fmt.Sprintf("@SP    // START UNARY -- %s", strings.Join(bits, " ")),
		"A=M-1",
	}

	switch bits[0] {
	case "not":
		out = append(out, fmt.Sprintf("M=!M    // END UNARY -- %s", strings.Join(bits, " ")))
	case "neg":
		out = append(out, "D=0")
		out = append(out, fmt.Sprintf("M=D-M    // END UNARY -- %s", strings.Join(bits, " ")))
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
		fmt.Sprintf("@SP    // START JUMP -- %s", strings.Join(bits, " ")), // load the SP into both A-reg and mem
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
		fmt.Sprintf("(JUMP.CONTINUE.%d)    // END JUMP -- %s", jumpCounter, strings.Join(bits, " ")), // continuation point
	}

	jumpCounter++
	sp++

	return sp, strings.Join(out, "\n"), nil
}

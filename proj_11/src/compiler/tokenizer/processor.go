// Package tokenizer splits each line as needed
package tokenizer

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// Tokenize takes the filename and corresponding raw Jack language and processes the
// lines into valid tokens for the parser to handle (precursor to proper Jack XML)
func Tokenize(fp, k string, v [][]byte) ([]byte, error) {
	var buf bytes.Buffer
	var ok bool

	for _, x := range v {
		if bytes.Contains(x, []byte("\"")) {
			x, ok = checkSpace(fmt.Sprintf("%s", x))
			if !ok {
				return buf.Bytes(), fmt.Errorf("could not match string quotes: %s", x)
			}
		}

		// split on space first
		bits := strings.Split(fmt.Sprintf("%s", x), " ")
		if len(bits) == 0 {
			return buf.Bytes(), fmt.Errorf("received invalid instruction: %s", x)
		}

		out, err := getTypes(bits)
		if err != nil {
			return buf.Bytes(), err
		}

		buf.WriteString(fmt.Sprintf("%s\n", strings.Join(out, "\n")))
	}

	return buf.Bytes(), nil
}

func checkSpace(x string) ([]byte, bool) {
	start := strings.Index(x, "\"")
	end := strings.Index(x[start+1:], "\"")
	substr := x[start : start+end+2]

	vals := []string{
		x[:start],
		strings.Replace(substr, " ", "=*=", -1),
		x[start+end+2:],
	}

	return []byte(strings.Join(vals, " ")), true
}

// token contains
func splitSemicolon(x string) ([]string, error) {
	var out []string

	idx := strings.LastIndex(x, ";")
	if idx != len(x)-1 {
		out = append(out, x)
		return out, nil
	}

	d := strings.Split(x, ";")
	d = append(d, ";")

	v, err := splitElements(d)
	if err != nil {
		return out, err
	}

	out = append(out, v...)
	return out, nil
}

func splitPunct(x, s string) ([]string, error) {
	var out []string

	d := []string{
		x[:strings.Index(x, s)],
		s,
	}

	if len(x[strings.Index(x, s)+1:]) > 0 {
		d = append(d, x[strings.Index(x, s)+1:])
	}

	v, err := splitElements(d)
	if err != nil {
		return out, err
	}

	out = append(out, v...)
	return out, nil
}

func splitElements(bits []string) ([]string, error) {
	var v, out []string
	var err error

	for _, x := range bits {
		// check if we are a terminating component
		switch {
		case len(x) == 1 && x != "":
			out = append(out, x)
			continue
		case strings.Contains(x, ";"):
			v, err = splitSemicolon(x)
		case strings.Contains(x, "."):
			v, err = splitPunct(x, ".")
		case strings.Contains(x, ","):
			v, err = splitPunct(x, ",")
		case strings.Contains(x, "["):
			v, err = splitPunct(x, "[")
		case strings.Contains(x, "("):
			v, err = splitPunct(x, "(")
		case strings.Contains(x, "]"):
			v, err = splitPunct(x, "]")
		case strings.Contains(x, ")"):
			v, err = splitPunct(x, ")")
		case strings.Contains(x, "-") && len(x) > 1:
			v, err = splitPunct(x, "-")
		case strings.Contains(x, "~") && len(x) > 1:
			v, err = splitPunct(x, "~")
		case x == "":
			continue
		default:
			out = append(out, x)
			continue
		}

		if err != nil {
			return out, err
		}

		out = append(out, v...)
	}

	return out, nil
}

// getTypes splits out the instruction into its individual parts
func getTypes(bits []string) ([]string, error) {
	var out []string

	check, err := splitElements(bits)
	if err != nil {
		return out, err
	}

	for _, x := range check {
		s, err := checkType(x)
		if err != nil {
			return out, err
		}

		switch s {
		case KEYWORD:
			out = append(out, fmt.Sprintf("<keyword> %s </keyword>", x))
		case SYMBOL:
			switch x {
			case "<":
				x = "&lt;"
			case ">":
				x = "&gt;"
			case "&":
				x = "&amp;"
			}

			out = append(out, fmt.Sprintf("<symbol> %s </symbol>", x))
		case INTCONST:
			out = append(out, fmt.Sprintf("<integerConstant> %s </integerConstant>", x))
		case STRINGCONST:
			x = strings.Replace(x, "=*=", " ", -1)
			x = strings.Trim(x, "\"")

			out = append(out, fmt.Sprintf("<stringConstant> %s </stringConstant>", x))
		case IDENTIFIER:
			out = append(out, fmt.Sprintf("<identifier> %s </identifier>", x))
		default:
		}
	}

	return out, nil
}

// checkType returns the type of the given token
func checkType(in string) (int, error) {
	if _, ok := keywords[in]; ok {
		return KEYWORD, nil
	}

	if _, ok := symbols[in]; ok {
		return SYMBOL, nil
	}

	if _, err := strconv.ParseInt(in, 10, 64); err == nil {
		return INTCONST, nil
	}

	if strings.HasPrefix(in, "\"") && strings.HasSuffix(in, "\"") {
		return STRINGCONST, nil
	}

	if re.MatchString(in) {
		return IDENTIFIER, nil
	}

	return UNKNOWN, fmt.Errorf("could not process value %s", in)
}

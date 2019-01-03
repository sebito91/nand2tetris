package parser

import (
	"bytes"
	"fmt"
	"strings"
)

// Compiled is a struct containing the raw data and parsed, compiled output
type Compiled struct {
	raw  []string // raw tokens
	d    []string // xml tokens
	vm   []string // vm tokens
	curr int
	max  int

	class       *symbolTable
	subroutines map[string]*symbolTable
}

// Parse receives the tokenized syntax and processes the data into valid
// Jack Virtual Machine XML
func Parse(data []byte) ([]byte, []byte, error) {
	var err error

	raw, err := convertData(data)
	if err != nil {
		return []byte{}, []byte{}, err
	}

	c := &Compiled{
		raw:  raw,
		curr: 0,
		max:  len(raw),
		class: &symbolTable{
			counts: map[string]int{
				STATIC:   0,
				THIS:     0,
				ARGUMENT: 0,
				LOCAL:    0,
			},
		},
		subroutines: make(map[string]*symbolTable),
	}

	return c.compileClass()
}

// convertData is a special handler to convert incoming byte-slices to strings
func convertData(data []byte) ([]string, error) {
	var s []string

	for _, x := range bytes.Split(data, []byte("\n")) {
		s = append(s, strings.TrimSpace(fmt.Sprintf("%s", x)))
	}

	return s, nil
}

func (c *Compiled) compileClass() ([]byte, []byte, error) {
	c.d = append(c.d, "<class>")
	c.d = append(c.d, c.handleClass()...) // only called once since we cannot have multiple class definitions
	c.d = append(c.d, c.handleClassVar()...)
	c.d = append(c.d, c.handleSubroutine()...)
	c.d = append(c.d, c.raw[c.curr:]...) // append the remaining instructions
	c.d = append(c.d, "</class>")

	// build the symbol tables first
	if err := c.buildSymbolTables(); err != nil {
		return []byte{}, []byte{}, err
	}

	if err := c.buildVM(); err != nil {
		return []byte{}, []byte{}, err
	}

	var buf, vm bytes.Buffer
	for _, x := range c.d {
		if x == "" {
			continue
		}

		buf.WriteString(fmt.Sprintf("%s\n", x))
	}

	for _, x := range c.vm {
		if x == "" {
			continue
		}

		vm.WriteString(fmt.Sprintf("%s\n", x))
	}

	return buf.Bytes(), vm.Bytes(), nil
}

func (c *Compiled) handleTerm() []string {
	var out []string

	out = append(out, "<term>")
	out = append(out, c.raw[c.curr])
	c.curr++

	if isUnaryOp(c.raw[c.curr-1]) {
		if isExprListOpen(c.raw[c.curr]) {
			out = append(out, "<term>")
			out = append(out, c.raw[c.curr])
			c.curr++

			out = append(out, "<expression>")
			out = append(out, c.handleExpression()...)
			out = append(out, "</expression>")

			out = append(out, c.raw[c.curr])
			c.curr++

			out = append(out, "</term>")
			out = append(out, "</term>")
			return out
		}
		out = append(out, c.handleTerm()...)
	}

	if isExprListOpen(c.raw[c.curr-1]) {
		out = append(out, "<expression>")
		out = append(out, c.handleExpression()...)
		out = append(out, "</expression>")

		out = append(out, c.raw[c.curr])
		c.curr++
	}

	if isExprOpen(c.raw[c.curr]) {
		out = append(out, c.handleExpression()...)
	}

	// likely a function call...(className | varName) . subName '(' expression ')'
	if isPeriod(c.raw[c.curr]) {
		out = append(out, c.raw[c.curr]) // append the '.'
		c.curr++

		out = append(out, c.raw[c.curr]) // append the next ident
		c.curr++

		out = append(out, c.handleExpression()...)
	}

	if isOp(c.raw[c.curr]) {
		out = append(out, "</term>")
		out = append(out, c.raw[c.curr]) // append the Op
		c.curr++
		//
		if isExprListOpen(c.raw[c.curr]) {
			out = append(out, "<term>")
			out = append(out, c.raw[c.curr])
			c.curr++

			out = append(out, "<expression>")
			out = append(out, c.handleExpression()...)
			out = append(out, "</expression>")

			out = append(out, c.raw[c.curr])
			c.curr++

			out = append(out, "</term>")
			return out
		}

		out = append(out, c.handleTerm()...) // handle the next tem
		return out
	}

	if isComma(c.raw[c.curr]) {
		out = append(out, "</term>")
		out = append(out, "</expression>")
		out = append(out, c.raw[c.curr]) // append the Op
		c.curr++

		out = append(out, "<expression>")
		out = append(out, c.handleTerm()...)
		return out
	}

	out = append(out, "</term>")

	return out
}

func (c *Compiled) handleExpression() []string {
	var out []string
	var done bool

	for {
		// check if integerConstant or stringConstant and handle <term></term>
		if isIntConst(c.raw[c.curr]) || isStrConst(c.raw[c.curr]) || isKeywordConst(c.raw[c.curr]) {
			out = append(out, c.handleTerm()...)

			if isSemicolon(c.raw[c.curr]) {
				break
			}

			continue
		}

		if isUnaryOp(c.raw[c.curr]) {
			out = append(out, c.handleTerm()...)
			continue
		}

		// handle identifier
		if isIdentifier(c.raw[c.curr]) && !isPeriod(c.raw[c.curr-1]) {
			out = append(out, c.handleTerm()...)

			continue
		}

		if isExprListOpen(c.raw[c.curr]) {
			if isIdentifier(c.raw[c.curr-1]) {
				out = append(out, c.raw[c.curr])
				c.curr++

				// inside a subroutine call, expressionList time
				out = append(out, "<expressionList>")
				if !isExprListClose(c.raw[c.curr]) {
					out = append(out, "<expression>")
					out = append(out, c.handleExpression()...)
					out = append(out, "</expression>")
				}
				out = append(out, "</expressionList>")

				out = append(out, c.raw[c.curr])
				c.curr++
				continue
			}

			out = append(out, c.handleTerm()...)
			break
		}

		// array call, needs a list of items
		if isExprOpen(c.raw[c.curr]) {
			out = append(out, c.raw[c.curr])
			c.curr++

			out = append(out, "<expression>")
			out = append(out, c.handleExpression()...)
			out = append(out, "</expression>")

			out = append(out, c.raw[c.curr])
			c.curr++
			continue
		}

		if isExprListClose(c.raw[c.curr]) {
			break
		}

		if isExprClose(c.raw[c.curr]) {
			break
		}

		if isSemicolon(c.raw[c.curr]) {
			break
		}

		// add in case of identifier
		out = append(out, c.raw[c.curr])
		c.curr++

		if done {
			break
		}
	}

	return out
}

func (c *Compiled) handleLet() []string {
	var out []string
	var done bool

	if isLet(c.raw[c.curr]) {
		out = append(out, "<letStatement>")
		for {
			if strings.HasPrefix(c.raw[c.curr], "<symbol>") && (strings.Contains(c.raw[c.curr], "=") || strings.Contains(c.raw[c.curr], "[")) {
				out = append(out, c.raw[c.curr])
				c.curr++

				out = append(out, "<expression>")
				out = append(out, c.handleExpression()...)
				out = append(out, "</expression>")
			}

			if isSemicolon(c.raw[c.curr]) {
				done = true
			}

			out = append(out, c.raw[c.curr])
			c.curr++

			if done {
				break
			}
		}
	}

	if len(out) > 0 {
		out = append(out, "</letStatement>")
	}

	return out
}

func (c *Compiled) handleIf() []string {
	var out []string

	if isIf(c.raw[c.curr]) {
		out = append(out, "<ifStatement>")
		for {
			if isExprListOpen(c.raw[c.curr]) {
				out = append(out, c.raw[c.curr])
				c.curr++

				out = append(out, "<expression>")
				out = append(out, c.handleExpression()...)
				out = append(out, "</expression>")
			}

			if strings.HasPrefix(c.raw[c.curr], "<symbol>") && strings.Contains(c.raw[c.curr], "{") {
				out = append(out, c.raw[c.curr])
				c.curr++

				out = append(out, "<statements>")
				out = append(out, c.handleStatements()...)
				out = append(out, "</statements>")

				// print out the closing '}'
				out = append(out, c.raw[c.curr])
				c.curr++

				if c.raw[c.curr] != "<keyword> else </keyword>" {
					break
				}
			}

			out = append(out, c.raw[c.curr])
			c.curr++
		}
	}

	if len(out) > 0 {
		out = append(out, "</ifStatement>")
	}

	return out
}

func (c *Compiled) handleDo() []string {
	var out []string
	var done bool

	if isDo(c.raw[c.curr]) {
		out = append(out, "<doStatement>")
		for {
			if isExprListOpen(c.raw[c.curr]) {
				out = append(out, c.raw[c.curr])
				c.curr++

				out = append(out, "<expressionList>")
				if !isExprListClose(c.raw[c.curr]) {
					out = append(out, "<expression>")
					out = append(out, c.handleExpression()...)
					out = append(out, "</expression>")
				}
				out = append(out, "</expressionList>")
			}

			if strings.HasPrefix(c.raw[c.curr], "<symbol>") && strings.Contains(c.raw[c.curr], ";") {
				done = true
			}

			out = append(out, c.raw[c.curr])
			c.curr++

			if done {
				break
			}
		}
	}

	if len(out) > 0 {
		out = append(out, "</doStatement>")
	}

	return out
}

func (c *Compiled) handleWhile() []string {
	var out []string
	var done bool

	if isWhile(c.raw[c.curr]) {
		out = append(out, "<whileStatement>")
		for {
			if isExprListOpen(c.raw[c.curr]) {
				out = append(out, c.raw[c.curr])
				c.curr++

				out = append(out, "<expression>")
				out = append(out, c.handleExpression()...)
				out = append(out, "</expression>")
			}

			if strings.HasPrefix(c.raw[c.curr], "<symbol>") && strings.Contains(c.raw[c.curr], "{") {
				out = append(out, c.raw[c.curr])
				c.curr++

				out = append(out, "<statements>")
				out = append(out, c.handleStatements()...)
				out = append(out, "</statements>")

				out = append(out, c.raw[c.curr])
				c.curr++
				break
			}

			if isSemicolon(c.raw[c.curr]) {
				done = true
			}

			out = append(out, c.raw[c.curr])
			c.curr++

			if done {
				break
			}
		}
	}

	if len(out) > 0 {
		out = append(out, "</whileStatement>")
	}

	return out
}

func (c *Compiled) handleReturn() []string {
	var out []string

	if isReturn(c.raw[c.curr]) {
		out = append(out, "<returnStatement>")
		out = append(out, c.raw[c.curr])
		c.curr++

		for {
			if !isSemicolon(c.raw[c.curr]) {
				out = append(out, "<expression>")
				out = append(out, c.handleExpression()...)
				out = append(out, "</expression>")
				continue
			}

			out = append(out, c.raw[c.curr])
			c.curr++
			break
		}
	}

	if len(out) > 0 {
		out = append(out, "</returnStatement>")
	}

	return out
}

// handleStatements is a recursive call...
func (c *Compiled) handleStatements() []string {
	var out []string

	s := c.raw[c.curr]
	if !strings.HasPrefix(s, "<keyword>") {
		return out
	}

	switch {
	case strings.Contains(s, "let"):
		out = append(out, c.handleLet()...)
	case strings.Contains(s, "if"):
		out = append(out, c.handleIf()...)
	case strings.Contains(s, "while"):
		out = append(out, c.handleWhile()...)
	case strings.Contains(s, "do"):
		out = append(out, c.handleDo()...)
	case strings.Contains(s, "return"):
		out = append(out, c.handleReturn()...)
	}

	return append(out, c.handleStatements()...)
}

// handleVarDec is a recursive call...
func (c *Compiled) handleVarDec() []string {
	var out []string
	var done bool

	if strings.HasPrefix(c.raw[c.curr], "<keyword>") && strings.Contains(c.raw[c.curr], "var") {
		out = append(out, "<varDec>")
		for {
			if strings.HasPrefix(c.raw[c.curr], "<symbol>") && strings.Contains(c.raw[c.curr], ";") {
				done = true
			}

			out = append(out, c.raw[c.curr])
			c.curr++

			if done {
				break
			}
		}
	}

	if len(out) > 0 {
		out = append(out, "</varDec>")
		return append(out, c.handleVarDec()...)
	}

	return out
}

// handleSubroutineBody is a singular call but with multiple parts (varDec, statements)...
func (c *Compiled) handleSubroutineBody() []string {
	var out []string
	var done bool

	if strings.HasPrefix(c.raw[c.curr], "<symbol>") && strings.Contains(c.raw[c.curr], "{") {
		out = append(out, "<subroutineBody>")
		for {
			// handle variable declarations ...
			if strings.HasPrefix(c.raw[c.curr], "<keyword>") && strings.Contains(c.raw[c.curr], "var") {
				out = append(out, c.handleVarDec()...)
			}

			if strings.HasPrefix(c.raw[c.curr], "<keyword>") {
				out = append(out, "<statements>")
				out = append(out, c.handleStatements()...)
			}

			if strings.HasPrefix(c.raw[c.curr], "<symbol>") && strings.Contains(c.raw[c.curr], "}") {
				out = append(out, "</statements>")
				done = true
			}

			out = append(out, c.raw[c.curr])
			c.curr++

			if done {
				break
			}
		}
	}

	if len(out) > 0 {
		out = append(out, "</subroutineBody>")
	}

	return out
}

// handleParameterList is a singular call but with multiple parts (varDec, statements)...
func (c *Compiled) handleParameterList() []string {
	var out = []string{"<parameterList>"}

	for {
		if strings.HasPrefix(c.raw[c.curr], "<symbol>") && strings.Contains(c.raw[c.curr], ")") {
			break
		}

		out = append(out, c.raw[c.curr])
		c.curr++
	}

	out = append(out, "</parameterList>")

	return out
}

// handleSubroutine is a recursive call...
func (c *Compiled) handleSubroutine() []string {
	var out []string
	var done bool

	if strings.HasPrefix(c.raw[c.curr], "<keyword>") && (strings.Contains(c.raw[c.curr], "constructor") || strings.Contains(c.raw[c.curr], "function") || strings.Contains(c.raw[c.curr], "method")) {
		out = append(out, "<subroutineDec>")
		for {
			// handle params first
			if strings.HasPrefix(c.raw[c.curr], "<symbol>") && strings.Contains(c.raw[c.curr], "(") {
				out = append(out, c.raw[c.curr])
				c.curr++

				out = append(out, c.handleParameterList()...)
			}

			// handle the subroutine body...
			if strings.HasPrefix(c.raw[c.curr], "<symbol>") && strings.Contains(c.raw[c.curr], "{") {
				out = append(out, c.handleSubroutineBody()...)
				break
			}

			// finish up...
			if strings.HasPrefix(c.raw[c.curr], "<symbol>") && strings.Contains(c.raw[c.curr], "}") {
				done = true
			}

			out = append(out, c.raw[c.curr])
			c.curr++

			if done {
				break
			}
		}
	}

	if len(out) > 0 {
		out = append(out, "</subroutineDec>")
		return append(out, c.handleSubroutine()...)
	}

	return out
}

func (c *Compiled) handleClassVar() []string {
	var out []string
	var done bool

	if strings.HasPrefix(c.raw[c.curr], "<keyword>") && (strings.Contains(c.raw[c.curr], "field") || strings.Contains(c.raw[c.curr], "static")) {
		out = append(out, "<classVarDec>")
		for {
			if strings.HasPrefix(c.raw[c.curr], "<symbol>") && strings.Contains(c.raw[c.curr], ";") {
				done = true
			}

			out = append(out, c.raw[c.curr])
			c.curr++

			if done {
				break
			}
		}
	}

	if len(out) > 0 {
		out = append(out, "</classVarDec>")
		return append(out, c.handleClassVar()...)
	}

	return out
}

// handleClass defines the class elements
func (c *Compiled) handleClass() []string {
	var out []string
	var done bool

	for {
		if strings.HasPrefix(c.raw[c.curr], "<symbol>") && strings.Contains(c.raw[c.curr], "{") {
			done = true
		}

		// add the class name
		if strings.HasPrefix(c.raw[c.curr], "<keyword>") && strings.Contains(c.raw[c.curr], "class") {
			s := strings.Split(c.raw[c.curr+1], " ")
			c.class.name = s[1]
		}

		out = append(out, c.raw[c.curr])
		c.curr++

		if done {
			break
		}
	}

	return out
}

package parser

import (
	"fmt"
	"strings"
)

var whileCount int
var ifCount int

func (c *Compiled) handleVMTerm(id int, subName string) (int, error) {
	var op, elem string
	x := id
	var f []string // possible function call
	var numArgs int

	var isArray string

	for {
		if isTermEnd(c.d[x]) {
			if len(f) > 0 {
				if sym, ok := c.fetchToken(f[0], subName); ok {
					if !isBuiltin(sym.elem) {
						f[0] = sym.elem
						numArgs++

						c.vm = append(c.vm, fmt.Sprintf("push %s %d", sym.scope, sym.index))
					}
				}

				c.vm = append(c.vm, fmt.Sprintf("call %s %d // first one", strings.Join(f, ""), numArgs))
			}

			if elem != "" && isArray == "" {
				c.vm = append(c.vm, elem)
			} else if isArray != "" {
				c.vm = append(c.vm, isArray)
			}

			id = x
			break
		}

		if isIntConst(c.d[x]) {
			f := strings.Split(c.d[x], " ")
			c.vm = append(c.vm, fmt.Sprintf("push constant %s", f[1]))
		}

		if isKeywordConst(c.d[x]) {
			f := strings.Split(c.d[x], " ")

			switch f[1] {
			case "true":
				c.vm = append(c.vm, "push constant 0")
				c.vm = append(c.vm, "not")
			case "false":
				c.vm = append(c.vm, "push constant 0")
			case "null":
			case "this":
				c.vm = append(c.vm, "push pointer 0")
			}

			x++ // skip past the const
			continue
		}

		if isStrConst(c.d[x]) {
			s := strings.Trim(c.d[x], "</stringConstant>")
			s = s[1 : len(s)-1] // safely remove the artificial space we added

			c.vm = append(c.vm, fmt.Sprintf("push constant %d", len(s)))
			c.vm = append(c.vm, "call String.new 1")
			for _, v := range s {
				c.vm = append(c.vm, fmt.Sprintf("push constant %d", v))
				c.vm = append(c.vm, "call String.appendChar 2")
			}
		}

		if isUnaryOp(c.d[x-1]) && isExpression(c.d[x-3]) {
			op = "neg"
			if c.d[x-1] == "<symbol> ~ </symbol>" {
				op = "not"
			}

			x++ // get past the '<term>"

			if isExprListOpen(c.d[x]) {
				y, err := c.handleVMExpression(x, subName)
				if err != nil {
					return id, err
				}

				c.vm = append(c.vm, op)
				x = y
				x++ // increment past the ')'
				continue

			}

			y, err := c.handleVMTerm(x, subName)
			if err != nil {
				return id, err
			}
			c.vm = append(c.vm, op)

			x = y
		}

		if isExprListOpen(c.d[x]) {
			y, err := c.handleVMExpression(x, subName)
			if err != nil {
				return id, err
			}

			x = y
		}

		if isIdentifier(c.d[x]) && !isPeriod(c.d[x+1]) {
			f := strings.Split(c.d[x], " ")
			if sym, ok := c.fetchToken(f[1], subName); ok {
				elem = fmt.Sprintf("push %s %d", sym.scope, sym.index)
			}
		}

		if isExprOpen(c.d[x]) {
			// grab the prev identifier
			var prev string
			p := strings.Split(c.d[x-1], " ")
			sym, ok := c.fetchToken(p[1], subName)
			if ok {
				prev = fmt.Sprintf("push %s %d", sym.scope, sym.index)
				isArray = fmt.Sprintf("push that %d // isArray", sym.index)
			}

			x++ // skip past the '['

			y, err := c.handleVMExpression(x, subName)
			if err != nil {
				return id, err
			}

			x = y
			c.vm = append(c.vm, prev)
			c.vm = append(c.vm, "add")
			c.vm = append(c.vm, "pop pointer 1")
		}

		if isIdentifier(c.d[x]) && isPeriod(c.d[x+1]) {
			var expr bool

			for {
				if isExprListOpen(c.d[x]) && !isExprListClose(c.d[x+3]) {
					x++ // skip past '('
					x++ // skip past '<expressionList>'
					numArgs++

					y, err := c.handleVMExpression(x, subName)
					if err != nil {
						return id, err
					}

					x = y
					expr = true
				}

				if isExprListOpen(c.d[x]) && isExprListClose(c.d[x+3]) {
					// empty <expressionList>
					x++
					x++
					x++
					break
				}

				if isComma(c.d[x]) {
					x++ // skip past ','
					numArgs++

					y, err := c.handleVMExpression(x, subName)
					if err != nil {
						return id, err
					}

					x = y
				}

				if isExprListClose(c.d[x]) && !isSemicolon(c.d[x+1]) {
					expr = false
					break
				}

				// build up the function name to call up to first '('
				if !isExprListOpen(c.d[x]) && !expr {
					elem := strings.Split(c.d[x], " ")
					f = append(f, elem[1])
				}
				x++
			}
		}

		x++
	}

	return id, nil
}

func (c *Compiled) handleVMExpression(id int, subName string) (int, error) {
	x := id

	var mathFunc string

	for {
		if isExpressionEnd(c.d[x]) {
			switch mathFunc {
			case "+":
				c.vm = append(c.vm, fmt.Sprintf("add"))
			case "-":
				c.vm = append(c.vm, fmt.Sprintf("sub"))
			case "*":
				c.vm = append(c.vm, fmt.Sprintf("call Math.multiply 2"))
			case "/":
				c.vm = append(c.vm, fmt.Sprintf("call Math.divide 2"))
			case "&gt;":
				c.vm = append(c.vm, "gt")
			case "&lt;":
				c.vm = append(c.vm, "lt")
			case "&amp;":
				c.vm = append(c.vm, "and")
			case "=":
				c.vm = append(c.vm, "eq")
			case "|":
				c.vm = append(c.vm, "or")
			default:
			}

			id = x
			break
		}

		if isTerm(c.d[x]) {
			y, err := c.handleVMTerm(x, subName)
			if err != nil {
				return id, err
			}

			x = y
		}

		if isOp(c.d[x]) {
			f := strings.Split(c.d[x], " ")
			mathFunc = f[1]

			x++ // skip past symbol and handle next term
			y, err := c.handleVMTerm(x, subName)
			if err != nil {
				return id, err
			}

			x = y
		}

		x++
	}

	return id, nil
}

func (c *Compiled) fetchToken(f, subName string) (*symbol, bool) {
	for _, x := range c.subroutines[subName].symbols {
		if x.name == f {
			return x, true
		}
	}

	for _, x := range c.class.symbols {
		if x.name == f {
			return x, true
		}
	}

	return nil, false
}

func (c *Compiled) handleVMStatements(id int, subName string) (int, error) {
	x := id

	for {
		if isStatementsEnd(c.d[x]) {
			id = x
			break
		}

		if isDoFunc(c.d[x]) {
			y, err := c.handleVMDo(x, subName)
			if err != nil {
				return id, err
			}

			x = y
		}

		if isLetFunc(c.d[x]) {
			y, err := c.handleVMLet(x, subName)
			if err != nil {
				return id, err
			}

			x = y
		}

		if isWhileFunc(c.d[x]) {
			whileCount++
			y, err := c.handleVMWhile(x, whileCount-1, subName)
			if err != nil {
				return id, err
			}

			x = y
		}

		if isIfFunc(c.d[x]) {
			ifCount++
			y, err := c.handleVMIf(x, ifCount-1, subName)
			if err != nil {
				return id, err
			}

			x = y
		}

		if isReturnFunc(c.d[x]) {
			y, err := c.handleVMReturn(x, subName)
			if err != nil {
				return id, err
			}

			x = y
		}

		x++
	}

	return id, nil
}

func (c *Compiled) handleVMIf(id, ifCnt int, subName string) (int, error) {
	x := id + 1
	x++ // skip past the `if`

	for {
		if isIfFuncEnd(c.d[x]) {
			id = x
			break
		}

		if isExprListOpen(c.d[x]) && !isExprListClose(c.d[x+1]) {
			x++ // skip past '('
			x++ // skip past '<expression>'

			y, err := c.handleVMExpression(x, subName)
			if err != nil {
				return id, err
			}

			x = y

			c.vm = append(c.vm, fmt.Sprintf("if-goto IF_TRUE%d", ifCnt))
			c.vm = append(c.vm, fmt.Sprintf("goto IF_FALSE%d", ifCnt))
			c.vm = append(c.vm, fmt.Sprintf("label IF_TRUE%d", ifCnt))

			y, err = c.handleVMStatements(x, subName)
			if err != nil {
				return id, err
			}

			x = y
			x++ // skip past the '</statements>'

			if !strings.Contains(c.d[x+1], "else") {
				c.vm = append(c.vm, fmt.Sprintf("label IF_FALSE%d", ifCnt))
				continue
			}
			c.vm = append(c.vm, fmt.Sprintf("goto IF_END%d", ifCnt))
			x++ // skip past the 'else'
			x++ // skip past the '<statements>'

			// think this should only happen when there is an 'else' block
			c.vm = append(c.vm, fmt.Sprintf("label IF_FALSE%d", ifCnt))

			y, err = c.handleVMStatements(x, subName)
			if err != nil {
				return id, err
			}

			x = y
			c.vm = append(c.vm, fmt.Sprintf("label IF_END%d", ifCnt))
			continue
		}

		// handle the statements
		if strings.HasPrefix(c.d[x], "<symbol>") && strings.Contains(c.d[x], "{") {
			break
		}

		x++
	}

	return id, nil
}

func (c *Compiled) handleVMWhile(id, whileCnt int, subName string) (int, error) {
	x := id + 1
	x++ // skip past the `while`

	c.vm = append(c.vm, fmt.Sprintf("label WHILE_EXP%d", whileCnt))

	for {
		if isWhileFuncEnd(c.d[x]) {
			id = x
			break
		}

		if isExprListOpen(c.d[x]) && !isExprListClose(c.d[x+1]) {
			x++ // skip past '('
			x++ // skip past '<expression>'

			y, err := c.handleVMExpression(x, subName)
			if err != nil {
				return id, err
			}

			x = y

			c.vm = append(c.vm, "not")
			c.vm = append(c.vm, fmt.Sprintf("if-goto WHILE_END%d", whileCnt))

			y, err = c.handleVMStatements(x, subName)
			if err != nil {
				return id, err
			}

			x = y
			c.vm = append(c.vm, fmt.Sprintf("goto WHILE_EXP%d", whileCnt))
			c.vm = append(c.vm, fmt.Sprintf("label WHILE_END%d", whileCnt))
			continue
		}

		// handle the statements
		if strings.HasPrefix(c.d[x], "<symbol>") && strings.Contains(c.d[x], "{") {
			//			out = append(out, c.raw[c.curr])
			//			c.curr++

			//			out = append(out, c.handleStatements()...)

			//			out = append(out, c.raw[c.curr])
			//			c.curr++
			break
		}

		x++
	}

	return id, nil
}

func (c *Compiled) handleVMLet(id int, subName string) (int, error) {
	var elem string

	var isArray bool

	x := id + 1
	x++ // skip past the `let`

	for {
		if isLetFuncEnd(c.d[x]) {
			if !isArray {
				c.vm = append(c.vm, elem)
			} else {
				c.vm = append(c.vm, "pop temp 0")
				c.vm = append(c.vm, "pop pointer 1")
				c.vm = append(c.vm, "push temp 0")
				c.vm = append(c.vm, "pop that 0")
			}

			id = x
			break
		}

		if isIdentifier(c.d[x]) {
			f := strings.Split(c.d[x], " ")
			if sym, ok := c.fetchToken(f[1], subName); ok {
				elem = fmt.Sprintf("pop %s %d", sym.scope, sym.index)
			}
		}

		if isExprOpen(c.d[x]) {
			// grab the prev identifier
			var prev string
			isArray = true
			p := strings.Split(c.d[x-1], " ")
			sym, ok := c.fetchToken(p[1], subName)
			if ok {
				prev = fmt.Sprintf("push %s %d", sym.scope, sym.index)
				//				c.vm = append(c.vm,
			}

			x++ // skip past the '['

			y, err := c.handleVMExpression(x, subName)
			if err != nil {
				return id, err
			}

			x = y
			c.vm = append(c.vm, prev)
			c.vm = append(c.vm, "add // vmlet")
		}

		if strings.Contains(c.d[x], "=") {
			x++ // skip past the '='

			y, err := c.handleVMExpression(x, subName)
			if err != nil {
				return id, err
			}

			x = y
		}

		x++
	}
	return id, nil
}

func (c *Compiled) handleVMDo(id int, subName string) (int, error) {
	var f []string
	var expr, isClass bool
	var numArgs int

	x := id + 1
	x++ // skip past the `do`

	for {
		if isDoFuncEnd(c.d[x]) {
			id = x
			break
		}

		if isExprListOpen(c.d[x]) && !isExprListClose(c.d[x+3]) {
			x++ // skip past '('
			x++ // skip past '<expressionList>'
			numArgs++

			y, err := c.handleVMExpression(x, subName)
			if err != nil {
				return id, err
			}

			x = y
			expr = true
		}

		if isExprListOpen(c.d[x]) && isExprListClose(c.d[x+3]) {
			// empty <expressionList>
			x++
			x++
			x++
			continue
		}

		if isComma(c.d[x]) {
			x++ // skip past ','
			numArgs++

			y, err := c.handleVMExpression(x, subName)
			if err != nil {
				return id, err
			}

			x = y
		}

		if isExprListClose(c.d[x]) && !isSemicolon(c.d[x+1]) {
			expr = false
			x++ // skip past the ')' that is not terminating line
			continue
		}

		// close the main expression list
		if isExprListClose(c.d[x]) && isSemicolon(c.d[x+1]) {
			if len(f) > 1 {
				if sym, ok := c.fetchToken(f[0], subName); ok {
					if !isBuiltin(sym.elem) {
						f[0] = sym.elem

						if !isClass {
							c.vm = append(c.vm, fmt.Sprintf("push %s %d", sym.scope, sym.index))
							numArgs++
						}
					}
				}
			}

			if len(f) == 1 {
				// assume this is a local func
				if !isClass {
					c.vm = append(c.vm, "push pointer 0")
					numArgs++
				}
				f = []string{
					c.class.name,
					".",
					f[0],
				}
			}
			c.vm = append(c.vm, fmt.Sprintf("call %s %d // funtime", strings.Join(f, ""), numArgs))
			c.vm = append(c.vm, "pop temp 0")
		}

		// build up the function name to call up to first '('
		if !isExprListOpen(c.d[x]) && !expr {
			elem := strings.Split(c.d[x], " ")
			for _, x := range c.class.symbols {
				if x.name == elem[1] {
					isClass = true
					c.vm = append(c.vm, fmt.Sprintf("push %s %d", x.scope, x.index))
					numArgs++
					break
				}
			}

			f = append(f, elem[1])

			if isDo(c.d[x-1]) && isIdentifier(c.d[x]) && isExprListOpen(c.d[x+1]) {
				// maybe landed on local value
				isClass = true
				c.vm = append(c.vm, "push pointer 0")
				numArgs++
			}
		}

		x++
	}

	return id, nil
}

func (c *Compiled) handleVMReturn(id int, subName string) (int, error) {
	retType := c.subroutines[subName].retType

	if retType == "void" {
		c.vm = append(c.vm, []string{
			"push constant 0",
			"return",
		}...)
	}

	if retType != "void" {
		id++ // skip past '<returnStatement>'
		id++ // skip past 'return'
		y, err := c.handleVMExpression(id, subName)
		if err != nil {
			return id, nil
		}

		id = y
		c.vm = append(c.vm, "return")
	}

	return id, nil
}

func (c *Compiled) handleFunction(id int) (int, error) {
	var subName string
	x := id + 1

	for {
		if isSubroutineEnd(c.d[x]) {
			id = x
			break
		}

		if isSubroutineStart(c.d[x]) {
			x++ // move to retType
			x++ // move to identifier

			f := strings.Split(c.d[x], " ")
			if _, ok := c.subroutines[f[1]]; !ok {
				return id, fmt.Errorf("could not find function token: %s", f[1])
			}

			subName = f[1]

			params := c.subroutines[subName].counts[LOCAL]
			if c.subroutines[subName].counts[ARGUMENT] > 0 && c.subroutines[subName].symbols[0].name == "this" {
				params++
			}

			if strings.Contains(c.d[x-2], "method") {
				params--
			}

			// function call
			c.vm = append(c.vm, fmt.Sprintf("function %s.%s %d", c.class.name, subName, params))

			if strings.Contains(c.d[x-2], "constructor") {
				if len(c.class.symbols) > 0 {
					var val int

					for _, x := range c.class.symbols {
						if x.scope != STATIC {
							val++
						}
					}
					c.vm = append(c.vm, fmt.Sprintf("push constant %d", val))
					c.vm = append(c.vm, "call Memory.alloc 1")
					c.vm = append(c.vm, "pop pointer 0")
				}
			} else {
				if c.subroutines[subName].counts[ARGUMENT] > 0 && c.subroutines[subName].symbols[0].name == "this" {
					c.vm = append(c.vm, "push argument 0")
					c.vm = append(c.vm, "pop pointer 0")
				}
			}
		}

		if isDoFunc(c.d[x]) {
			y, err := c.handleVMDo(x, subName)
			if err != nil {
				return id, err
			}

			x = y
		}

		if isLetFunc(c.d[x]) {
			y, err := c.handleVMLet(x, subName)
			if err != nil {
				return id, err
			}

			x = y
		}

		if isWhileFunc(c.d[x]) {
			whileCount++
			y, err := c.handleVMWhile(x, whileCount-1, subName)
			if err != nil {
				return id, err
			}

			x = y
		}

		if isIfFunc(c.d[x]) {
			ifCount++
			y, err := c.handleVMIf(x, ifCount-1, subName)
			if err != nil {
				return id, err
			}

			x = y
		}

		if isReturnFunc(c.d[x]) {
			y, err := c.handleVMReturn(x, subName)
			if err != nil {
				return id, err
			}

			x = y
		}

		x++

	}

	return id, nil
}

func (c *Compiled) buildVM() error {
	var id int

	for {
		if id == len(c.d)-1 {
			break
		}

		if isSubroutine(c.d[id]) {
			x, err := c.handleFunction(id)
			if err != nil {
				return err
			}

			id = x
		}

		id++
	}

	return nil
}

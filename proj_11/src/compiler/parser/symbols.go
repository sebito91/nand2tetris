package parser

import (
	"strings"
)

func (c *Compiled) vardecTable(id int, subName string) (int, error) {
	var symbols []*symbol
	x := id

	var retType string

	isVar := false // random boolean flag to see if we're inside a var definition

	for {
		if isVarDecEnd(c.d[x]) {
			id = x
			break
		}

		if isVarDec(c.d[x]) {
			x++ // increment to actual vars
			x++
			isVar = true

			k := strings.Split(c.d[x], " ")

			retType = k[1]
			x++ // increment to the identifier
		}

		if isIdentifier(c.d[x]) && isVar {
			i := strings.Split(c.d[x], " ")

			c.subroutines[subName].symbols = append(c.subroutines[subName].symbols, &symbol{
				name:  i[1],
				elem:  retType,
				scope: LOCAL,
				index: c.subroutines[subName].counts[LOCAL],
			})

			c.subroutines[subName].counts[LOCAL]++
		}

		if isSemicolon(c.d[x]) {
			isVar = false
		}

		x++
	}

	c.subroutines[subName].symbols = append(c.subroutines[subName].symbols, symbols...)

	return id, nil
}

func (c *Compiled) subroutineTable(id int) (int, error) {
	var subName, retType string
	var symbols []*symbol
	x := id + 1

	isParam := false // random boolean flag to see if we're inside a parameterList

	for {
		if isSubroutineEnd(c.d[x]) {
			id = x
			break
		}

		if isSubroutineStart(c.d[x]) {
			x++ // increment to actual return type
			t := strings.Split(c.d[x], " ")

			x++ // increment to the name of the sub
			f := strings.Split(c.d[x], " ")
			retType = t[1]
			subName = f[1]

			c.subroutines[subName] = &symbolTable{
				retType: retType,
				name:    subName,
				counts: map[string]int{
					STATIC:   0,
					THIS:     0,
					ARGUMENT: 0,
					LOCAL:    0,
				},
			}

			if strings.Contains(c.d[x-2], "method") {
				c.subroutines[subName].symbols = []*symbol{
					&symbol{
						name:  "this",
						elem:  c.class.name,
						scope: ARGUMENT,
						index: c.subroutines[subName].counts[ARGUMENT],
					},
				}
				c.subroutines[subName].counts[ARGUMENT]++
			}
		}

		if isKeyword(c.d[x]) && isParam {
			k := strings.Split(c.d[x], " ")
			i := strings.Split(c.d[x+1], " ")

			c.subroutines[subName].symbols = append(c.subroutines[subName].symbols, &symbol{
				name:  i[1],
				elem:  k[1],
				scope: ARGUMENT,
				index: c.subroutines[subName].counts[ARGUMENT],
			})

			c.subroutines[subName].counts[ARGUMENT]++
			x++ // increment past the identifier
		}

		if isParamList(c.d[x]) {
			isParam = true
		}

		if isParamListEnd(c.d[x]) {
			isParam = false
		}

		if isVarDec(c.d[x]) {
			y, err := c.vardecTable(x, subName)
			if err != nil {
				return x, err
			}

			x = y
		}

		x++
	}

	c.subroutines[subName].symbols = append(c.subroutines[subName].symbols, symbols...)

	return id, nil
}

func (c *Compiled) classTable(id int) (int, error) {
	var symbols []*symbol
	x := id + 1

	sym := &symbol{}
	for {
		if isClassVarDecEnd(c.d[x]) {
			id = x
			symbols = append(symbols, sym)
			break
		}

		s := strings.Split(c.d[x], " ")
		if isKeyword(c.d[x]) {
			switch s[1] {
			case "field":
				sym.scope = THIS
				sym.index = c.class.counts[THIS]
				c.class.counts[THIS]++
			case "static":
				sym.scope = STATIC
				sym.index = c.class.counts[STATIC]
				c.class.counts[STATIC]++
			case "boolean", "int", "char":
				sym.elem = s[1]
			}
			x++ // increment to the next identifier/keyword

			t := strings.Split(c.d[x], " ")
			sym.elem = t[1]

			x++ // increment to the next handler
			continue
		}

		if isIdentifier(c.d[x]) {
			sym.name = s[1]
		}

		if isComma(c.d[x]) {
			symbols = append(symbols, sym)
			d := len(symbols)
			sym = &symbol{
				scope: symbols[d-1].scope,
				elem:  symbols[d-1].elem,
				index: c.class.counts[symbols[d-1].scope],
			}
			c.class.counts[symbols[d-1].scope]++
		}

		x++
	}

	c.class.symbols = append(c.class.symbols, symbols...)

	return id, nil
}

func (c *Compiled) buildSymbolTables() error {
	var id int

	for {
		if id == len(c.d)-1 {
			break
		}

		if isClassVarDec(c.d[id]) {
			x, err := c.classTable(id)
			if err != nil {
				return err
			}
			id = x
		}

		if isSubroutine(c.d[id]) {
			x, err := c.subroutineTable(id)
			if err != nil {
				return err
			}

			id = x
		}
		id++
	}

	return nil
}

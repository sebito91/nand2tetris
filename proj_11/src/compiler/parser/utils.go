package parser

import "strings"

const (
	// STATIC is the `static` class-level var
	STATIC = "static"

	// THIS is the 'field' class-level var
	THIS = "this"

	// ARGUMENT is the `argument` subroutine-level parameter
	ARGUMENT = "argument"

	// LOCAL is the `var` subroutine-level local variable
	LOCAL = "local"
)

// symbolTable is the default struct to handle our symbol table references
// throughout the code. The class-level ST will be housed within the `Compiled`
// struct, but subroutine-level STs will be created/destroyed with respective
// scope
type symbolTable struct {
	name    string         // class name
	retType string         // return type
	counts  map[string]int // index for each var type
	symbols []*symbol
}

// symbol is the unique element that makes up a symbol tables
type symbol struct {
	name  string
	elem  string
	scope string
	index int
}

func isBuiltin(x string) bool {
	return x == "boolean" || x == "char" || x == "int"
}

func isStatementsEnd(x string) bool {
	return x == "</statements>"
}

func isReturnFunc(x string) bool {
	return x == "<returnStatement>"
}

func isExpression(x string) bool {
	return x == "<expression>"
}

func isExpressionEnd(x string) bool {
	return x == "</expression>"
}

func isTerm(x string) bool {
	return x == "<term>"
}

func isTermEnd(x string) bool {
	return x == "</term>"
}

func isVarDec(x string) bool {
	return x == "<varDec>"
}

func isVarDecEnd(x string) bool {
	return x == "</varDec>"
}

func isParamList(x string) bool {
	return x == "<parameterList>"
}

func isParamListEnd(x string) bool {
	return x == "</parameterList>"
}

func isSubroutine(x string) bool {
	return x == "<subroutineDec>"
}

func isSubroutineStart(x string) bool {
	return strings.Contains(x, "constructor") || strings.Contains(x, "function") || strings.Contains(x, "method")
}

func isSubroutineEnd(x string) bool {
	return x == "</subroutineDec>"
}

func isClassVarDec(x string) bool {
	return x == "<classVarDec>"
}

func isClassVarDecEnd(x string) bool {
	return x == "</classVarDec>"
}

func isSemicolon(x string) bool {
	return strings.HasPrefix(x, "<symbol>") && strings.Contains(x, ";")
}

func isKeyword(x string) bool {
	return strings.HasPrefix(x, "<keyword>")
}

func isReturn(x string) bool {
	return x == "<keyword> return </keyword>"
}

func isLet(x string) bool {
	return x == "<keyword> let </keyword>"
}

func isIf(x string) bool {
	return x == "<keyword> if </keyword>"
}

func isWhile(x string) bool {
	return x == "<keyword> while </keyword>"
}

func isDo(x string) bool {
	return x == "<keyword> do </keyword>"
}

func isDoFunc(x string) bool {
	return x == "<doStatement>"
}

func isDoFuncEnd(x string) bool {
	return x == "</doStatement>"
}

func isLetFunc(x string) bool {
	return x == "<letStatement>"
}

func isLetFuncEnd(x string) bool {
	return x == "</letStatement>"
}

func isWhileFunc(x string) bool {
	return x == "<whileStatement>"
}

func isWhileFuncEnd(x string) bool {
	return x == "</whileStatement>"
}

func isIfFunc(x string) bool {
	return x == "<ifStatement>"
}

func isIfFuncEnd(x string) bool {
	return x == "</ifStatement>"
}

func isOp(x string) bool {
	return strings.Contains(x, "<symbol>") && (strings.Contains(x, "=") || strings.Contains(x, "+") || strings.Contains(x, "-") || strings.Contains(x, "*") || strings.Contains(x, "&") || strings.Contains(x, "|") || strings.Contains(x, "&lt;") || strings.Contains(x, "&gt;") || strings.Contains(x, "&amp;") || x == "<symbol> / </symbol>")
}

func isUnaryOp(x string) bool {
	return strings.Contains(x, "<symbol>") && (strings.Contains(x, "-") || strings.Contains(x, "~"))
}

func isIntConst(x string) bool {
	return strings.Contains(x, "<integerConstant>")
}

func isStrConst(x string) bool {
	return strings.Contains(x, "<stringConstant>")
}

func isIdentifier(x string) bool {
	return strings.Contains(x, "<identifier>")
}

func isExprListOpen(x string) bool {
	return x == "<symbol> ( </symbol>"
}

func isExprOpen(x string) bool {
	return x == "<symbol> [ </symbol>"
}

func isExprListClose(x string) bool {
	return x == "<symbol> ) </symbol>"
}

func isExprClose(x string) bool {
	return x == "<symbol> ] </symbol>"
}

func isPeriod(x string) bool {
	return x == "<symbol> . </symbol>"
}

func isComma(x string) bool {
	return x == "<symbol> , </symbol>"
}

func isKeywordConst(x string) bool {
	return strings.Contains(x, "<keyword>") && (strings.Contains(x, "true") || strings.Contains(x, "false") || strings.Contains(x, "null") || strings.Contains(x, "this"))
}

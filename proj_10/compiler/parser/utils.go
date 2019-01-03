package parser

import "strings"

func isSemicolon(x string) bool {
	return strings.HasPrefix(x, "<symbol>") && strings.Contains(x, ";")
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

package cache

// This file deals with lexical matters of HTTP

func isWhiteSpace(c byte) bool {
	return c == '\t' || c == ' '
}

func isSeparator(c byte) bool {
	switch c {
	case '(', ')', '<', '>', '@', ',', ';', ':', '\\', '"', '/', '[', ']', '?', '=', '{', '}', ' ', '\t':
		return true
	}
	return false
}

func isCtl(c byte) bool { return (c <= 31) || c == 127 }

func isChar(c byte) bool { return c <= 127 } //nolint

func isAnyText(c byte) bool { return !isCtl(c) }

func isQdText(c byte) bool { return isAnyText(c) && c != '"' }

func isToken(c byte) bool { return isChar(c) && !isCtl(c) && !isSeparator(c) }

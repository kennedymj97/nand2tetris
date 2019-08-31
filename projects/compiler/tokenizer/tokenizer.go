package tokenizer

import (
	"bytes"
	"strconv"
	"unicode"
	"unicode/utf8"
)

func removeComment(data []byte) (bool, int) {
	var width int
	if string(data[:2]) == "//" {
		width = bytes.IndexByte(data, '\n') + 1
		return true, width
	} else if string(data[:3]) == "/**" {
		width = bytes.Index(data, []byte("*/")) + 2
		return true, width
	}
	return false, -1
}

func isStringConstant(r rune) bool {
	return r == '"'
}

func isSymbol(r rune) bool {
	switch r {
	case '{', '}', '(', ')', '[', ']', '.', ',', ';', '+', '-', '*', '/', '&', '|', '<', '>', '=', '~':
		return true
	}
	return false
}

// isSpace reports whether the character is a Unicode white space character.
// We avoid dependency on the unicode package, but check validity of the implementation
// in the tests.
func isSpace(r rune) bool {
	if r <= '\u00FF' {
		// Obvious ASCII ones: \t through \r plus space. Plus two Latin-1 oddballs.
		switch r {
		case ' ', '\t', '\n', '\v', '\f', '\r':
			return true
		case '\u0085', '\u00A0':
			return true
		}
		return false
	}
	// High-valued ones.
	if '\u2000' <= r && r <= '\u200a' {
		return true
	}
	switch r {
	case '\u1680', '\u2028', '\u2029', '\u202f', '\u205f', '\u3000':
		return true
	}
	return false
}

// ScanTokens is a split function for a Scanner that returns each token in a jack file.
func ScanTokens(data []byte, atEOF bool) (advance int, token []byte, err error) {
	var r rune
	var width int
	start := 0
	// Skip leading spaces and comments.
	for width = 0; start < len(data); start += width {
		r, width = utf8.DecodeRune(data[start:])
		// If the first 2 runes are // then find where the end of the line is and shift the start value
		if isComment, commentWidth := removeComment(data[start:]); isComment {
			width = commentWidth
		} else if !isSpace(r) {
			break
		}
	}

	// If the first rune is a symbol or string constant return it
	r, width = utf8.DecodeRune(data[start:])
	if isSymbol(r) {
		return start + width, data[start : start+width], nil
	} else if isStringConstant(r) {
		constantEndIndex := bytes.IndexByte(data[start+1:], '"') + 2
		return start + constantEndIndex, data[start : start+constantEndIndex], nil
	}

	// Scan until space or comment, marking end of word.
	for width, i := 0, start; i < len(data); i += width {
		r, width = utf8.DecodeRune(data[i:])
		if isSpace(r) || isSymbol(r) {
			return i, data[start:i], nil
		}
	}

	// If we're at EOF, we have a final, non-empty, non-terminated word. Return it.
	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}

	// Request more data.
	return start, nil, nil
}

func isKeyword(token string) bool {
	keywords := []string{
		"class",
		"method",
		"function",
		"constructor",
		"int",
		"boolean",
		"char",
		"void",
		"var",
		"static",
		"field",
		"let",
		"do",
		"if",
		"else",
		"while",
		"return",
		"true",
		"false",
		"null",
		"this",
	}

	for _, keyword := range keywords {
		if token == keyword {
			return true
		}
	}

	return false
}

func isInt(token string) bool {
	if _, err := strconv.Atoi(token); err == nil {
		return true
	}
	return false
}

func isIdentifier(token string) bool {
	if isInt(string(token[0])) {
		return false
	}
	for _, char := range token {
		if !unicode.IsDigit(char) && !unicode.IsLetter(char) && char != '_' {
			return false
		}
	}
	return true
}

func tokenType(token string) string {
	if isKeyword(token) {
		return "KEYWORD"
	} else if isSymbol(rune(token[0])) {
		return "SYMBOL"
	} else if isStringConstant(rune(token[0])) {
		return "STRING_CONST"
	} else if isInt(token) {
		return "INT_CONST"
	} else if isIdentifier(token) {
		return "IDENTIFIER"
	}
	return ""
}

func keyword(token string) (updatedToken string) {
	return token
}

func symbol(token string) (updatedToken string) {
	switch token {
	case "<":
		token = "&lt;"
	case ">":
		token = "&gt;"
	case "&":
		token = "&amp;"
	case `"`:
		token = "&quot;"
	}
	return token
}

func stringConst(token string) (updatedToken string) {
	token = token[1 : len(token)-1]
	return token
}

func integerConst(token string) (updatedToken string) {
	return token
}

func identifier(token string) (updatedToken string) {
	return token
}

func ParseToken(token string) (string, string) {
	tokenType := tokenType(token)
	switch tokenType {
	case "KEYWORD":
		token := keyword(token)
		return token, tokenType
	case "SYMBOL":
		token := symbol(token)
		return token, tokenType
	case "STRING_CONST":
		token := stringConst(token)
		return token, tokenType
	case "INTEGER_CONST":
		token := integerConst(token)
		return token, tokenType
	case "IDENTIFIER":
		token := identifier(token)
		return token, tokenType
	default:
		return "", ""
	}
}

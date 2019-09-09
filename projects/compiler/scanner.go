package compiler

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"strconv"
	"unicode"
	"unicode/utf8"
)

type category uint8

// Defined the possible token categories
const (
	unknown category = iota
	keyword
	symbol
	stringConst
	intConst
	identifier
)

// Below are functions used to check if a token belongs to a category according to the Jack grammar
// specification.
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

func isSymbol(token string) bool {
	switch token {
	case "{", "}", "(", ")", "[", "]", ".", ",", ";", "+", "-", "*", "/", "&", "|", "<", ">", "=", "~":
		return true
	}
	return false
}

func isStringConstant(token string) bool {
	return string(token[0]) == `"`
}

func isIntConstant(token string) bool {
	if _, err := strconv.Atoi(token); err == nil {
		return true
	}
	return false
}

func isIdentifier(token string) bool {
	if isIntConstant(string(token[0])) {
		return false
	}
	for _, char := range token {
		if !unicode.IsDigit(char) && !unicode.IsLetter(char) && char != '_' {
			return false
		}
	}
	return true
}

func tokenCategory(token string) category {
	if isKeyword(token) {
		return keyword
	} else if isSymbol(token) {
		return symbol
	} else if isStringConstant(token) {
		return stringConst
	} else if isIntConstant(token) {
		return intConst
	} else if isIdentifier(token) {
		return identifier
	}
	return unknown
}

func formatSymbol(token string) (updatedToken string) {
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

func formatStringConst(token string) (updatedToken string) {
	token = token[1 : len(token)-1]
	return token
}

type token struct {
	value    string
	category category
}

func newToken(tokenValue string) (*token, error) {
	category := tokenCategory(tokenValue)
	switch category {
	case keyword:
		return &token{tokenValue, category}, nil
	case symbol:
		tokenValue := formatSymbol(tokenValue)
		return &token{tokenValue, category}, nil
	case stringConst:
		tokenValue := formatStringConst(tokenValue)
		return &token{tokenValue, category}, nil
	case intConst:
		return &token{tokenValue, category}, nil
	case identifier:
		return &token{tokenValue, category}, nil
	}
	// return an error here if the token is unknown
	return &token{tokenValue, category}, fmt.Errorf("invalid token: %s", tokenValue)
}

func (t *token) isType() bool {
	if t.category == identifier {
		return true
	}

	switch t.value {
	case "int", "char", "boolean":
		return true
	default:
		return false
	}
}

func (t *token) isOp() bool {
	switch t.value {
	case "+", "-", "*", "/", "&amp;", "|", "&lt;", "&gt;", "=":
		return true
	default:
		return false
	}
}

func (t *token) isUnaryOp() bool {
	switch t.value {
	case "-", "~":
		return true
	default:
		return false
	}
}

func (t *token) isKeywordConstant() bool {
	switch t.value {
	case "true", "false", "null", "this":
		return true
	default:
		return false
	}
}

type scanner struct {
	*bufio.Scanner
	token *token
}

func isComment(data []byte) (bool, int) {
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
func scanTokens(data []byte, atEOF bool) (advance int, token []byte, err error) {
	var r rune
	var width int
	start := 0
	// Skip leading spaces and comments.
	for width = 0; start < len(data); start += width {
		r, width = utf8.DecodeRune(data[start:])
		if comment, commentWidth := isComment(data[start:]); comment {
			width = commentWidth
		} else if !isSpace(r) {
			break
		}
	}

	// If the first rune is a symbol or string constant return it
	r, width = utf8.DecodeRune(data[start:])
	if isSymbol(string(r)) {
		return start + width, data[start : start+width], nil
	} else if isStringConstant(string(r)) {
		constantEndIndex := bytes.IndexByte(data[start+1:], '"') + 2
		return start + constantEndIndex, data[start : start+constantEndIndex], nil
	}

	// Scan until space or comment, marking end of word.
	for width, i := 0, start; i < len(data); i += width {
		r, width = utf8.DecodeRune(data[i:])
		if isSpace(r) || isSymbol(string(r)) {
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

func newScanner(file io.Reader) *scanner {
	bufioScanner := bufio.NewScanner(file)
	bufioScanner.Split(scanTokens)
	return &scanner{bufioScanner, &token{}}
}

func (s *scanner) advance() {
	s.Scan()
	tokenValue := s.Text()
	token, err := newToken(tokenValue)
	if err != nil {
		log.Fatal(err)
	}
	s.token = token
}

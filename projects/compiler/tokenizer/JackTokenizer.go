package tokenizer

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

type Category uint8

// Defined the possible token categories
const (
	Unknown Category = iota
	Keyword
	Symbol
	StringConst
	IntConst
	Identifier
)

func (c Category) String() string {
	switch c {
	case Keyword:
		return "keyword"
	case Symbol:
		return "symbol"
	case StringConst:
		return "stringConstant"
	case IntConst:
		return "integerConstant"
	case Identifier:
		return "identifier"
	default:
		return ""
	}
}

// Below are functions used to check if a token belongs to a Category according to the Jack grammar
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

func tokenCategory(token string) Category {
	if isKeyword(token) {
		return Keyword
	} else if isSymbol(token) {
		return Symbol
	} else if isStringConstant(token) {
		return StringConst
	} else if isIntConstant(token) {
		return IntConst
	} else if isIdentifier(token) {
		return Identifier
	}
	return Unknown
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
	Value    string
	Category Category
}

func newToken(tokenValue string) (*token, error) {
	Category := tokenCategory(tokenValue)
	switch Category {
	case Keyword:
		return &token{tokenValue, Category}, nil
	case Symbol:
		tokenValue := formatSymbol(tokenValue)
		return &token{tokenValue, Category}, nil
	case StringConst:
		tokenValue := formatStringConst(tokenValue)
		return &token{tokenValue, Category}, nil
	case IntConst:
		return &token{tokenValue, Category}, nil
	case Identifier:
		return &token{tokenValue, Category}, nil
	}
	// return an error here if the token is unknown
	return &token{tokenValue, Category}, fmt.Errorf("invalid token: %s", tokenValue)
}

func (t *token) IsType() bool {
	if t.Category == Identifier {
		return true
	}

	switch t.Value {
	case "int", "char", "boolean":
		return true
	default:
		return false
	}
}

func (t *token) IsOp() bool {
	switch t.Value {
	case "+", "-", "*", "/", "&amp;", "|", "&lt;", "&gt;", "=":
		return true
	default:
		return false
	}
}

func (t *token) IsUnaryOp() bool {
	switch t.Value {
	case "-", "~":
		return true
	default:
		return false
	}
}

func (t *token) isKeywordConstant() bool {
	switch t.Value {
	case "true", "false", "null", "this":
		return true
	default:
		return false
	}
}

type Scanner struct {
	*bufio.Scanner
	Token *token
}

func NewScanner(file io.Reader) *Scanner {
	bufioScanner := bufio.NewScanner(file)
	bufioScanner.Split(scanTokens)
	const maxCap = 4096 * 2
	buf := make([]byte, maxCap)
	bufioScanner.Buffer(buf, maxCap)
	return &Scanner{bufioScanner, &token{}}
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

func isComment(data []byte) (bool, int) {
	var width int
	if string(data[:2]) == "//" {
		width = bytes.IndexByte(data, '\n') + 1
		return true, width
	} else if len(data) > 2 && string(data[:3]) == "/**" {
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

func (s *Scanner) Advance() {
	s.Scan()
	tokenValue := s.Text()
	token, err := newToken(tokenValue)
	if err != nil {
		log.Fatal(err)
	}
	s.Token = token
}

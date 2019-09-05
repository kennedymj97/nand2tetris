package compiler

import (
	"errors"
	"strconv"
	"unicode"
)

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

// TokenType is an enum for the different token types
type tokenType uint8

// Defined the possible token types
const (
	unknown tokenType = iota
	keyword
	symbol
	stringConst
	intConst
	identifier
)

func newTokenType(token string) tokenType {
	if isKeyword(token) {
		return keyword
	} else if isSymbol(rune(token[0])) {
		return symbol
	} else if isStringConstant(rune(token[0])) {
		return stringConst
	} else if isInt(token) {
		return intConst
	} else if isIdentifier(token) {
		return identifier
	}
	return unknown
}

func formatKeyword(token string) (updatedToken string) {
	return token
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

func formatIntegerConst(token string) (updatedToken string) {
	return token
}

func formatIdentifier(token string) (updatedToken string) {
	return token
}

type token struct {
	value     string
	tokenType tokenType
}

func newToken(tokenValue string) (*token, error) {
	tokenType := newTokenType(tokenValue)
	switch tokenType {
	case keyword:
		tokenValue := formatKeyword(tokenValue)
		return &token{tokenValue, tokenType}, nil
	case symbol:
		tokenValue := formatSymbol(tokenValue)
		return &token{tokenValue, tokenType}, nil
	case stringConst:
		tokenValue := formatStringConst(tokenValue)
		return &token{tokenValue, tokenType}, nil
	case intConst:
		tokenValue := formatIntegerConst(tokenValue)
		return &token{tokenValue, tokenType}, nil
	case identifier:
		tokenValue := formatIdentifier(tokenValue)
		return &token{tokenValue, tokenType}, nil
	}
	// return an error here if the token is unknown
	return &token{tokenValue, tokenType}, errors.New("invalid token")
}

func (t *token) isType() bool {
	if t.tokenType == identifier {
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

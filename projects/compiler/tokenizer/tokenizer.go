package tokenizer

import (
	"bytes"
	"unicode/utf8"
)

func ScanTokens(data []byte, atEOF bool) (advance int, token []byte, err error) {
	start := 0
	// skip leading whitespace
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if !isSpace(r) {
			break
		}
	}
	// scan until space or symbol
	for width, i := 0, start; i < len(data); i += width {
		var r rune
		r, width := utf8.DecodeRune(data[i:])

		if isSpace(r) || isSymbol(r) {
			return i + width, data[start:i], nil
		}
	}
}

func removeComment(data []byte) []byte {
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		return data[i+1 : len(data)]
	}
	return data
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

package tokenizer

import (
	"bytes"
	"unicode/utf8"
)

// ScanTokens is a split function for a Scanner that returns each token in a jack file.
func ScanTokens(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Skip leading spaces.
	var r rune
	var width int
	start := 0
	for width = 0; start < len(data); start += width {
		r, width = utf8.DecodeRune(data[start:])
		if !isSpace(r) {
			break
		}
	}

	// NEED A WAY TO PARSE OUT COMMENTS AND PARSE OUT STRINGS

	r, width = utf8.DecodeRune(data[start:])

	if isSymbol(r) {
		return start + width, data[start : start+width], nil
	}

	// Scan until space, marking end of word.
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

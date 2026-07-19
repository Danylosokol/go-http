package headers

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
)

const crlf = "\r\n"

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	indx := bytes.Index(data, []byte(crlf))
	if indx == -1 {
		return 0, false, nil
	}

	if indx == 0 {
		return 2, true, nil
	}

	headerLineParts := strings.SplitN(string(data[:indx]), ":", 2)
	if len(headerLineParts) != 2 {
		return 0, false, fmt.Errorf("Expected to have key and a value split by :")
	}

	key := strings.ToLower(headerLineParts[0])
	if strings.TrimSpace(key) != key {
		return 0, false, fmt.Errorf("Invalid spacing header")
	}

	if !validTokens([]byte(key)) {
		return 0, false, fmt.Errorf("Invalid character in header key")
	}

	value := strings.TrimSpace(headerLineParts[1])

	h.Set(key, value)

	return len(data[:indx]) + 2, false, nil
}

func (h Headers) Get(key string) (string, bool) {
	key = strings.ToLower(key)
	value, ok := h[key]
	return value, ok
}

var specialCharacters = []byte{'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~'}

func isValidChar(c byte) bool {
	if c >= 'a' && c <= 'z' ||
		c >= '0' && c <= '9' {
		return true
	}
	return slices.Contains(specialCharacters, c)
}

func validTokens(data []byte) bool {
	for _, c := range data {
		if !isValidChar(c) {
			return false
		}
	}
	return true
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	v, ok := h[key]
	if ok {
		value = strings.Join([]string{v, value}, ", ")
	}
	h[key] = value
}

func (h Headers) Override(key, value string) {
	key = strings.ToLower(key)
	h[key] = value
}

func (h Headers) Remove(key string) {
	key = strings.ToLower(key)
	delete(h, key)
}

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
	var specialCharacters = []string{"!", "#", "$", "%", "&", "'", "*", "+", "-", ".", "^", "_", "`", "|", "~"}
	fmt.Printf("data: %v \n", data)
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
	for _, r := range key{
		c := string(r)
		if !(c >= "a" && c <= "z") && !(c >= "0" && c <= "9") && !slices.Contains(specialCharacters, c) {
			return 0, false, fmt.Errorf("Invalid character in header key")
		}
	}

	value := strings.TrimSpace(headerLineParts[1])

	h.Set(key, value)

	return len(data[:indx]) + 2, false, nil
}

func (h Headers) Set(key, value string) {
	h[key] = value
}

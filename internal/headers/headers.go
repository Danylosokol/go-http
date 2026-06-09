package headers

import (
	"bytes"
	"fmt"
	"strings"
)

const crlf = "\r\n"

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	fmt.Printf("data: %v \n", data)
	indx := bytes.Index(data, []byte(crlf))
	if indx == -1 {
		return 0, false, nil
	} 

	if indx == 0 {
		return 0, true, nil
	}

	headerLineParts := strings.SplitN(string(data[:indx]), ":", 2)
	if len(headerLineParts) != 2 {
		return 0, false, fmt.Errorf("Expected to have key and a value split by :")
	}
	
	key := headerLineParts[0]
	if len(strings.TrimSpace(key)) != len(key) {
		return 0, false, fmt.Errorf("Invalid spacing header")
	}
	
	value := strings.TrimSpace(headerLineParts[1])
	
	h[key] = value
	
	return len(data[:indx]) + 2, false, nil
}

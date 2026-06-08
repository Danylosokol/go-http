package request

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"errors"
)

type ParserState int

const (
	Initialized ParserState = iota
	Done
)

type Request struct {
	RequestLine RequestLine
	ParserState ParserState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"
const bufferSize = 8

func (r *Request) parse(data []byte) (int, error) {
	switch r.ParserState {
	case Initialized:
		bytesRead, requestLine, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		if bytesRead == 0 {
			return 0, nil
		}

		r.RequestLine = *requestLine
		r.ParserState = Done
		return bytesRead, nil
	case Done:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	default:
		return 0, fmt.Errorf("error: unkown state")
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := Request{
		RequestLine: RequestLine{},
		ParserState: Initialized,
	}
	buffer := make([]byte, bufferSize)
	readToIndex := 0
	for request.ParserState != Done{
		if  readToIndex >= len(buffer) {
			newBufferCap := len(buffer) * 2
			tempBuffer := make([]byte, newBufferCap)
			copy(tempBuffer, buffer)
			buffer = tempBuffer
		}

		bytesRead, err := reader.Read(buffer[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				request.ParserState = Done
				break
			} 
			return nil, fmt.Errorf("we hit some weird error reading from the reader: %s", err)
		}
		readToIndex += bytesRead
		
		bytesParsed, err := request.parse(buffer)
		if err != nil {
			return nil, fmt.Errorf("we hit some error parsing the buffer: %s", err)
		}
		if bytesParsed == 0 {
			continue
		}
		copy(buffer, buffer[bytesParsed:])
		readToIndex -= bytesParsed
	}
	return &request, nil
}

func parseRequestLine(data []byte) (int, *RequestLine, error) {
	indx := bytes.Index(data, []byte(crlf))
	if indx == -1 {
		return 0, nil, nil
	}
	requestLineText := string(data[:indx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return 0, nil, err
	}
	return indx, requestLine, nil
}

func requestLineFromString(str string) (*RequestLine, error) {
	requestLineParts := strings.Split(str, " ")
	if len(requestLineParts) != 3 {
		return nil, fmt.Errorf("invalid number of parts in the request line. Expected 3, recieved: %d", len(requestLineParts))
	}
	method := requestLineParts[0]
	for _, c := range method {
		if c < 'A' || c > 'Z' {
			return nil, fmt.Errorf("invalid method name: expected capital alphabetic characters, recieved: %s", method)
		}
	}
	requestTarget := requestLineParts[1]

	httpVersion := strings.Split(requestLineParts[2], "/")
	if len(httpVersion) != 2 {
		return nil, fmt.Errorf("malformed http version")
	}

	httpPart := httpVersion[0]
	if httpPart != "HTTP" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", httpPart)
	}

	version := httpVersion[1]
	if version != "1.1" && version != "2" && version != "3" {
		return nil, fmt.Errorf("invalid http version: %s", version)
	}
	return &RequestLine{
		HttpVersion:   version,
		RequestTarget: requestTarget,
		Method:        method,
	}, nil
}

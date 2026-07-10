package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/danylo-sokol/go-http/internal/headers"
)

type ParserState int

const (
	requestStateInitialized ParserState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	ParserState ParserState
	Headers     headers.Headers
	Body        []byte
}

const crlf = "\r\n"
const bufferSize = 8
const contentLengthKey = "Content-Length"

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := Request{
		RequestLine: RequestLine{},
		ParserState: requestStateInitialized,
		Headers:     headers.NewHeaders(),
	}
	buffer := make([]byte, bufferSize)
	readToIndex := 0
	for request.ParserState != requestStateDone {
		if readToIndex >= len(buffer) {
			newBufferCap := len(buffer) * 2
			tempBuffer := make([]byte, newBufferCap)
			copy(tempBuffer, buffer)
			buffer = tempBuffer
		}
		numBytesRead, err := reader.Read(buffer[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if request.ParserState != requestStateDone {
					return nil, fmt.Errorf("incomplite request, in state %d, read n bytes on EOF: %d \n", request.ParserState, numBytesRead)
				}
			}
			return nil, fmt.Errorf("we hit some weird error reading from the reader: %s", err)
		}
		readToIndex += numBytesRead
		bytesParsed, err := request.parse(buffer[:readToIndex])
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

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.ParserState != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		if n == 0 {
			break
		}
		totalBytesParsed += n
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.ParserState {
	case requestStateInitialized:
		bytesRead, requestLine, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		if bytesRead == 0 {
			return 0, nil
		}

		r.RequestLine = *requestLine
		r.ParserState = requestStateParsingHeaders
		return bytesRead, nil
	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.ParserState = requestStateParsingBody
		}
		return n, nil
	case requestStateParsingBody:
		value, err := r.Headers.Get(contentLengthKey)
		if err != nil {
			if err.Error() == "Key not found" {
				r.ParserState = requestStateDone
				return 0, nil
			}
			return 0, err
		}
		r.Body = append(r.Body, data...)
		contentLength, parseErr := strconv.Atoi(value)
		if parseErr != nil {
			return 0, fmt.Errorf("error: could not parse Content-Length value: %v", parseErr)
		}
		if contentLength < len(r.Body) {
			return 0, fmt.Errorf("error: body is longer than reported Content-Length: %d", len(r.Body))
		}
		if contentLength == len(r.Body) {
			r.ParserState = requestStateDone
		}
		return len(data), nil
	case requestStateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	default:
		return 0, fmt.Errorf("error: unkown state")
	}
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
	return indx + 2, requestLine, nil
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

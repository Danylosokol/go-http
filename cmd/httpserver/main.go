package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/danylo-sokol/go-http/internal/request"
	"github.com/danylo-sokol/go-http/internal/response"
	"github.com/danylo-sokol/go-http/internal/server"
)

const port = 42069
const httpbinPrefix = "/httpbin/"
const httpbinBase = "https://httpbin.org"

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v\n", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	hasHttpbinPrefix := strings.HasPrefix(req.RequestLine.RequestTarget, httpbinPrefix)
	if hasHttpbinPrefix {
		handlerHttpbin(w, req)
		return
	}
	if req.RequestLine.RequestTarget == "/yourproblem" {
		handler400(w, req)
		return
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		handler500(w, req)
		return
	}
	handler200(w, req)
	return
}

func handlerHttpbin(w *response.Writer, r *request.Request) {
	httpbinRoute := strings.TrimPrefix(r.RequestLine.RequestTarget, httpbinPrefix)
	url := fmt.Sprintf("%s/%s", httpbinBase, httpbinRoute)

	res, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error making GET request to httpbin: %v\n", err)
		handler500(w, r)
		return
	}
	defer res.Body.Close()

	w.WriteStatusLine(response.StatusCodeSuccess)
	h := response.GetDefaultHeaders(0)
	h.Remove("Content-Length")
	h.Set("Transfer-Encoding", "chunked")
	w.WriteHeaders(h)

	buffer := make([]byte, 1024)
	for {
		n, err := res.Body.Read(buffer)
		fmt.Printf("read %d bytes from response body\n", n)
		if n > 0 {
			fmt.Printf("final chunk of size %v read\n", n)
			_, err := w.WriteChunkedBody(buffer[:n])
			if err != nil {
				fmt.Printf("error writing chunked body: %v\n", err)
				break
			}
		}
		if err == io.EOF {
			fmt.Printf("Finished reading response body\n")
			break
		}
		if err != nil {
			fmt.Printf("Error reading response body: %v\n", err)
			break
		}
	}
	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		fmt.Printf("error writing chunked body done: %v\n", err)
	}
}

func handler400(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeBadRequest)
	body := []byte(`<html>
<head>
<title>400 Bad Request</title>
</head>
<body>
<h1>Bad Request</h1>
<p>Your request honestly kinda sucked.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
	return
}

func handler500(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeInternalServerError)
	body := []byte(`<html>
<head>
<title>500 Internal Server Error</title>
</head>
<body>
<h1>Internal Server Error</h1>
<p>Okay, you know what? This one is on me.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler200(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeSuccess)
	body := []byte(`<html>
<head>
<title>200 OK</title>
</head>
<body>
<h1>Success!</h1>
<p>Your request was an absolute banger.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
	return
}

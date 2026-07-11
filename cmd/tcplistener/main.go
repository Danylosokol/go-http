package main

import (
	"fmt"
	"log"
	"net"

	"github.com/danylo-sokol/go-http/internal/request"
)

const port = ":42069"

func main() {
	tcplistener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("error listening to tcp traffic: %s\n", err)
	}
	defer tcplistener.Close()
	fmt.Println("listening for TCP traffic on", port)
	for {
		conn, err := tcplistener.Accept()
		if err != nil {
			log.Fatalf("error accepting connection: %s\n", err)
		}
		fmt.Printf("accepted connection from: %s\n", conn.RemoteAddr())

		request, errRequest := request.RequestFromReader(conn)

		if errRequest != nil {
			log.Fatalf("something went wrong with reading the request...")
		}

		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %s\n", request.RequestLine.Method)
		fmt.Printf("- Target: %s\n", request.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", request.RequestLine.HttpVersion)
		fmt.Printf("Headers:\n")
		for key, value := range request.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}
		fmt.Printf("Body:\n")
		fmt.Printf("%s\n", string(request.Body))
		fmt.Printf("connection to %s closed\n", conn.RemoteAddr())
	}
}

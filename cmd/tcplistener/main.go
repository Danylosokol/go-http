package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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

		linesChannel := getLinesChannel(conn)

		for line := range linesChannel {
			fmt.Println(line)
		}
		fmt.Printf("connection to %s closed", conn.RemoteAddr())
	}
}

func getLinesChannel(conn net.Conn) <-chan string {
	linesChannel := make(chan string)

	go func() {
		defer conn.Close()
		defer close(linesChannel)
		currentLine := ""
		for {
			buffer := make([]byte, 8)
			n, err := conn.Read(buffer)
			if err != nil {
				if currentLine != "" {
					linesChannel <- currentLine
					currentLine = ""
				}
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("error: %s\n", err.Error())
				break
			}

			str := string(buffer[:n])
			parts := strings.Split(str, "\n")
			for i := 0; i < len(parts)-1; i++ {
				newLine := currentLine + parts[i]
				linesChannel <- newLine
				currentLine = ""
			}
			currentLine += parts[len(parts)-1]
		}
	}()

	return linesChannel
}

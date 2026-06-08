package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const addr = "localhost:42069"

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Fatalf("Couldn't resolve address: %s; error: %s\n", addr, err)
	}

	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatalf("Couldn't establish udp connection: %s; error: %s\n", udpAddr.String(), err)
	}

	fmt.Printf("We established a connection: %s\n", udpConn.RemoteAddr())
	defer udpConn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Error reading string from stdin: %s", err)
		}
		fmt.Printf("Line is: %s", line)
		bytesWritten, err := udpConn.Write([]byte(line))
		if err != nil {
			log.Fatalf("Error writing line to udp: %s", err)
		}
		fmt.Printf("Bytes written: %d\n", bytesWritten)
	}
}

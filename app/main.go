package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to resolve ip:port")
		os.Exit(1)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn *net.TCPConn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.ToLower(line) == "ping" {
			conn.Write([]byte("+PONG\r\n"))
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from connection: ", err)
		os.Exit(1)
	}
}

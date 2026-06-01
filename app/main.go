package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

func main() {
	connectedClients := 0

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

	fmt.Println("Hermes is flying!!")

	var wg sync.WaitGroup
	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		connectedClients++

		fmt.Printf("accepted connection from: %s, connected clients: %d\n", conn.RemoteAddr(), connectedClients)

		wg.Go(func() {
			handleConnection(conn)
		})
	}
	wg.Wait()
	fmt.Println("all connections closed, shutting down...")
}

func handleConnection(conn *net.TCPConn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := strings.ToLower(scanner.Text())
		fmt.Println(line)
		switch line {
		case "ping":
			conn.Write([]byte("+PONG\r\n"))
		default:
			conn.Write([]byte(line))
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from connection: ", err)
		os.Exit(1)
	}
}

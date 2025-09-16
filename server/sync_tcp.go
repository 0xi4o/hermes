package server

import (
	"io"
	"log"
	"net"
	"strconv"

	"github.com/0xi4o/hermes/config"
)

func read_command(c net.Conn) (string, error) {
	var buf []byte = make([]byte, 512)
	n, err := c.Read(buf[:])
	if err != nil {
		return "", err
	}
	return string(buf[:n]), nil
}

func respond(cmd string, c net.Conn) error {
	if _, err := c.Write([]byte(cmd)); err != nil {
		return err
	}
	return nil
}

func RunSyncTCPServer() {
	log.Printf("starting hermes on host: %s, port: %d\n", config.Host, config.Port)

	con_clients := 0

	listener, err := net.Listen("tcp", config.Host+":"+strconv.Itoa(config.Port))
	if err != nil {
		panic(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		con_clients += 1
		log.Printf("client connected with address: %s, concurrent clients: %d\n", conn.RemoteAddr(), con_clients)

		for {
			cmd, err := read_command(conn)
			if err != nil {
				conn.Close()
				con_clients -= 1
				log.Printf("client disconnected: %s, concurrent clients: %d\n", conn.RemoteAddr(), con_clients)
				if err == io.EOF {
					break
				}
				log.Println(err)
			}
			log.Println(cmd)
			if err = respond(cmd, conn); err != nil {
				log.Print("err write: ", err)
			}
		}
	}
}

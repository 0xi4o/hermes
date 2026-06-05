package server

import (
	"errors"
	"fmt"
	"io"
	"net"

	"sync"

	"github.com/codecrafters-io/redis-starter-go/core"
	"github.com/codecrafters-io/redis-starter-go/data"
)

type Server struct {
	Host             string
	Port             int
	ConnectedClients int
	KV               map[string]any
}

func NewServer(host string, port int) *Server {
	kv := data.NewKV()
	return &Server{
		Host:             host,
		Port:             port,
		ConnectedClients: 0,
		KV:               kv,
	}
}

func (server *Server) Start(wg *sync.WaitGroup) error {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", server.Host, server.Port))
	if err != nil {
		return errors.New("failed to resolve ip:port")
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return errors.New("failed to bind to port 6379")
	}
	defer l.Close()

	fmt.Println("hermes is flying...")
	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			return fmt.Errorf("error accepting connection: %s\n", err.Error())
		}

		server.ConnectedClients++
		fmt.Printf("accepted connection from: %s, connected clients: %d\n", conn.RemoteAddr(), server.ConnectedClients)
		wg.Go(func() {
			server.handleConnection(conn)
		})
	}
}

func (server *Server) handleConnection(conn *net.TCPConn) {
	defer conn.Close()
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf[:])
		if err != nil {
			server.ConnectedClients--
			fmt.Printf("client disconnected: %s, connected clients: %d\n", conn.RemoteAddr(), server.ConnectedClients)
			if err == io.EOF {
				break
			}
			continue
		}
		resp := core.NewRESP()
		if n > 0 {
			err = resp.Decode(buf[:n])
			if err != nil {
				fmt.Println("unable to decode bytes into RESP struct: ", err)
				break
			}
			command, err := core.ParseCommand(resp)
			if err != nil {
				fmt.Println("unable to parse command from RESP struct: ", err)
				break
			}
			response, err := command.Execute(server.KV)
			if err != nil {
				fmt.Println("unable to execute command: ", err)
				break
			}
			sendResponse(conn, response)
		}
	}
}

func sendResponse(conn *net.TCPConn, response core.Response) {
	data, err := response.Encode()
	if err != nil {
		fmt.Println("error encoding response: ", err)
	}
	if _, err := conn.Write([]byte(data)); err != nil {
		fmt.Println("error writing to connection: ", err)
	}
}

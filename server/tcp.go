package server

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/core"
)

type Server struct {
	Host             string
	Port             int
	ConnectedClients int
}

func NewServer(host string, port int) *Server {
	return &Server{
		Host:             host,
		Port:             port,
		ConnectedClients: 0,
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
			err = resp.UnmarshalRESP(buf[:n])
			if err != nil {
				break
			}
			var command string
			var ok bool
			var args strings.Builder
			switch v := resp.Value.(type) {
			case []core.RESP:
				command, ok = v[0].Value.(string)
				if !ok {
					break
				}
				for i := 1; i < len(v); i++ {
					arg, ok := v[i].Value.(string)
					if !ok {
						break
					}
					args.WriteString(arg)
					args.WriteString(" ")
				}
			default:
				command, ok = v.(string)
				if !ok {
					break
				}
			}
			command = strings.ToLower(command)
			fmt.Println(command)
			switch command {
			case "ping":
				resp.Response = "pong"
				sendResponse(conn, resp)
			case "echo":
				resp.Response = strings.TrimSpace(args.String())
				sendResponse(conn, resp)
			default:
				resp.Response = string(buf[:n])
				sendResponse(conn, resp)
			}
		}
	}
}

func sendResponse(conn *net.TCPConn, resp core.RESP) {
	data, err := resp.MarshalRESP()
	if err != nil {
		fmt.Println("error marshaling resp: ", err)
	}
	if _, err := conn.Write([]byte(data)); err != nil {
		fmt.Println("error writing to connection: ", err)
	}
}

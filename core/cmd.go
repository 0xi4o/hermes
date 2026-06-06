package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Command struct {
	Cmd  string
	Args []string
}

func NewCommand() *Command {
	return &Command{}
}

func ParseCommand(resp RESP) (Command, error) {
	var cmd string
	var ok bool
	var args []string
	switch v := resp.Value.(type) {
	case []RESP:
		cmd, ok = v[0].Value.(string)
		if !ok {
			return Command{}, errors.New("unable to parse command from resp message")
		}
		if len(v) > 1 {
			for i := 1; i < len(v); i++ {
				switch v := v[i].Value.(type) {
				case int:
					s := strconv.Itoa(v)
					args = append(args, s)
				case string:
					args = append(args, v)
				}
			}
		}
	default:
		cmd, ok = v.(string)
		if !ok {
			return Command{}, errors.New("unable to parse command from resp message")
		}
	}
	return Command{Cmd: strings.ToUpper(cmd), Args: args}, nil
}

func (c *Command) Execute(kv map[string]any) (Response, error) {
	response := NewResponse()
	switch c.Cmd {
	case "ECHO":
		var data string
		if len(c.Args) > 1 {
			data = strings.Join(c.Args, " ")
		} else if len(c.Args) == 1 {
			data = c.Args[0]
		} else {
			data = ""
		}
		response.Type = BulkString
		response.Data = strings.TrimSpace(data)
		return response, nil
	case "GET":
		if len(c.Args) > 1 {
			return Response{}, errors.New("too many arguments for GET command")
		}
		if len(c.Args) < 1 {
			return Response{}, errors.New("too few arguments for GET command")
		}
		key := c.Args[0]
		response.Type = BulkString
		data, ok := kv[key]
		if !ok {
			response.Data = nil
			return response, nil
		}
		response.Data = data
		return response, nil
	case "PING":
		response.Type = SimpleString
		response.Data = "PONG"
		return response, nil
	case "SET":
		key, value := c.Args[0], c.Args[1]
		kv[key] = value
		response.Type = SimpleString
		response.Data = "OK"
		return response, nil
	default:
		return Response{}, fmt.Errorf("unknown command: %s", c.Cmd)
	}
}

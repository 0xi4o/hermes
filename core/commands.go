package core

import (
	"errors"
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
				arg, ok := v[i].Value.(string)
				if !ok {
					return Command{}, errors.New("unable to parse arguments from resp message")
				}
				args = append(args, arg)
			}
		}
	default:
		cmd, ok = v.(string)
		if !ok {
			return Command{}, errors.New("unable to parse command from resp message")
		}
	}
	return Command{Cmd: cmd, Args: args}, nil
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
	case "GET":
		if len(c.Args) > 1 {
			return Response{}, errors.New("too many arguments for GET command")
		}
		if len(c.Args) < 1 {
			return Response{}, errors.New("too few arguments for GET command")
		}
		key := c.Args[0]
		response.Type = BulkString
		response.Data = kv[key]
	case "PING":
		response.Type = SimpleString
		response.Data = "PONG"
	case "SET":
		if len(c.Args) > 2 {
			return Response{}, errors.New("too many arguments for SET command")
		}
		if len(c.Args) < 2 {
			return Response{}, errors.New("too few arguments for SET command")
		}
		key, value := c.Args[0], c.Args[1]
		kv[key] = value
		response.Type = SimpleString
		response.Data = "OK"
	default:
		return Response{}, errors.New("unknown command")
	}
	return response, nil
}

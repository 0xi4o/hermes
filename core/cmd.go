package core

import (
	"errors"
	"fmt"
	// "math"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/data"
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

func (c *Command) Execute() (Response, error) {
	switch c.Cmd {
	case "ECHO":
		return evalECHO(c.Args)
	case "GET":
		return evalGET(c.Args)
	case "LLEN":
		return evalLLEN(c.Args)
	case "LPUSH":
		return evalLPUSH(c.Args)
	case "LRANGE":
		return evalLRANGE(c.Args)
	case "PING":
		return evalPING(c.Args)
	case "RPUSH":
		return evalRPUSH(c.Args)
	case "SET":
		return evalSET(c.Args)
	case "TTL":
		return evalTTL(c.Args)
	default:
		return Response{}, fmt.Errorf("unknown command: %s", c.Cmd)
	}
}

func evalECHO(args []string) (Response, error) {
	var data string
	if len(args) > 1 {
		data = strings.Join(args, " ")
	} else if len(args) == 1 {
		data = args[0]
	} else {
		data = ""
	}
	return Response{Type: BulkString, Data: strings.TrimSpace(data)}, nil
}

func evalGET(args []string) (Response, error) {
	if len(args) != 1 {
		return Response{}, errors.New("wrong number of arguments for GET")
	}

	key := args[0]

	item, err := data.Store.Cache.Get(key)
	if err != nil {
		return Response{Type: BulkString}, nil
	}

	if item.ExpiresAt != -1 && item.ExpiresAt <= time.Now().UnixMilli() {
		return Response{Type: BulkString}, nil
	}

	return Response{Type: BulkString, Data: item.Value}, nil
}

func evalLLEN(args []string) (Response, error) {
	if len(args) != 1 {
		return Response{}, errors.New("wrong number of arguments for LLEN")
	}

	key := args[0]

	items, err := data.Store.Cache.Get(key)
	if err != nil {
		return Response{Type: SimpleError, Data: err.Error()}, err
	}

	return Response{Type: Integer, Data: items.Length}, nil
}

func evalLPUSH(args []string) (Response, error) {
	if len(args) <= 1 {
		return Response{}, errors.New("wrong number of arguments for SET")
	}

	key := args[0]

	err := data.Store.Cache.Prepend(key, args[1:])
	if err != nil {
		return Response{}, err
	}
	items, err := data.Store.Cache.Get(key)
	if err != nil {
		return Response{}, err
	}

	return Response{Type: Integer, Data: items.Length}, nil
}

func evalLRANGE(args []string) (Response, error) {
	if len(args) != 3 {
		return Response{}, errors.New("wrong number of arguments for LRANGE")
	}

	key := args[0]

	start, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return Response{}, errors.New("wrong number of arguments for LRANGE")
	}

	stop, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return Response{}, errors.New("wrong number of arguments for LRANGE")
	}

	items, err := data.Store.Cache.Get(key)
	if err != nil {
		return Response{Type: Array, Data: []string{}}, nil
	}

	if start >= items.Length {
		return Response{Type: Array, Data: []string{}}, nil
	}
	if start < 0 {
		start = max(0, start+items.Length)
	}
	if stop < 0 {
		stop = max(0, stop+items.Length)
	}
	if stop >= items.Length {
		stop = items.Length - 1
	}

	switch v := items.Value.(type) {
	case []string:
		return Response{Type: Array, Data: v[start : stop+1]}, nil
	default:
		return Response{}, errors.New("value is not a list")
	}
}

func evalPING(_ []string) (Response, error) {
	return Response{Type: SimpleString, Data: "PONG"}, nil
}

func evalRPUSH(args []string) (Response, error) {
	if len(args) <= 1 {
		return Response{}, errors.New("wrong number of arguments for SET")
	}

	key := args[0]

	err := data.Store.Cache.Append(key, args[1:])
	if err != nil {
		return Response{}, err
	}
	items, err := data.Store.Cache.Get(key)
	if err != nil {
		return Response{}, err
	}

	return Response{Type: Integer, Data: items.Length}, nil
}

func evalSET(args []string) (Response, error) {
	if len(args) <= 1 {
		return Response{}, errors.New("wrong number of arguments for SET")
	}

	key, value := args[0], args[1]

	var expiryDurationMs int64 = -1

	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "EX", "ex":
			i++
			if i == len(args) {
				return Response{}, errors.New("syntax error")
			}

			expiryDurationSec, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return Response{}, errors.New("EX value is not an integer or out of range")
			}
			expiryDurationMs = expiryDurationSec * 1000
			data.Store.Cache.Put(key, value, expiryDurationMs)
			return Response{Type: SimpleString, Data: "OK"}, nil
		case "PX", "px":
			i++
			if i == len(args) {
				return Response{}, errors.New("syntax error")
			}

			expiryDurationMs, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return Response{}, errors.New("PX value is not an integer or out of range")
			}
			data.Store.Cache.Put(key, value, expiryDurationMs)
			return Response{Type: SimpleString, Data: "OK"}, nil
		default:
			return Response{}, errors.New("syntax error, unknown argument for SET")
		}
	}

	data.Store.Cache.Put(key, value, expiryDurationMs)
	return Response{Type: SimpleString, Data: "OK"}, nil
}

func evalTTL(args []string) (Response, error) {
	if len(args) != 1 {
		return Response{}, errors.New("wrong number of arguments for TTL")
	}

	key := args[0]

	item, err := data.Store.Cache.Get(key)
	if err != nil {
		return Response{Type: Integer, Data: int64(-2)}, nil
	}

	if item.ExpiresAt == -1 {
		return Response{Type: Integer, Data: int64(-1)}, nil
	}

	durationMs := item.ExpiresAt - time.Now().UnixMilli()

	if durationMs < 0 {
		return Response{Type: Integer, Data: int64(-2)}, nil
	}

	return Response{Type: Integer, Data: int64(durationMs / 1000)}, nil
}

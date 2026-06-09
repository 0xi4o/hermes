package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const RESP_NIL string = "$-1\r\n"

type DataType int

const (
	SimpleString DataType = iota
	SimpleError
	Integer
	BulkString
	Array
)

type RESP struct {
	Type     DataType
	Value    any
	Offset   int
	Response Response
}

func NewRESP() RESP {
	return RESP{Offset: 0}
}

func (resp *RESP) Decode(data []byte) error {
	if len(data) == 0 {
		return errors.New("no data")
	}

	resp.Offset = 1

	switch data[0] {
	case '+':
		resp.Type = SimpleString
		value, offset, err := decodeSimpleString(data)
		if err != nil {
			return err
		}
		resp.Offset = offset
		resp.Value = value
	case '-':
		resp.Type = SimpleError
		value, offset, err := decodeSimpleString(data)
		if err != nil {
			return err
		}
		resp.Offset = offset
		resp.Value = value
	case ':':
		resp.Type = Integer
		value, offset, err := decodeInteger(data)
		if err != nil {
			return err
		}
		resp.Offset = offset
		resp.Value = value
	case '$':
		resp.Type = BulkString
		length, offset, err := decodeInteger(data)
		if err != nil {
			return err
		}
		resp.Offset = length + offset + 2
		resp.Value = string(data[offset : length+offset])
	case '*':
		resp.Type = Array
		length, offset, err := decodeInteger(data)
		if err != nil {
			return err
		}
		value, offset, err := decodeArray(data, length, offset)
		if err != nil {
			return err
		}
		resp.Offset = offset
		resp.Value = value
	}
	return nil
}

func decodeSimpleString(data []byte) (value string, offset int, err error) {
	pos := 1
	var s strings.Builder

	for ; data[pos] != '\r'; pos++ {
		err = s.WriteByte(data[pos])
		if err != nil {
			return "", 0, err
		}
	}

	return s.String(), pos + 2, nil
}

func decodeInteger(data []byte) (value int, offset int, err error) {
	s, offset, err := decodeSimpleString(data)
	if err != nil {
		return 0, 0, err
	}
	value, err = strconv.Atoi(s)
	if err != nil {
		return 0, 0, err
	}
	return value, offset, err
}

func decodeArray(data []byte, length, delta int) (values []RESP, offset int, err error) {
	values = []RESP{}
	offset = delta
	for range length {
		elem := NewRESP()
		err := elem.Decode(data[offset:])
		if err != nil {
			return []RESP{}, 0, err
		}
		offset += elem.Offset
		values = append(values, elem)
	}
	return values, offset, nil
}

type Response struct {
	Type DataType
	Data any
}

func NewResponse() Response {
	return Response{}
}

func (response *Response) Encode() ([]byte, error) {
	var data string
	switch response.Type {
	case Array:
		panic("unimplemented")
	case Integer:
		body, ok := response.Data.(int64)
		if !ok {
			return []byte{}, errors.New("cannot cast to type: int64")
		}
		data = fmt.Sprintf(":%d\r\n", body)
		return []byte(data), nil
	case BulkString:
		if response.Data == nil {
			return []byte(RESP_NIL), nil
		}
		body, ok := response.Data.(string)
		if !ok {
			return []byte{}, errors.New("cannot cast to type: string")
		}
		data = fmt.Sprintf("$%d\r\n%s\r\n", len(body), body)
		return []byte(data), nil
	case SimpleError:
		panic("unimplemented")
	case SimpleString:
		body, ok := response.Data.(string)
		if !ok {
			return []byte{}, errors.New("cannot cast to type: string")
		}
		data = fmt.Sprintf("+%s\r\n", body)
		return []byte(data), nil
	default:
		panic(fmt.Sprintf("unexpected core.DataType: %#v", response.Type))
	}
}

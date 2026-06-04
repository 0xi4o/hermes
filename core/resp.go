package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

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
	Response any
}

func NewRESP() RESP {
	return RESP{Offset: 0}
}

func (resp *RESP) MarshalRESP() ([]byte, error) {
	var data string
	switch resp.Type {
	case Array:
		response, ok := resp.Response.(string)
		if !ok {
			return []byte{}, errors.New("cannot cast to type: string")
		}
		data = fmt.Sprintf("$%d\r\n%s\r\n", len(response), response)
	// case Integer:
	// case BulkString:
	// 	fallthrough
	// case SimpleError:
	// 	fallthrough
	case SimpleString:
		response, ok := resp.Response.(string)
		if !ok {
			return []byte{}, errors.New("cannot cast to type: string")
		}
		data = fmt.Sprintf("+%s\r\n", response)
	default:
		panic(fmt.Sprintf("unexpected core.DataType: %#v", resp.Type))
	}
	return []byte(data), nil
}

func (resp *RESP) UnmarshalRESP(data []byte) error {
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
		err := elem.UnmarshalRESP(data[offset:])
		if err != nil {
			return []RESP{}, 0, err
		}
		offset += elem.Offset
		values = append(values, elem)
	}
	return values, offset, nil
}

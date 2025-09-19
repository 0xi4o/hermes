package core

import (
	"errors"
)

func Decode(data []byte) (any, error) {
	if len(data) == 0 {
		return nil, errors.New("no data")
	}
	value, _, err := DecodeOne(data)
	return value, err
}

func DecodeOne(data []byte) (any, int, error) {
	if len(data) == 0 {
		return nil, 0, errors.New("no data")
	}

	switch data[0] {
	case '+':
		return readSimpleString(data)
	case '-':
		return readError(data)
	case ':':
		return readInt64(data)
	case '$':
		return readBulkString(data)
	case '*':
		return readArray(data)
	}
	return nil, 0, nil
}

func readArray(data []byte) (any, int, error) {
	pos := 1

	count, delta := readLength(data[pos:])
	pos += delta

	var elems []any = make([]any, count)
	for i := range elems {
		elem, delta, err := DecodeOne(data[pos:])
		if err != nil {
			return nil, 0, err
		}
		elems[i] = elem
		pos += delta
	}

	return elems, pos, nil
}

func readBulkString(data []byte) (any, int, error) {
	pos := 1
	length, delta := readLength(data[pos:])
	pos += delta

	return string(data[pos:(pos + length)]), pos + length + 2, nil
}

func readInt64(data []byte) (any, int, error) {
	pos := 1
	var value int64 = 0

	for ; data[pos] != '\r'; pos++ {
		value = value*10 + int64(data[pos]-'0')
	}
	return value, pos + 2, nil
}

func readError(data []byte) (any, int, error) {
	return readSimpleString(data)
}

func readSimpleString(data []byte) (any, int, error) {
	pos := 1

	for ; data[pos] != '\r'; pos++ {
	}

	return string(data[1:pos]), pos + 2, nil
}

func readLength(data []byte) (int, int) {
	length, pos := 0, 0

	for pos = range data {
		byte := data[pos]
		if !(byte > '0' && byte < '9') {
			return length, pos + 2
		}
		length = length*10 + int(byte-'0')
	}

	return 0, 0
}

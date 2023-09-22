package main

// RESP and parse it.
import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

const (
	ARRAY         = '*'
	BULK          = '$'
	INTEGER       = ':'
	SIMPLE_STRING = '+'
	ERROR         = '-'
)

type Value struct {
	typ   string
	num   int
	str   string
	bulk  string
	array []Value
}

type Resp struct {
	Reader *bufio.Reader
}

func NewResp(reader io.Reader) *Resp {
	return &Resp{Reader: bufio.NewReader(reader)}
}

// func to read line
func (r *Resp) ReadLine() (line []byte, n int, err error) {
	for {
		b, err := r.Reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}
		n += 1
		line = append(line, b)
		if len(line) >= 2 && line[n-2] == '\r' {
			break
		}
	}
	return line, n, nil
}

// func to read integer from the buffer
func (r *Resp) readInteger() (x int, n int, err error) {
	line, n, err := r.ReadLine()
	if err != nil {
		return 0, 0, err
	}
	n += 1
	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return int(i64), n, nil
}

// func to read bulk string from the buffer
func (r *Resp) read_bulk() (Value, error) {
	size, _, err := r.readInteger()
	if err != nil {
		return Value{}, err
	}
	line, _, err := r.ReadLine()
	if err != nil {
		return Value{}, err
	}
	return Value{typ: "bulk", num: size, bulk: string(line)}, nil
}

// func to read array from the buffer
func (r *Resp) read_array() (Value, error) {
	size, _, err := r.readInteger()
	if err != nil {
		return Value{}, err
	}
	var array []Value
	for i := 0; i < size; i++ {
		val, err := r.ReadValue()
		if err != nil {
			return Value{}, err
		}
		array = append(array, val)
	}
	return Value{typ: "array", num: size, array: array}, nil
}

// func to read value from the buffer
func (r *Resp) ReadValue() (Value, error) {
	b, err := r.Reader.ReadByte()
	if err != nil {
		return Value{}, err
	}
	switch b {
	case BULK:
		return r.read_bulk()
	case ARRAY:
		return r.read_array()
	default:
		return Value{}, fmt.Errorf("unknown type:%c", b)
	}
}

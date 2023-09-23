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

func NewRespReader(reader io.Reader) *Resp {
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
		if len(line) >= 2 && line[len(line)-2] == '\r' {
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
		val, err := r.Read()
		if err != nil {
			return Value{}, err
		}
		array = append(array, val)
	}
	return Value{typ: "array", num: size, array: array}, nil
}

// func to read value from the buffer
func (r *Resp) Read() (Value, error) {
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


/*
@breif: This function is used to create a writer resp 
*/

func (v Value) Marshel() []byte{
	var resp []byte
	switch v.typ {
	case "string":
		resp = append(resp, SIMPLE_STRING)
		resp = append(resp, []byte(v.str)...)
		resp = append(resp, '\r')
		resp = append(resp, '\n')
	case "bulk":
		resp = append(resp, BULK)
		resp = append(resp, []byte(strconv.Itoa(v.num))...)
		resp = append(resp, '\r')
		resp = append(resp, '\n')
		resp = append(resp, []byte(v.bulk)...)
		resp = append(resp, '\r')
		resp = append(resp, '\n')
	case "array":
		resp = append(resp, ARRAY)
		resp = append(resp, []byte(strconv.Itoa(v.num))...)
		resp = append(resp, '\r')
		resp = append(resp, '\n')
		for _, val := range v.array {
			resp = append(resp, val.Marshel()...)
		}
	case "error":
		resp = append(resp, ERROR)
		resp = append(resp, []byte(v.str)...)
		resp = append(resp, '\r')
		resp = append(resp, '\n')
	case "null":
		resp = append(resp, BULK)
		resp = append(resp, '-')
		resp = append(resp, '1')
		resp = append(resp, '\r')
		resp = append(resp, '\n')

	}
	return resp
}

type Writer struct {
	writer io.Writer
}

func NewRespWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}
func (w *Writer) Write(v Value) error {
	_, err := w.writer.Write(v.Marshel())
	if err != nil {
		return err
	}
	return nil
}
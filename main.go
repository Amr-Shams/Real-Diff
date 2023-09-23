package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

)



func create_server(){
	listener, err := net.Listen("tcp", "localhost:6379")
	checkError(err)
	conn, err := listener.Accept()
	checkError(err)
	defer conn.Close() 
	for {
		resp := NewRespReader(conn)
		value, err := resp.ReadValue()
		checkError(err)
		_ = value
		Writer := NewRespWriter(conn)
		Writer.WriteValue(Value{typ: "string", str: "Hello World"})
	}
}
func checkError(err error) {
	if err != nil {
		if err == io.EOF {
			fmt.Println("Client closed connection")
			os.Exit(0)
		}
		log.Fatal("Error Occured: ", err)
		os.Exit(1)
	}
}


func main() {
	create_server()
}
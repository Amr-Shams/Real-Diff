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
	defer conn.Close() // close connection before exit
	for {
		resp := NewResp(conn)
		value, err := resp.ReadValue()
		checkError(err)
		fmt.Println(value)
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
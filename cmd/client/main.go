package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"scs/internal/protocol"
	"scs/internal/transport"
)

func main() {
	conn, err := net.Dial("tcp", os.Getenv("SERVER_ADDR"))
	if err != nil {
		log.Fatal(err)
	}
	defer func(conn net.Conn) {
		err = conn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(conn)

	responseData, err := transport.Receive(conn)
	if err != nil {
		log.Fatal(err)
	}

	response, _ := protocol.Decode(responseData)
	fmt.Println(response.Body)
}

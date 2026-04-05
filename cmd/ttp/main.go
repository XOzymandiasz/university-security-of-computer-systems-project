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
	listener, err := net.Listen("tcp", ":"+os.Getenv("PORT"))
	if err != nil {
		log.Fatal(err)
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(listener)

	fmt.Println("Third part listening on :8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println(err)
		}
	}(conn)

	data, err := transport.Receive(conn)
	if err != nil {
		log.Println(err)
		return
	}

	msg, _ := protocol.Decode(data)
	fmt.Println("Received:", msg.Type)

	if msg.Type == "Ping" {
		//response := protocol.Message{
		//	Type: "Pong",
		//	Body: "Cert",
		//}

		//encoded, _ := protocol.Encode(response)
		//err := transport.Send(conn, encoded)
		//if err != nil {
		//	return
		//}
	}
}

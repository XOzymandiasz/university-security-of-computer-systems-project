package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"scs/internal/identity"

	"scs/internal/protocol"
	"scs/internal/transport"
)

const baseDir = "/tmp/scs/ttp/"

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

	if msg.Type == "TTP_INIT" {
		response := protocol.Message{
			Type: "TTP_INIT",
			Body: identity.LoadRegistrationData(baseDir),
		}

		encoded, err := protocol.Encode(response)
		if err != nil {
			log.Println(err)
			return
		}
		err = transport.Send(conn, encoded)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"scs/internal/ttp"

	"scs/internal/identity"
	"scs/internal/protocol"
	"scs/internal/transport"
)

const baseDir = "/tmp/scs/server"

func main() {
	identity.EnsureIdentity(baseDir)
	registerToTtp()
	listener, err := net.Listen("tcp", ":"+os.Getenv("PORT"))
	if err != nil {
		log.Fatal(err)
	}
	defer func(listener net.Listener) {
		err = listener.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(listener)

	fmt.Println("Server listening on :8080")

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
		//	Body: ,
		//}

		//encoded, _ := protocol.Encode(response)
		//err := transport.Send(conn, encoded)
		//if err != nil {
		return
		//}
	}
}

func registerToTtp() {
	data := identity.LoadRegistrationData(baseDir)

	addr := os.Getenv("TTP_ADDR")
	if addr == "" {
		addr = "localhost:8081"
	}

	err := ttp.Register(addr, data)
	if err != nil {
		log.Fatal(err)
	}
}

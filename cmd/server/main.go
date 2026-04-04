package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"scs/internal/protocol"
	"scs/internal/transport"
)

func main() {
	registerToTtp()
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
		response := protocol.Message{
			Type: "Pong",
			Body: "Hello from server",
		}

		encoded, _ := protocol.Encode(response)
		err := transport.Send(conn, encoded)
		if err != nil {
			return
		}
	}
}

func registerToTtp() {
	ttpAddr := os.Getenv("TTP_ADDR")
	if ttpAddr == "" {
		ttpAddr = "localhost:8081"
	}

	var conn net.Conn
	var err error

	for i := 0; i < 15; i++ {
		conn, err = net.Dial("tcp", ttpAddr)
		if err == nil {
			break
		}

		log.Printf("waiting for TTP at %s: %v", ttpAddr, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal(err)
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(conn)

	msg := protocol.Message{
		Type: "Ping",
		Body: "Server's key",
	}

	encoded, _ := protocol.Encode(msg)
	err = transport.Send(conn, encoded)
	if err != nil {
		log.Fatal(err)
	}

	responseData, err := transport.Receive(conn)
	if err != nil {
		log.Fatal(err)
	}

	response, _ := protocol.Decode(responseData)
	fmt.Println(response.Body)
}

package main

import (
	"fmt"
	"log"
	"net"
	"scs-project/internal/crypto"
	"scs-project/internal/protocol"
	"scs-project/internal/transport"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(conn)

	privateKey, err := rsa.GenerateKeyPair()
	if err != nil {
		return
	}

	msg := protocol.Message{
		Type: "REGISTER",
		Body: privateKey,
	}

	encoded, _ := protocol.Encode(msg)
	err = transport.Send(conn, encoded)
	if err != nil {
		return
	}

	responseData, err := transport.Receive(conn)
	if err != nil {
		log.Fatal(err)
	}

	response, _ := protocol.Decode(responseData)
	fmt.Println(response.Body)
}

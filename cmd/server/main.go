package main

import (
	"crypto/rsa"
	"fmt"
	"log"
	"net"
	"os"
	"scs/internal/protocol"
	"scs/internal/transport"
	ttpClient "scs/internal/ttpClient"

	"scs/internal/identity"
)

const baseDir = "/tmp/scs/server/"

//func main() {
//	app, err := client.NewAppFromEnv()
//	if err != nil {
//		log.Fatalln(err)
//	}
//
//	if err = app.Run(); err != nil {
//		log.Fatalln(err)
//	}
//}

func main() {
	addr, err := ttpClient.AddrFromEnv("TTP_ADDR")
	if err != nil {
		log.Fatalln(err)
	}
	var ttpPublicKey *rsa.PublicKey
	ttpPublicKey, err = ttpClient.Init(addr)
	if err != nil {
		log.Fatalln(err)
	}

	identity.EnsureIdentity(baseDir)
	data := identity.LoadRegistrationData(baseDir)

	//var certificateBase64 string
	_, err = ttpClient.Register(addr, ttpPublicKey, data)
	if err != nil {
		log.Fatalln(err)
	}

	startApi()
}

func startApi() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT must be set")
	}

	addr := ":" + port

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer func(listener net.Listener) {
		err = listener.Close()
		if err != nil {

		}
	}(listener)

	fmt.Println("server TCP listening on", addr)

	for {
		var conn net.Conn
		conn, err = listener.Accept()
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

		}
	}(conn)
	var err error
	var data []byte
	data, err = transport.Receive(conn)
	if err != nil {
		log.Println(err)
		return
	}
	var msg protocol.Message
	msg, err = protocol.Decode(data)
	if err != nil {
		sendError(conn, err)
		return
	}

	switch msg.Type {
	case "READ_MESSAGE":
		handleReadMessage(conn)

	default:
		sendError(conn, fmt.Errorf("unknown message type: %s", msg.Type))
	}
}

func handleReadMessage(conn net.Conn) {
	message, err := readMessage()
	if err != nil {
		sendError(conn, err)
		return
	}

	response := protocol.Message{
		Type: "MESSAGE",
		Body: message,
	}
	var encoded []byte
	encoded, err = protocol.Encode(response)
	if err != nil {
		log.Println(err)
		return
	}

	if err = transport.Send(conn, encoded); err != nil {
		log.Println(err)
	}
}

func readMessage() (string, error) {
	data, err := os.ReadFile("/app/message")
	if err != nil {
		return "", fmt.Errorf("read message file: %w", err)
	}

	return string(data), nil
}

func sendError(conn net.Conn, err error) {
	log.Println(err)

	response := protocol.Message{
		Type: "ERROR",
		Body: err.Error(),
	}

	encoded, encodeErr := protocol.Encode(response)
	if encodeErr != nil {
		log.Println(encodeErr)
		return
	}

	if sendErr := transport.Send(conn, encoded); sendErr != nil {
		log.Println(sendErr)
	}
}

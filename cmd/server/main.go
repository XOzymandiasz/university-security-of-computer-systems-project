package main

import (
	"crypto/rsa"
	"fmt"
	"log"
	"net"
	"os"
	"scs/internal/protocol"
	"scs/internal/transport"
	"scs/internal/ttp"

	"scs/internal/identity"
)

const baseDir = "/tmp/scs/server/"

func main() {
	ttpPublicKey := initToTtp()
	identity.EnsureIdentity(baseDir)
	registerToTtp(ttpPublicKey)

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
	defer listener.Close()

	fmt.Println("server TCP listening on", addr)

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
	log.Println("HTTP /api/message hit:")
	defer conn.Close()

	data, err := transport.Receive(conn)
	if err != nil {
		log.Println(err)
		return
	}

	msg, err := protocol.Decode(data)
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

	encoded, err := protocol.Encode(response)
	if err != nil {
		log.Println(err)
		return
	}

	if err = transport.Send(conn, encoded); err != nil {
		log.Println(err)
	}
}

func readMessage() (string, error) {
	data, err := os.ReadFile("/app/cmd/server/message.txt")
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

func initToTtp() *rsa.PublicKey {
	addr := os.Getenv("TTP_ADDR")
	if addr == "" {
		log.Fatal("TTP_ADDR env variable not set")
	}

	ttpPublicKey, err := ttp.Init(addr)
	if err != nil {
		return nil
	}

	if ttpPublicKey == nil {
		log.Fatal("TTP public key is nil")
	}

	return ttpPublicKey
}

func registerToTtp(ttpPublicKey *rsa.PublicKey) {
	if ttpPublicKey == nil {
		log.Fatal("cannot register to TTP: public key is nil")
	}
	data := identity.LoadRegistrationData(baseDir)
	encryptedID, err := identity.EncryptWithPublicKeyBase64([]byte(data.ID), ttpPublicKey)
	if err != nil {
		log.Fatal(err)
	}

	data.ID = encryptedID

	addr := os.Getenv("TTP_ADDR")
	if addr == "" {
		log.Fatal("TTP_ADDR env variable not set")
	}

	err = ttp.Register(addr, data)
	if err != nil {
		log.Fatal(err)
	}
}

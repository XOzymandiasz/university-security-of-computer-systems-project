package main

import (
	"encoding/json"
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
	identity.EnsureIdentity(baseDir)
	ttpPort := os.Getenv("PORT")
	listener, err := net.Listen("tcp", ":"+ttpPort)
	if err != nil {
		log.Fatal(err)
	}
	defer func(listener net.Listener) {
		err = listener.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(listener)

	fmt.Println("Third part listening on " + ttpPort)

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
			log.Println(err)
		}
	}(conn)

	data, err := transport.Receive(conn)
	if err != nil {
		log.Println(err)
		return
	}

	msg, _ := protocol.Decode(data)

	switch msg.Type {
	case "TTP_INIT":
		handleTtpInit(conn)
	case "REGISTER":
		handleRegister(conn, msg)
	default:
		handleUnknown(conn, msg)
	}
}

func handleTtpInit(conn net.Conn) {
	responseData := identity.LoadRegistrationData(baseDir)

	response := protocol.Message{
		Type: "TTP_PUBLIC_KEY",
		Body: responseData.EncPublicKey,
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

func handleRegister(conn net.Conn, msg protocol.Message) {
	reg, err := registrationDataFromBody(msg.Body)
	if err != nil {
		sendError(conn, err)
		return
	}

	ttpEncPrivateKey, err := identity.LoadPrivateKey(baseDir + "enc.key")
	if err != nil {
		sendError(conn, err)
		return
	}

	decryptedIDBytes, err := identity.DecryptWithPrivateKeyBase64(reg.ID, ttpEncPrivateKey)
	if err != nil {
		sendError(conn, err)
		return
	}

	decryptedID := string(decryptedIDBytes)

	userAuthPublicKey, err := identity.ParsePublicKeyFromBase64(reg.AuthPublicKey)
	if err != nil {
		sendError(conn, err)
		return
	}

	ttpAuthPrivateKey, err := identity.LoadPrivateKey(baseDir + "auth.key")
	if err != nil {
		sendError(conn, err)
		return
	}

	certificateBase64, err := identity.CreateCertificateBase64(
		decryptedID,
		userAuthPublicKey,
		ttpAuthPrivateKey,
	)
	if err != nil {
		sendError(conn, err)
		return
	}

	response := protocol.Message{
		Type: "CERTIFICATE",
		Body: certificateBase64,
	}

	encoded, err := protocol.Encode(response)
	if err != nil {
		log.Println(err)
		return
	}

	if err = transport.Send(conn, encoded); err != nil {
		log.Println(err)
		return
	}
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

func registrationDataFromBody(body any) (protocol.RegistrationData, error) {
	raw, err := json.Marshal(body)
	if err != nil {
		return protocol.RegistrationData{}, err
	}

	var reg protocol.RegistrationData
	if err = json.Unmarshal(raw, &reg); err != nil {
		return protocol.RegistrationData{}, err
	}

	if reg.ID == "" {
		return protocol.RegistrationData{}, fmt.Errorf("empty registration ID")
	}

	if reg.AuthPublicKey == "" {
		return protocol.RegistrationData{}, fmt.Errorf("empty auth public key")
	}

	if reg.EncPublicKey == "" {
		return protocol.RegistrationData{}, fmt.Errorf("empty enc public key")
	}

	return reg, nil
}

func handleUnknown(conn net.Conn, msg protocol.Message) {
	log.Println("Unknown message type: " + msg.Type)
	response := protocol.Message{
		Type: "Error",
		Body: "Unknown message type: " + msg.Type}

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

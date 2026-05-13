package main

import (
	"crypto/rsa"
	"fmt"
	"log"
	"net"
	"os"
	"scs/internal/identity"
	"scs/internal/protocol"
	"scs/internal/transport"
	"scs/internal/ttp"
)

const baseDir = "/tmp/scs/client/"

func main() {
	ttpPublicKey := initToTtp()
	identity.EnsureIdentity(baseDir)
	registerToTtp(ttpPublicKey)
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

func initToTtp() *rsa.PublicKey {
	addr := os.Getenv("TTP_ADDR")
	if addr == "" {
		addr = "localhost:8081"
	}

	ttpPublicKey, err := ttp.Init(addr)
	if err != nil {
		log.Fatal(err)
	}

	return ttpPublicKey
}

func registerToTtp(ttpPublicKey *rsa.PublicKey) {
	data := identity.LoadRegistrationData(baseDir)

	encryptedID, err := identity.EncryptWithPublicKeyBase64([]byte(data.ID), ttpPublicKey)
	if err != nil {
		log.Fatal(err)
	}

	data.ID = encryptedID

	addr := os.Getenv("TTP_ADDR")
	if addr == "" {
		addr = "localhost:8081"
	}

	err = ttp.Register(addr, data)
	if err != nil {
		log.Fatal(err)
	}
}

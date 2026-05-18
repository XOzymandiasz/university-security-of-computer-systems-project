package main

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
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
	data := identity.LoadRegistrationData(baseDir)

	registerToTtp(ttpPublicKey, data)

	startApi()
}

func startApi() {
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/api/message", handleMessage)

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT must be set")
	}

	addr := ":" + port

	fmt.Println("client API listening on", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("client is healthy"))
	if err != nil {
		return
	}
}

type MessageResponse struct {
	Body any `json:"body"`
}

func handleMessage(w http.ResponseWriter, r *http.Request) {
	log.Println("HTTP /api/message hit:", r.Method)
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	responseBody, err := sendToServer()
	if err != nil {
		http.Error(w, "server error: "+err.Error(), http.StatusBadGateway)
		return
	}

	text, ok := responseBody.(string)
	if !ok {
		http.Error(w, fmt.Sprintf("invalid server response body type: %T", responseBody), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(MessageResponse{
		Body: text,
	})
	if err != nil {
		log.Println(err)
	}
}

func sendToServer() (any, error) {
	serverAddr := os.Getenv("SERVER_ADDR")
	if serverAddr == "" {
		return nil, fmt.Errorf("SERVER_ADDR env variable not set")
	}

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, err
	}
	defer func(conn net.Conn) {
		err = conn.Close()
		if err != nil {

		}
	}(conn)

	msg := protocol.Message{
		Type: "READ_MESSAGE",
		Body: nil,
	}

	var encoded []byte
	encoded, err = protocol.Encode(msg)
	if err != nil {
		return nil, err
	}

	if err = transport.Send(conn, encoded); err != nil {
		return nil, err
	}

	var responseData []byte
	responseData, err = transport.Receive(conn)
	if err != nil {
		return nil, err
	}

	var response protocol.Message
	response, err = protocol.Decode(responseData)
	if err != nil {
		return nil, err
	}

	if response.Type == "ERROR" {
		return nil, fmt.Errorf("%v", response.Body)
	}

	if response.Type != "MESSAGE" {
		return nil, fmt.Errorf("unexpected response type: %s", response.Type)
	}

	return response.Body, nil
}

func initToTtp() *rsa.PublicKey {
	addr := os.Getenv("TTP_ADDR")
	if addr == "" {
		log.Fatal("TTP_ADDR env variable not set")
	}

	ttpPublicKey, err := ttp.Init(addr)
	if err != nil {
		log.Fatal(err)
	}

	if ttpPublicKey == nil {
		log.Fatal("TTP public key is nil")
	}

	return ttpPublicKey
}

func registerToTtp(ttpPublicKey *rsa.PublicKey, data protocol.RegistrationData) {
	if ttpPublicKey == nil {
		log.Fatal("cannot register to TTP: public key is nil")
	}
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

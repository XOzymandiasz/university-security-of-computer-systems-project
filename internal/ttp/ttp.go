package ttp

import (
	"crypto/rsa"
	"fmt"
	"net"
	"scs/internal/identity"
	"scs/internal/protocol"
	"scs/internal/transport"
)

func Init(addr string) (*rsa.PublicKey, error) {
	response, err := sendMessage(addr, protocol.Message{
		Type: "TTP_INIT",
		Body: nil,
	})
	if err != nil {
		return nil, err
	}

	if response.Type != "TTP_PUBLIC_KEY" {
		return nil, fmt.Errorf("unexpected response type: %s", response.Type)
	}

	keyBase64, ok := response.Body.(string)
	if !ok {
		return nil, fmt.Errorf("invalid body type: %T", response.Body)
	}

	var key *rsa.PublicKey
	key, err = identity.ParsePublicKeyFromBase64(keyBase64)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func Register(addr string, data protocol.RegistrationData) (string, error) {
	response, err := sendMessage(addr, protocol.Message{
		Type: "REGISTER",
		Body: data,
	})
	if err != nil {
		return "", err
	}

	if response.Type == "ERROR" {
		body, ok := response.Body.(string)
		if !ok {
			return "", fmt.Errorf("TTP returned ERROR with invalid body type: %T", response.Body)
		}

		return "", fmt.Errorf("TTP error: %s", body)
	}

	if response.Type != "CERTIFICATE" {
		return "", fmt.Errorf("unexpected response type: %s", response.Type)
	}

	certificateBase64, ok := response.Body.(string)
	if !ok {
		return "", fmt.Errorf("invalid certificate body type: %T", response.Body)
	}

	if certificateBase64 == "" {
		return "", fmt.Errorf("empty certificate")
	}

	return certificateBase64, nil
}

func sendMessage(addr string, msg protocol.Message) (protocol.Message, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return protocol.Message{}, err
	}
	defer func(conn net.Conn) {
		err = conn.Close()
		if err != nil {
			return
		}
	}(conn)

	var encoded []byte
	encoded, err = protocol.Encode(msg)
	if err != nil {
		return protocol.Message{}, err
	}

	if err = transport.Send(conn, encoded); err != nil {
		return protocol.Message{}, err
	}

	var responseData []byte
	responseData, err = transport.Receive(conn)
	if err != nil {
		return protocol.Message{}, err
	}

	var response protocol.Message
	response, err = protocol.Decode(responseData)
	if err != nil {
		return protocol.Message{}, err
	}

	return response, nil
}

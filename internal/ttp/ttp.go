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
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	defer func(conn net.Conn) {
		err = conn.Close()
		if err != nil {
			return
		}
	}(conn)

	msg := protocol.Message{
		Type: "TTP_INIT",
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

func Register(addr string, data protocol.RegistrationData) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer func(conn net.Conn) {
		err = conn.Close()
		if err != nil {

		}
	}(conn)

	msg := protocol.Message{
		Type: "REGISTER",
		Body: data,
	}

	var encoded []byte
	encoded, err = protocol.Encode(msg)
	if err != nil {
		return err
	}

	err = transport.Send(conn, encoded)
	if err != nil {
		return err
	}

	var responseData []byte
	responseData, err = transport.Receive(conn)
	if err != nil {
		return err
	}

	_, err = protocol.Decode(responseData)
	if err != nil {
		return err
	}

	return nil
}

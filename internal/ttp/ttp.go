package ttp

import (
	"crypto/rsa"
	"fmt"
	"net"
	"scs/internal/protocol"
	"scs/internal/transport"
)

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
		Type: "Register",
		Body: data,
	}

	encoded, err := protocol.Encode(msg)
	if err != nil {
		return err
	}

	err = transport.Send(conn, encoded)
	if err != nil {
		return err
	}

	responseData, err := transport.Receive(conn)
	if err != nil {
		return err
	}

	response, err := protocol.Decode(responseData)
	if err != nil {
		return err
	}

	fmt.Println(response.Body)
	return nil
}

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

	encoded, err := protocol.Encode(msg)
	if err != nil {
		return nil, err
	}

	if err = transport.Send(conn, encoded); err != nil {
		return nil, err
	}

	responseData, err := transport.Receive(conn)
	if err != nil {
		return nil, err
	}

	response, err := protocol.Decode(responseData)
	if err != nil {
		return nil, err
	}

	if response.Type != "TTP_PUBLIC_KEY" {
		return nil, fmt.Errorf("unexpected response type: %s", response.Type)
	}

	key, ok := response.Body.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("invalid body type: %T", response.Body)
	}

	return key, nil
}

package ttp

import (
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

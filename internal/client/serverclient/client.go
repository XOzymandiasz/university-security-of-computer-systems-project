package serverclient

import (
	"fmt"
	"net"
	"scs/internal/protocol"
	"scs/internal/transport"
)

type Client struct {
	addr string
}

func New(addr string) *Client {
	return &Client{addr: addr}
}

func (c *Client) ReadMessage() (string, error) {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return "", err
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
		return "", err
	}

	if err = transport.Send(conn, encoded); err != nil {
		return "", err
	}

	var responseData []byte
	responseData, err = transport.Receive(conn)
	if err != nil {
		return "", err
	}

	var response protocol.Message
	response, err = protocol.Decode(responseData)
	if err != nil {
		return "", err
	}

	if response.Type == "ERROR" {
		return "", fmt.Errorf("%v", response.Body)
	}

	if response.Type != "MESSAGE" {
		return "", fmt.Errorf("unexpected response type: %s", response.Type)
	}

	text, ok := response.Body.(string)
	if !ok {
		return "", fmt.Errorf("invalid server response body type: %T", response.Body)
	}

	return text, nil
}

func (c *Client) HealthCheck() string {
	return "successfully health check"
}

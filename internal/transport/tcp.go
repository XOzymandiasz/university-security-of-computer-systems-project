package transport

import (
	"encoding/binary"
	"io"
	"net"
)

func Send(conn net.Conn, data []byte) error {
	length := uint32(len(data))

	if err := binary.Write(conn, binary.BigEndian, length); err != nil {
		return err
	}

	_, err := conn.Write(data)

	return err
}

func Receive(conn net.Conn) ([]byte, error) {
	var length uint32

	if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	data := make([]byte, length)

	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, err
	}

	return data, nil
}

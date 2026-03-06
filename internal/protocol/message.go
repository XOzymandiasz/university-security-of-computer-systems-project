package protocol

import "encoding/json"

type Message struct {
	Type string `json:"type"`
	Body string `json:"body"`
}

func Encode(msg Message) ([]byte, error) {
	return json.Marshal(msg)
}

func Decode(data []byte) (Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	return msg, err
}

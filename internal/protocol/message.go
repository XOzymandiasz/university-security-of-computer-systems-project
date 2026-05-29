package protocol

type Message struct {
	Type string `json:"type"`
	Body any    `json:"body"`
}

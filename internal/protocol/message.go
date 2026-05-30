package protocol

type MessageRequest struct {
	Body string `json:"body"`
}

type MessageResponse struct {
	Body string `json:"body"`
}

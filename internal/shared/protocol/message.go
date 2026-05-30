package protocol

type MessageRequest struct {
	EncryptedBody string `json:"encrypted_body"`
}

type MessageResponse struct {
	EncryptedBody string `json:"encrypted_body"`
}

type UIMessageRequest struct {
	Body string `json:"body"`
}

type UIMessageResponse struct {
	Body string `json:"body"`
}

package httpapi

type MessageResponse struct {
	Body string `json:"body"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

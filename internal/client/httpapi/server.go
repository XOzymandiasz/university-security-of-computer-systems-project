package httpapi

import "net/http"

type ReadMessageUseCase interface {
	ReadMessage(msg string) (string, error)
}

type AuthenticateUseCase interface {
	Authenticate() error
}

type Server struct {
	readMessage  ReadMessageUseCase
	authenticate AuthenticateUseCase
	baseDir      string
}

func New(
	readMessage ReadMessageUseCase,
	authenticate AuthenticateUseCase,
	baseDir string,
) *Server {
	return &Server{
		readMessage:  readMessage,
		authenticate: authenticate,
		baseDir:      baseDir,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/message", s.handleMessage)
	mux.HandleFunc("/api/authenticate", s.handleAuthenticate)

	return mux
}

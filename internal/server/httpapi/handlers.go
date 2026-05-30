package httpapi

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"scs/internal/identity"
	"scs/internal/protocol"
	"strings"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("healthy"))
}

func (s *Server) handleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request protocol.MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	defer func() {
		_ = r.Body.Close()
	}()

	if request.Body == "" {
		http.Error(w, "message body is required", http.StatusBadRequest)
		return
	}

	response := protocol.MessageResponse{
		Body: strings.ToUpper(request.Body),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "encode response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleAuthenticate(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var clientReq protocol.ClientAuthenticateRequest
	if err := json.NewDecoder(request.Body).Decode(&clientReq); err != nil {
		http.Error(writer, "invalid request body", http.StatusBadRequest)
		return
	}

	serverData := identity.LoadRegistrationData(s.baseDir)

	serverCertificate, err := identity.LoadCertificate(s.baseDir)
	if err != nil {
		http.Error(writer, "load server certificate: "+err.Error(), http.StatusInternalServerError)
		return
	}

	ttpReq := protocol.AuthenticateRequest{
		ServerID:               serverData.EncryptedID,
		ServerCertificate:      serverCertificate,
		ClientEncryptedPayload: clientReq.ClientEncryptedPayload,
	}

	ttpResp, err := s.ttpClient.Authenticate(ttpReq)
	if err != nil {
		http.Error(writer, "ttp authenticate: "+err.Error(), http.StatusUnauthorized)
		return
	}

	serverEncPrivateKey, err := identity.LoadPrivateKey(filepath.Join(s.baseDir, "enc.key"))
	if err != nil {
		http.Error(writer, "load server enc private key: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sessionKey, err := identity.DecryptWithPrivateKeyBase64(
		ttpResp.EncryptedSessionKeyForServer,
		serverEncPrivateKey,
	)
	if err != nil {
		http.Error(writer, "decrypt server session key: "+err.Error(), http.StatusUnauthorized)
		return
	}

	if err = identity.SaveSessionKey(s.baseDir, sessionKey); err != nil {
		http.Error(writer, "save server session key: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := protocol.ClientAuthenticateResponse{
		OK:                           true,
		EncryptedSessionKeyForClient: ttpResp.EncryptedSessionKeyForClient,
		Message:                      "authenticated",
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(writer).Encode(resp)
}

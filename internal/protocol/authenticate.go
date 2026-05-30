package protocol

type AuthenticateRequest struct {
	ServerID               string `json:"server_id"`
	ServerCertificate      string `json:"server_certificate"`
	ClientEncryptedPayload string `json:"client_encrypted_payload"`
}

type AuthenticateClientPayload struct {
	ClientID          string `json:"client_id"`
	ClientCertificate string `json:"client_certificate"`
}

type AuthenticateResponse struct {
	OK                           bool   `json:"ok"`
	EncryptedSessionKeyForClient string `json:"encrypted_session_key_for_client"`
	EncryptedSessionKeyForServer string `json:"encrypted_session_key_for_server"`
	Message                      string `json:"message"`
}

type ClientAuthenticateRequest struct {
	ClientEncryptedPayload string `json:"client_encrypted_payload"`
}

type ClientAuthenticateResponse struct {
	OK                           bool   `json:"ok"`
	EncryptedSessionKeyForClient string `json:"encrypted_session_key_for_client"`
	Message                      string `json:"message"`
}

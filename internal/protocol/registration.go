package protocol

type RegisterRequest struct {
	EncryptedID   string     `json:"encrypted_id"`
	EncPublicKey  string     `json:"enc_public_key"`
	AuthPublicKey string     `json:"auth_public_key"`
	Role          EntityRole `json:"role"`
}

type RegisterResponse struct {
	Certificate string `json:"certificate"`
}

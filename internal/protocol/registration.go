package protocol

type RegistrationData struct {
	ID            string `json:"id"`
	EncPublicKey  string `json:"enc_public_key"`
	AuthPublicKey string `json:"auth_public_key"`
}

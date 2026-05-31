package protocol

// RegisterRequest reprezentuje żądanie rejestracji aplikacji w TTP.
//
// Struktura zawiera identyfikator aplikacji, publiczny klucz szyfrujący,
// publiczny klucz uwierzytelniający oraz rolę aplikacji. Na podstawie tych
// danych TTP może utworzyć certyfikat X.509 dla klienta albo serwera.
//
// @field EncryptedID Identyfikator aplikacji przesyłany do TTP.
// @field EncPublicKey Publiczny klucz RSA używany do szyfrowania danych dla aplikacji.
// @field AuthPublicKey Publiczny klucz RSA używany do weryfikacji podpisów aplikacji.
// @field Role Rola aplikacji w protokole, czyli klient albo serwer.
type RegisterRequest struct {
	EncryptedID   string     `json:"encrypted_id"`
	EncPublicKey  string     `json:"enc_public_key"`
	AuthPublicKey string     `json:"auth_public_key"`
	Role          EntityRole `json:"role"`
}

// RegisterResponse reprezentuje odpowiedź TTP po rejestracji aplikacji.
//
// Odpowiedź zawiera certyfikat X.509 wygenerowany i podpisany przez TTP.
// Certyfikat jest później używany podczas procesu uwierzytelniania.
//
// @field Certificate Certyfikat X.509 aplikacji w formacie Base64.
type RegisterResponse struct {
	Certificate string `json:"certificate"`
}

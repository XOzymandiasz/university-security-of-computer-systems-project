package protocol

// InitResponse reprezentuje odpowiedź TTP na początkowe żądanie inicjalizacji.
//
// Struktura zawiera publiczny klucz szyfrujący TTP. Klient lub serwer używa
// tego klucza do zaszyfrowania danych przesyłanych do TTP podczas dalszych
// etapów rejestracji i uwierzytelniania.
//
// @field TTPEncPublicKey Publiczny klucz szyfrujący TTP w formacie Base64.
type InitResponse struct {
	TTPEncPublicKey string `json:"ttp_enc_public_key"`
}

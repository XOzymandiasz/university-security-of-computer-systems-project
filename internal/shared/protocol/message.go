package protocol

// MessageRequest reprezentuje zaszyfrowane żądanie wysłania wiadomości.
//
// Struktura jest używana w komunikacji klient-serwer po poprawnym
// uwierzytelnieniu stron i uzyskaniu wspólnego klucza sesyjnego AES.
//
// @field EncryptedBody Treść wiadomości zaszyfrowana kluczem sesyjnym.
type MessageRequest struct {
	EncryptedBody string `json:"encrypted_body"`
}

// MessageResponse reprezentuje zaszyfrowaną odpowiedź serwera.
//
// Odpowiedź zawiera treść zaszyfrowaną tym samym kluczem sesyjnym,
// który został wcześniej przekazany klientowi i serwerowi przez TTP.
//
// @field EncryptedBody Treść odpowiedzi zaszyfrowana kluczem sesyjnym.
type MessageResponse struct {
	EncryptedBody string `json:"encrypted_body"`
}

// UIMessageRequest reprezentuje jawne żądanie wiadomości z interfejsu użytkownika.
//
// Struktura jest używana między aplikacją webową a lokalnym API klienta.
// Zawiera niezaszyfrowaną treść wpisaną przez użytkownika, która następnie
// jest szyfrowana przed wysłaniem do serwera.
//
// @field Body Jawna treść wiadomości wpisana w interfejsie użytkownika.
type UIMessageRequest struct {
	Body string `json:"body"`
}

// UIMessageResponse reprezentuje jawną odpowiedź zwracaną do interfejsu użytkownika.
//
// Struktura zawiera odszyfrowaną treść odpowiedzi serwera, która może zostać
// wyświetlona użytkownikowi w aplikacji webowej.
//
// @field Body Jawna treść odpowiedzi po odszyfrowaniu.
type UIMessageResponse struct {
	Body string `json:"body"`
}

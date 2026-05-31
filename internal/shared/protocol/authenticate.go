// Package protocol definiuje struktury komunikatów używanych podczas
// uwierzytelniania klienta i serwera z udziałem TTP.
package protocol

// AuthenticateRequest reprezentuje żądanie uwierzytelnienia wysyłane do TTP.
//
// Struktura zawiera dane serwera oraz zaszyfrowany pakiet klienta.
// TTP wykorzystuje te informacje do weryfikacji certyfikatu serwera,
// podpisu serwera oraz tożsamości klienta.
//
// @field ServerID Identyfikator serwera biorącego udział w uwierzytelnianiu.
// @field ServerCertificate Certyfikat X.509 serwera w formacie Base64.
// @field ServerSignature Podpis cyfrowy serwera potwierdzający jego tożsamość.
// @field ClientEncryptedPayload Zaszyfrowany pakiet danych klienta.
type AuthenticateRequest struct {
	ServerID               string `json:"server_id"`
	ServerCertificate      string `json:"server_certificate"`
	ServerSignature        string `json:"server_signature"`
	ClientEncryptedPayload string `json:"client_encrypted_payload"`
}

// AuthenticateClientPayload reprezentuje dane klienta przekazywane do TTP.
//
// Pakiet zawiera identyfikator klienta, jego certyfikat oraz podpis cyfrowy.
// Dane te są szyfrowane przed wysłaniem, aby mogły zostać odczytane wyłącznie
// przez usługę TTP.
//
// @field ClientID Identyfikator klienta.
// @field ClientCertificate Certyfikat X.509 klienta w formacie Base64.
// @field ClientSignature Podpis cyfrowy klienta potwierdzający jego tożsamość.
type AuthenticateClientPayload struct {
	ClientID          string `json:"client_id"`
	ClientCertificate string `json:"client_certificate"`
	ClientSignature   string `json:"client_signature"`
}

// AuthenticateResponse reprezentuje odpowiedź TTP po procesie uwierzytelniania.
//
// W przypadku powodzenia odpowiedź zawiera klucz sesyjny zaszyfrowany osobno
// dla klienta oraz serwera. Dzięki temu obie strony mogą uzyskać ten sam
// klucz AES, ale tylko przy użyciu własnych kluczy prywatnych.
//
// @field OK Informuje, czy uwierzytelnianie zakończyło się powodzeniem.
// @field EncryptedSessionKeyForClient Klucz sesyjny zaszyfrowany dla klienta.
// @field EncryptedSessionKeyForServer Klucz sesyjny zaszyfrowany dla serwera.
// @field Message Komunikat opisujący wynik operacji.
type AuthenticateResponse struct {
	OK                           bool   `json:"ok"`
	EncryptedSessionKeyForClient string `json:"encrypted_session_key_for_client"`
	EncryptedSessionKeyForServer string `json:"encrypted_session_key_for_server"`
	Message                      string `json:"message"`
}

// ClientAuthenticateRequest reprezentuje żądanie uwierzytelnienia klienta
// przekazywane do serwera.
//
// Serwer otrzymuje zaszyfrowany pakiet klienta i przekazuje go dalej do TTP
// razem ze swoimi danymi uwierzytelniającymi.
//
// @field ClientEncryptedPayload Zaszyfrowany pakiet danych klienta.
type ClientAuthenticateRequest struct {
	ClientEncryptedPayload string `json:"client_encrypted_payload"`
}

// ClientAuthenticateResponse reprezentuje odpowiedź serwera po uwierzytelnieniu klienta.
//
// Odpowiedź zawiera wynik operacji oraz klucz sesyjny przeznaczony dla klienta,
// jeśli TTP pozytywnie zweryfikowało klienta i serwer.
//
// @field OK Informuje, czy proces uwierzytelniania zakończył się powodzeniem.
// @field EncryptedSessionKeyForClient Klucz sesyjny zaszyfrowany kluczem publicznym klienta.
// @field Message Komunikat opisujący wynik operacji.
type ClientAuthenticateResponse struct {
	OK                           bool   `json:"ok"`
	EncryptedSessionKeyForClient string `json:"encrypted_session_key_for_client"`
	Message                      string `json:"message"`
}

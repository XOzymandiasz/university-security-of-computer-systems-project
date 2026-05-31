package usecase

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"path/filepath"

	"scs/internal/shared/identity"
	"scs/internal/shared/protocol"
)

// ServerAuthenticator definiuje operację uwierzytelnienia klienta przez serwer.
//
// Interfejs pozwala warstwie przypadku użycia wywołać serwer bez zależności
// od konkretnej implementacji klienta HTTP.
type ServerAuthenticator interface {
	Authenticate(req protocol.ClientAuthenticateRequest) (protocol.ClientAuthenticateResponse, error)
}

// TTPInitializer definiuje operację pobrania publicznego klucza TTP.
//
// Interfejs jest używany przed zaszyfrowaniem pakietu klienta, aby aplikacja
// mogła pozyskać aktualny publiczny klucz szyfrujący Trusted Third Party.
type TTPInitializer interface {
	Init() (*rsa.PublicKey, error)
}

// Authenticate realizuje przypadek użycia uwierzytelniania klienta.
//
// Struktura przechowuje katalog lokalnej tożsamości klienta oraz zależności
// potrzebne do komunikacji z serwerem i TTP.
type Authenticate struct {
	baseDir      string
	serverClient ServerAuthenticator
	ttpClient    TTPInitializer
}

// NewAuthenticate tworzy nowy przypadek użycia uwierzytelniania klienta.
//
// Funkcja przyjmuje katalog bazowy lokalnej tożsamości oraz klientów
// komunikacyjnych używanych do kontaktu z serwerem i TTP.
//
// @param baseDir Katalog bazowy lokalnej tożsamości klienta.
// @param serverClient Klient odpowiedzialny za wysłanie żądania uwierzytelnienia do serwera.
// @param ttpClient Klient odpowiedzialny za pobranie publicznego klucza TTP.
// @return Wskaźnik do nowej instancji Authenticate.
func NewAuthenticate(
	baseDir string,
	serverClient ServerAuthenticator,
	ttpClient TTPInitializer,
) *Authenticate {
	return &Authenticate{
		baseDir:      baseDir,
		serverClient: serverClient,
		ttpClient:    ttpClient,
	}
}

// Authenticate wykonuje proces uwierzytelniania klienta z udziałem TTP.
//
// Funkcja pobiera publiczny klucz TTP, odczytuje lokalne dane klienta,
// ładuje certyfikat i klucz prywatny klienta, tworzy podpis identyfikatora,
// szyfruje pakiet uwierzytelniający dla TTP, a następnie przekazuje go
// do serwera. Po pozytywnej odpowiedzi odszyfrowuje klucz sesyjny AES
// i zapisuje go lokalnie do dalszej komunikacji z serwerem.
//
// @return Błąd procesu uwierzytelniania lub nil w przypadku powodzenia.
func (a *Authenticate) Authenticate() error {
	ttpPublicKey, err := a.ttpClient.Init()
	if err != nil {
		return fmt.Errorf("ttp init: %w", err)
	}

	clientData, err := identity.LoadRegistrationData(a.baseDir)

	if err != nil {
		return fmt.Errorf("load registration data: %w", err)
	}

	var clientCertificate string
	clientCertificate, err = identity.LoadCertificate(a.baseDir)
	if err != nil {
		return fmt.Errorf("load client certificate: %w", err)
	}

	var clientAuthPrivateKey *rsa.PrivateKey
	clientAuthPrivateKey, err = identity.LoadPrivateKey(filepath.Join(a.baseDir, "auth.key"))
	if err != nil {
		return fmt.Errorf("load client auth private key: %w", err)
	}

	var clientSignature string
	clientSignature, err = identity.SignBase64([]byte(clientData.EncryptedID), clientAuthPrivateKey)
	if err != nil {
		return fmt.Errorf("sign client id: %w", err)
	}

	payload := protocol.AuthenticateClientPayload{
		ClientID:          clientData.EncryptedID,
		ClientCertificate: clientCertificate,
		ClientSignature:   clientSignature,
	}

	var payloadBytes []byte
	payloadBytes, err = json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal client auth payload: %w", err)
	}

	var encryptedPayload string
	encryptedPayload, err = identity.EncryptLargePayloadWithPublicKeyBase64(payloadBytes, ttpPublicKey)
	if err != nil {
		return fmt.Errorf("encrypt client auth payload: %w", err)
	}

	var serverResp protocol.ClientAuthenticateResponse
	serverResp, err = a.serverClient.Authenticate(protocol.ClientAuthenticateRequest{
		ClientEncryptedPayload: encryptedPayload,
	})
	if err != nil {
		return fmt.Errorf("server authenticate: %w", err)
	}

	var clientEncPrivateKey *rsa.PrivateKey
	clientEncPrivateKey, err = identity.LoadPrivateKey(filepath.Join(a.baseDir, "enc.key"))
	if err != nil {
		return fmt.Errorf("load client enc private key: %w", err)
	}

	var sessionKey []byte
	sessionKey, err = identity.DecryptWithPrivateKeyBase64(
		serverResp.EncryptedSessionKeyForClient,
		clientEncPrivateKey,
	)
	if err != nil {
		return fmt.Errorf("decrypt client session key: %w", err)
	}

	if err = identity.SaveSessionKey(a.baseDir, sessionKey); err != nil {
		return fmt.Errorf("save client session key: %w", err)
	}

	return nil
}

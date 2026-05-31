// Package client zawiera klienta HTTP używanego do komunikacji z TTP
// oraz serwerem aplikacyjnym w ramach protokołu uwierzytelniania.
package client

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"scs/internal/shared/identity"
	"scs/internal/shared/protocol"
)

// Client reprezentuje klienta HTTP komunikującego się z wybraną usługą.
//
// Struktura przechowuje adres zdalnej usługi oraz katalog bazowy lokalnej
// tożsamości aplikacji, w którym znajdują się certyfikaty, klucze i klucz sesyjny.
type Client struct {
	addr    string
	baseDir string
}

// New tworzy nową instancję klienta HTTP.
//
// Funkcja zapisuje adres usługi oraz katalog bazowy tożsamości aplikacji.
// Utworzony klient może następnie wykonywać żądania inicjalizacji,
// rejestracji, uwierzytelniania i wymiany wiadomości.
//
// @param addr Adres zdalnej usługi w formacie host:port.
// @param baseDir Katalog bazowy lokalnej tożsamości aplikacji.
// @return Wskaźnik do nowej instancji klienta.
func New(addr string, baseDir string) *Client {
	return &Client{
		addr:    addr,
		baseDir: baseDir,
	}
}

// Init pobiera publiczny klucz szyfrujący TTP.
//
// Funkcja wysyła żądanie inicjalizacyjne do TTP, odbiera publiczny klucz
// szyfrujący TTP w formacie Base64 i parsuje go do struktury RSA.
// Klucz ten jest używany do szyfrowania danych przesyłanych do TTP.
//
// @return Publiczny klucz RSA usługi TTP lub błąd żądania.
func (c *Client) Init() (*rsa.PublicKey, error) {
	resp, err := http.Get(c.url("/api/init"))
	if err != nil {
		return nil, fmt.Errorf("ttp init request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ttp init failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var response protocol.InitResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode init response: %w", err)
	}

	keyBase64 := response.TTPEncPublicKey

	var key *rsa.PublicKey
	key, err = identity.ParsePublicKeyFromBase64(keyBase64)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	return key, nil
}

// Register rejestruje klienta lub serwer w usłudze TTP.
//
// Funkcja wysyła żądanie rejestracyjne zawierające identyfikator, klucze
// publiczne oraz rolę aplikacji. W odpowiedzi TTP zwraca certyfikat X.509,
// który jest później wykorzystywany w procesie uwierzytelniania.
//
// @param req Dane rejestracyjne aplikacji.
// @return Certyfikat X.509 w formacie Base64 lub błąd rejestracji.
func (c *Client) Register(req protocol.RegisterRequest) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal register request: %w", err)
	}

	var httpReq *http.Request
	httpReq, err = http.NewRequest(
		http.MethodPost,
		c.url("/api/register"),
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("create register request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	resp, err = http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("ttp register request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ = io.ReadAll(resp.Body)
		return "", fmt.Errorf("ttp register failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var response protocol.RegisterResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode register response: %w", err)
	}

	if response.Certificate == "" {
		return "", fmt.Errorf("empty certificate in register response")
	}

	return response.Certificate, nil
}

// Authenticate wysyła żądanie uwierzytelnienia klienta do serwera.
//
// Funkcja przekazuje do serwera zaszyfrowany pakiet klienta. Serwer rozszerza
// żądanie o własne dane i przekazuje je do TTP. Po pozytywnej weryfikacji
// odpowiedź zawiera klucz sesyjny zaszyfrowany dla klienta.
//
// @param req Żądanie uwierzytelnienia klienta zawierające zaszyfrowany pakiet.
// @return Odpowiedź uwierzytelniania lub błąd w przypadku odrzucenia.
func (c *Client) Authenticate(req protocol.ClientAuthenticateRequest) (protocol.ClientAuthenticateResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return protocol.ClientAuthenticateResponse{}, fmt.Errorf("marshal client authenticate request: %w", err)
	}

	var httpReq *http.Request
	httpReq, err = http.NewRequest(
		http.MethodPost,
		c.url("/api/authenticate"),
		bytes.NewReader(body),
	)
	if err != nil {
		return protocol.ClientAuthenticateResponse{}, fmt.Errorf("create client authenticate request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	resp, err = http.DefaultClient.Do(httpReq)
	if err != nil {
		return protocol.ClientAuthenticateResponse{}, fmt.Errorf("client authenticate request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return protocol.ClientAuthenticateResponse{}, fmt.Errorf(
			"client authenticate failed: status=%d body=%s",
			resp.StatusCode,
			string(respBody),
		)
	}

	var response protocol.ClientAuthenticateResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return protocol.ClientAuthenticateResponse{}, fmt.Errorf("decode client authenticate response: %w", err)
	}

	if !response.OK {
		return protocol.ClientAuthenticateResponse{}, fmt.Errorf("authentication rejected: %s", response.Message)
	}

	return response, nil
}

// ReadMessage wysyła zaszyfrowaną wiadomość do serwera i odczytuje odpowiedź.
//
// Funkcja wczytuje lokalny klucz sesyjny AES, szyfruje treść wiadomości,
// wysyła ją do serwera, a następnie odszyfrowuje otrzymaną odpowiedź.
// Po zakończeniu wymiany usuwa lokalny klucz sesyjny, zamykając sesję.
//
// @param msg Jawna treść wiadomości wpisana przez użytkownika.
// @return Odszyfrowana odpowiedź serwera lub błąd komunikacji.
func (c *Client) ReadMessage(msg string) (string, error) {
	sessionKey, err := identity.LoadSessionKey(c.baseDir)
	if err != nil {
		return "", fmt.Errorf("client is not authenticated - missing session key: %w", err)
	}

	var encryptedBody string
	encryptedBody, err = identity.EncryptWithSessionKeyBase64([]byte(msg), sessionKey)
	if err != nil {
		return "", fmt.Errorf("encrypt message: %w", err)
	}

	req := protocol.MessageRequest{
		EncryptedBody: encryptedBody,
	}

	var body []byte
	body, err = json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal message request: %w", err)
	}

	var httpReq *http.Request
	httpReq, err = http.NewRequest(
		http.MethodPost,
		c.url("/api/message"),
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("create message request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	resp, err = http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("server read message request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ = io.ReadAll(resp.Body)
		return "", fmt.Errorf("server read message failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var response protocol.MessageResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode message response: %w", err)
	}

	var plaintext []byte
	plaintext, err = identity.DecryptWithSessionKeyBase64(response.EncryptedBody, sessionKey)
	if err != nil {
		return "", fmt.Errorf("decrypt message response: %w", err)
	}

	_ = identity.DeleteSessionKey(c.baseDir)

	return string(plaintext), nil
}

// url buduje pełny adres HTTP na podstawie ścieżki endpointu.
//
// @param path Ścieżka endpointu rozpoczynająca się od znaku slash.
// @return Pełny adres URL usługi.
func (c *Client) url(path string) string {
	return "http://" + c.addr + path
}

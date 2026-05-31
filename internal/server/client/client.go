// Package client zawiera klienta HTTP używanego przez serwer do komunikacji z TTP.
//
// Pakiet odpowiada za inicjalizację połączenia z TTP, rejestrację aplikacji
// oraz przekazywanie żądań uwierzytelnienia klient-serwer do zaufanej strony trzeciej.
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

// RegisterRequest reprezentuje uproszczone żądanie rejestracji.
//
// Struktura zawiera identyfikator oraz publiczny klucz aplikacji.
// W aktualnym protokole główne żądanie rejestracji jest definiowane
// przez protocol.RegisterRequest, dlatego ten typ może pełnić rolę pomocniczą.
//
// @field ID Identyfikator rejestrowanej aplikacji.
// @field PublicKey Publiczny klucz aplikacji.
type RegisterRequest struct {
	ID        string `json:"id"`
	PublicKey string `json:"public_key"`
}

// Client reprezentuje klienta HTTP komunikującego się z usługą TTP.
//
// Struktura przechowuje adres TTP i udostępnia metody pozwalające pobrać
// publiczny klucz TTP, zarejestrować aplikację oraz wykonać uwierzytelnianie.
type Client struct {
	addr string
}

// New tworzy nowego klienta HTTP do komunikacji z TTP.
//
// @param addr Adres usługi TTP w formacie host:port.
// @return Wskaźnik do nowej instancji Client.
func New(addr string) *Client {
	return &Client{addr: addr}
}

// Init pobiera publiczny klucz szyfrujący TTP.
//
// Funkcja wysyła żądanie inicjalizacyjne do endpointu TTP, odbiera klucz
// publiczny w formacie Base64 i parsuje go do struktury RSA.
//
// @return Publiczny klucz RSA usługi TTP lub błąd inicjalizacji.
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

// Register rejestruje aplikację w TTP.
//
// Funkcja wysyła do TTP dane rejestracyjne aplikacji, takie jak zaszyfrowany
// identyfikator, publiczne klucze RSA oraz rola aplikacji. W odpowiedzi
// otrzymuje certyfikat X.509 wygenerowany przez TTP.
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
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ttp register failed: status=%d body=%s", resp.StatusCode, string(respBody))
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

// Authenticate przekazuje do TTP żądanie uwierzytelnienia klienta i serwera.
//
// Funkcja wysyła komplet danych uwierzytelniających do TTP. Po pozytywnej
// weryfikacji TTP zwraca klucz sesyjny AES zaszyfrowany osobno dla klienta
// i serwera.
//
// @param req Żądanie uwierzytelnienia zawierające dane klienta i serwera.
// @return Odpowiedź TTP z zaszyfrowanymi kluczami sesyjnymi lub błąd.
func (c *Client) Authenticate(req protocol.AuthenticateRequest) (protocol.AuthenticateResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return protocol.AuthenticateResponse{}, fmt.Errorf("marshal ttp authenticate request: %w", err)
	}

	var httpReq *http.Request
	httpReq, err = http.NewRequest(
		http.MethodPost,
		c.url("/api/authenticate"),
		bytes.NewReader(body),
	)
	if err != nil {
		return protocol.AuthenticateResponse{}, fmt.Errorf("create ttp authenticate request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	resp, err = http.DefaultClient.Do(httpReq)
	if err != nil {
		return protocol.AuthenticateResponse{}, fmt.Errorf("ttp authenticate request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return protocol.AuthenticateResponse{}, fmt.Errorf(
			"ttp authenticate failed: status=%d body=%s",
			resp.StatusCode,
			string(respBody),
		)
	}

	var response protocol.AuthenticateResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return protocol.AuthenticateResponse{}, fmt.Errorf("decode ttp authenticate response: %w", err)
	}

	if !response.OK {
		return protocol.AuthenticateResponse{}, fmt.Errorf("authentication rejected: %s", response.Message)
	}

	if response.EncryptedSessionKeyForClient == "" {
		return protocol.AuthenticateResponse{}, fmt.Errorf("empty encrypted session key for client")
	}

	if response.EncryptedSessionKeyForServer == "" {
		return protocol.AuthenticateResponse{}, fmt.Errorf("empty encrypted session key for server")
	}

	return response, nil
}

// url buduje pełny adres HTTP dla endpointu TTP.
//
// @param path Ścieżka endpointu rozpoczynająca się od znaku slash.
// @return Pełny adres URL usługi TTP.
func (c *Client) url(path string) string {
	return "http://" + c.addr + path
}

package identity

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// EncryptWithPublicKeyBase64 szyfruje dane kluczem publicznym RSA.
//
// Funkcja wykorzystuje algorytm RSA-OAEP z SHA-256. Wynik szyfrowania
// jest kodowany do formatu Base64, aby mógł być bezpiecznie przesyłany
// w komunikatach JSON pomiędzy aplikacjami.
//
// @param data Dane jawne przeznaczone do zaszyfrowania.
// @param pub Klucz publiczny RSA odbiorcy.
// @return Zaszyfrowane dane w formacie Base64 lub błąd szyfrowania.
func EncryptWithPublicKeyBase64(data []byte, pub *rsa.PublicKey) (string, error) {
	hash := sha256.New()

	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, pub, data, nil)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptWithPrivateKeyBase64 odszyfrowuje dane kluczem prywatnym RSA.
//
// Funkcja dekoduje dane z formatu Base64, a następnie odszyfrowuje je
// algorytmem RSA-OAEP z SHA-256 przy użyciu klucza prywatnego odbiorcy.
//
// @param encoded Zaszyfrowane dane zakodowane w formacie Base64.
// @param privateKey Klucz prywatny RSA odbiorcy.
// @return Odszyfrowane dane jawne lub błąd odszyfrowywania.
func DecryptWithPrivateKeyBase64(encoded string, privateKey *rsa.PrivateKey) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	hash := sha256.New()

	var plaintext []byte
	plaintext, err = rsa.DecryptOAEP(hash, rand.Reader, privateKey, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// HybridEncryptedPayload przechowuje dane zaszyfrowane metodą hybrydową.
//
// Struktura zawiera zaszyfrowany klucz AES, nonce oraz szyfrogram.
// Jest używana przy szyfrowaniu większych danych, których nie należy
// szyfrować bezpośrednio algorytmem RSA.
type HybridEncryptedPayload struct {
	EncryptedKey string `json:"encrypted_key"`
	Nonce        string `json:"nonce"`
	Ciphertext   string `json:"ciphertext"`
}

// EncryptLargePayloadWithPublicKeyBase64 szyfruje większe dane metodą hybrydową.
//
// Funkcja generuje losowy 256-bitowy klucz AES, szyfruje dane algorytmem
// AES-GCM, a następnie szyfruje klucz AES za pomocą RSA-OAEP. Cały pakiet
// wynikowy jest serializowany do JSON i kodowany jako Base64.
//
// @param data Dane jawne przeznaczone do zaszyfrowania.
// @param pub Klucz publiczny RSA odbiorcy.
// @return Pakiet zaszyfrowany metodą hybrydową w formacie Base64 lub błąd.
func EncryptLargePayloadWithPublicKeyBase64(data []byte, pub *rsa.PublicKey) (string, error) {
	aesKey := make([]byte, 32)

	if _, err := rand.Read(aesKey); err != nil {
		return "", err
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}

	var gcm cipher.AEAD
	gcm, err = cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())

	if _, err = rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nil, nonce, data, nil)

	var encryptedKey []byte
	encryptedKey, err = rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		pub,
		aesKey,
		nil,
	)
	if err != nil {
		return "", err
	}

	payload := HybridEncryptedPayload{
		EncryptedKey: base64.StdEncoding.EncodeToString(encryptedKey),
		Nonce:        base64.StdEncoding.EncodeToString(nonce),
		Ciphertext:   base64.StdEncoding.EncodeToString(ciphertext),
	}

	var payloadBytes []byte
	payloadBytes, err = json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(payloadBytes), nil
}

// EncryptWithSessionKeyBase64 szyfruje dane wspólnym kluczem sesyjnym AES-256.
//
// Funkcja jest używana po pozytywnym uwierzytelnieniu klienta i serwera.
// Dane są szyfrowane algorytmem AES-GCM z losowym nonce, a wynik jest
// serializowany do JSON i kodowany jako Base64.
//
// @param plaintext Dane jawne przeznaczone do zaszyfrowania.
// @param sessionKey 256-bitowy klucz sesyjny AES.
// @return Zaszyfrowany pakiet danych w formacie Base64 lub błąd.
func EncryptWithSessionKeyBase64(plaintext []byte, sessionKey []byte) (string, error) {
	if len(sessionKey) != 32 {
		return "", fmt.Errorf("invalid AES-256 session key length: %d", len(sessionKey))
	}

	block, err := aes.NewCipher(sessionKey)
	if err != nil {
		return "", err
	}

	var gcm cipher.AEAD
	gcm, err = cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	payload := HybridEncryptedPayload{
		Nonce:      base64.StdEncoding.EncodeToString(nonce),
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
	}

	var payloadBytes []byte
	payloadBytes, err = json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(payloadBytes), nil
}

// DecryptWithSessionKeyBase64 odszyfrowuje dane wspólnym kluczem sesyjnym AES-256.
//
// Funkcja dekoduje pakiet Base64, odczytuje nonce oraz szyfrogram,
// a następnie odszyfrowuje dane algorytmem AES-GCM przy użyciu
// wspólnego klucza sesyjnego.
//
// @param encoded Zaszyfrowany pakiet danych w formacie Base64.
// @param sessionKey 256-bitowy klucz sesyjny AES.
// @return Odszyfrowane dane jawne lub błąd odszyfrowywania.
func DecryptWithSessionKeyBase64(encoded string, sessionKey []byte) ([]byte, error) {
	if len(sessionKey) != 32 {
		return nil, fmt.Errorf("invalid AES-256 session key length: %d", len(sessionKey))
	}

	payloadBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	var payload HybridEncryptedPayload
	if err = json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, err
	}

	var nonce []byte
	nonce, err = base64.StdEncoding.DecodeString(payload.Nonce)
	if err != nil {
		return nil, err
	}

	var ciphertext []byte
	ciphertext, err = base64.StdEncoding.DecodeString(payload.Ciphertext)
	if err != nil {
		return nil, err
	}

	var block cipher.Block
	block, err = aes.NewCipher(sessionKey)
	if err != nil {
		return nil, err
	}

	var gcm cipher.AEAD
	gcm, err = cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, nonce, ciphertext, nil)
}

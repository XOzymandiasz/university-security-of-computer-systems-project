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

func EncryptWithPublicKeyBase64(data []byte, pub *rsa.PublicKey) (string, error) {
	hash := sha256.New()

	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, pub, data, nil)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

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

type HybridEncryptedPayload struct {
	EncryptedKey string `json:"encrypted_key"`
	Nonce        string `json:"nonce"`
	Ciphertext   string `json:"ciphertext"`
}

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

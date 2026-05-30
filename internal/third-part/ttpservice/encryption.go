package ttpservice

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"scs/internal/shared/identity"
)

func DecryptLargePayloadWithPrivateKeyBase64(encoded string, privateKey *rsa.PrivateKey) ([]byte, error) {
	payloadBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	var payload identity.HybridEncryptedPayload
	if err = json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, err
	}

	var encryptedKey []byte
	encryptedKey, err = base64.StdEncoding.DecodeString(payload.EncryptedKey)
	if err != nil {
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

	var aesKey []byte
	aesKey, err = rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		privateKey,
		encryptedKey,
		nil,
	)
	if err != nil {
		return nil, err
	}

	var block cipher.Block
	block, err = aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	var gcm cipher.AEAD
	gcm, err = cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	var plaintext []byte
	plaintext, err = gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

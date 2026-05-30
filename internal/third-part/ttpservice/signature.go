package ttpservice

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
)

func VerifySignatureBase64(data []byte, signatureBase64 string, publicKey *rsa.PublicKey) error {
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return err
	}

	digest := sha256.Sum256(data)

	return rsa.VerifyPSS(
		publicKey,
		crypto.SHA256,
		digest[:],
		signature,
		nil,
	)
}

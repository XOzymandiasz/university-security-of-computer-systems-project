package ttpservice

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
)

// VerifySignatureBase64 weryfikuje podpis cyfrowy RSA-PSS zapisany w formacie Base64.
//
// Funkcja dekoduje podpis z formatu Base64, oblicza skrót SHA-256
// z przekazanych danych, a następnie sprawdza podpis przy użyciu
// publicznego klucza RSA danego podmiotu.
//
// @param data Dane, których podpis ma zostać zweryfikowany.
// @param signatureBase64 Podpis cyfrowy zakodowany w formacie Base64.
// @param publicKey Publiczny klucz RSA używany do weryfikacji podpisu.
// @return Błąd weryfikacji podpisu lub nil w przypadku poprawnego podpisu.
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

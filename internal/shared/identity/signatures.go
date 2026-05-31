package identity

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
)

// SignBase64 tworzy podpis cyfrowy danych przy użyciu klucza prywatnego RSA.
//
// Funkcja oblicza skrót SHA-256 z przekazanych danych, a następnie podpisuje
// go algorytmem RSA-PSS. Wynikowy podpis jest kodowany do formatu Base64,
// aby mógł zostać przesłany w komunikatach JSON protokołu.
//
// @param data Dane, które mają zostać podpisane.
// @param privateKey Klucz prywatny RSA używany do utworzenia podpisu.
// @return Podpis cyfrowy w formacie Base64 lub błąd podpisywania.
func SignBase64(data []byte, privateKey *rsa.PrivateKey) (string, error) {
	digest := sha256.Sum256(data)

	signature, err := rsa.SignPSS(
		rand.Reader,
		privateKey,
		crypto.SHA256,
		digest[:],
		nil,
	)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

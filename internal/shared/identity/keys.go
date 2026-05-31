package identity

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"os"
)

// ParsePublicKeyFromBase64 parsuje klucz publiczny RSA zapisany w formacie Base64.
//
// Funkcja dekoduje dane Base64, odczytuje klucz publiczny w formacie PKIX
// i sprawdza, czy otrzymany klucz jest kluczem RSA.
//
// @param encoded Klucz publiczny zakodowany w formacie Base64.
// @return Klucz publiczny RSA lub błąd parsowania.
func ParsePublicKeyFromBase64(encoded string) (*rsa.PublicKey, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	var key any
	key, err = x509.ParsePKIXPublicKey(keyBytes)
	if err != nil {
		return nil, err
	}

	pub, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("no RSA public key")
	}

	return pub, nil
}

// LoadPrivateKey wczytuje klucz prywatny RSA z pliku PEM.
//
// Funkcja odczytuje plik z dysku, dekoduje blok PEM i parsuje znajdujący się
// w nim klucz prywatny w formacie PKCS#1.
//
// @param path Ścieżka do pliku zawierającego klucz prywatny RSA.
// @return Klucz prywatny RSA lub błąd odczytu albo parsowania.
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyPEM, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return nil, errors.New("failed to decode PEM")
	}

	var privateKey *rsa.PrivateKey
	privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// loadPublicKeyBase64 wczytuje klucz publiczny RSA z pliku klucza prywatnego.
//
// Funkcja odczytuje klucz prywatny z pliku PEM, pobiera z niego część publiczną,
// serializuje ją do formatu PKIX i zwraca jako tekst Base64. Wartość ta jest
// wysyłana do TTP podczas rejestracji aplikacji.
//
// @param path Ścieżka do pliku zawierającego klucz prywatny RSA.
// @return Klucz publiczny RSA w formacie Base64 lub błąd odczytu.
func loadPublicKeyBase64(path string) (string, error) {
	keyPEM, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM")
	}

	var privateKey *rsa.PrivateKey
	privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	var publicKey []byte
	publicKey, err = x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(publicKey), nil
}

// ensureKey tworzy plik z kluczem prywatnym RSA, jeśli jeszcze nie istnieje.
//
// Funkcja sprawdza obecność pliku klucza. Jeśli plik nie istnieje, generuje
// nową parę kluczy RSA o długości 4096 bitów i zapisuje klucz prywatny
// w formacie PEM.
//
// @param path Ścieżka do pliku, w którym ma znajdować się klucz prywatny RSA.
// @return Błąd tworzenia albo zapisu klucza lub nil w przypadku powodzenia.
func ensureKey(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}

	var privateKey *rsa.PrivateKey
	privateKey, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}
	keyBytes := x509.MarshalPKCS1PrivateKey(privateKey)

	pemBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyBytes}
	var file *os.File
	file, err = os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	return pem.Encode(file, pemBlock)
}

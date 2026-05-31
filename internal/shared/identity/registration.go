package identity

import (
	"os"
	"path/filepath"
	"scs/internal/shared/protocol"
)

// LoadRegistrationData przygotowuje dane rejestracyjne aplikacji dla TTP.
//
// Funkcja odczytuje lokalny identyfikator aplikacji oraz publiczne klucze RSA
// używane do uwierzytelniania i szyfrowania. Zebrane dane są zwracane jako
// struktura RegisterRequest, która następnie może zostać wysłana do TTP
// podczas procesu rejestracji klienta albo serwera.
//
// @param baseDir Katalog bazowy zawierający pliki lokalnej tożsamości aplikacji.
// @return Struktura żądania rejestracji lub błąd odczytu danych.
func LoadRegistrationData(baseDir string) (protocol.RegisterRequest, error) {
	idBytes, err := os.ReadFile(filepath.Join(baseDir, idFileName))
	if err != nil {
		return protocol.RegisterRequest{}, err
	}

	var authPub string
	authPub, err = loadPublicKeyBase64(filepath.Join(baseDir, authFileName))
	if err != nil {
		return protocol.RegisterRequest{}, err
	}

	var encPub string
	encPub, err = loadPublicKeyBase64(filepath.Join(baseDir, encFileName))
	if err != nil {
		return protocol.RegisterRequest{}, err
	}

	return protocol.RegisterRequest{
		EncryptedID:   string(idBytes),
		AuthPublicKey: authPub,
		EncPublicKey:  encPub,
	}, nil
}

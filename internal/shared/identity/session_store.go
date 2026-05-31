package identity

import (
	"encoding/base64"
	"os"
	"path/filepath"
)

// sessionFileName określa nazwę pliku używanego do przechowywania klucza sesyjnego.
const sessionFileName = "session.key"

// SaveSessionKey zapisuje klucz sesyjny AES w lokalnym magazynie aplikacji.
//
// Funkcja tworzy katalog bazowy, jeśli jeszcze nie istnieje, koduje klucz sesyjny
// do formatu Base64 i zapisuje go w pliku. Klucz sesyjny jest używany po poprawnym
// uwierzytelnieniu klienta i serwera przez TTP.
//
// @param baseDir Katalog bazowy, w którym ma zostać zapisany klucz sesyjny.
// @param sessionKey Klucz sesyjny AES otrzymany od TTP.
// @return Błąd zapisu lub nil w przypadku powodzenia.
func SaveSessionKey(baseDir string, sessionKey []byte) error {
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(sessionKey)

	return os.WriteFile(filepath.Join(baseDir, sessionFileName), []byte(encoded), 0600)
}

// LoadSessionKey wczytuje klucz sesyjny AES z lokalnego magazynu aplikacji.
//
// Funkcja odczytuje zapisany klucz sesyjny z pliku, a następnie dekoduje go
// z formatu Base64 do postaci bajtowej używanej przez funkcje szyfrowania
// i odszyfrowywania danych.
//
// @param baseDir Katalog bazowy, w którym znajduje się plik klucza sesyjnego.
// @return Klucz sesyjny AES jako tablica bajtów lub błąd odczytu.
func LoadSessionKey(baseDir string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(baseDir, sessionFileName))
	if err != nil {
		return nil, err
	}

	return base64.StdEncoding.DecodeString(string(data))
}

// DeleteSessionKey usuwa lokalnie zapisany klucz sesyjny AES.
//
// Funkcja usuwa plik klucza sesyjnego po zakończeniu sesji komunikacyjnej.
// Brak pliku nie jest traktowany jako błąd, ponieważ sesja mogła już zostać
// wcześniej zamknięta.
//
// @param baseDir Katalog bazowy, w którym znajduje się plik klucza sesyjnego.
// @return Błąd usuwania lub nil w przypadku powodzenia.
func DeleteSessionKey(baseDir string) error {
	path := filepath.Join(baseDir, sessionFileName)

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

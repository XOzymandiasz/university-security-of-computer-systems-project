package identity

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
)

var (
	idFileName   = "id.txt"
	authFileName = "auth.key"
	encFileName  = "enc.key"
	certFileName = "cert.key"
)

// EnsureIdentity tworzy lokalną tożsamość aplikacji, jeśli jeszcze nie istnieje.
//
// Funkcja przygotowuje katalog bazowy oraz upewnia się, że istnieją dwa klucze RSA:
// klucz uwierzytelniający, klucz szyfrujący oraz publiczny identyfikator aplikacji.
// Jest wywoływana podczas inicjalizacji klienta albo serwera przed rejestracją w TTP.
//
// @param baseDir Katalog bazowy, w którym przechowywane są pliki tożsamości.
// @return Błąd inicjalizacji tożsamości lub nil w przypadku powodzenia.
func EnsureIdentity(baseDir string) error {
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return err
	}

	if err := ensureKey(filepath.Join(baseDir, authFileName)); err != nil {
		return err
	}

	if err := ensureKey(filepath.Join(baseDir, encFileName)); err != nil {
		return err
	}

	if err := ensureId(filepath.Join(baseDir, idFileName)); err != nil {
		return err
	}

	return nil
}

// ensureId tworzy publiczny identyfikator aplikacji, jeśli plik jeszcze nie istnieje.
//
// Identyfikator jest generowany na podstawie losowych bajtów z kryptograficznie
// bezpiecznego generatora, a następnie skracany funkcją SHA-256 i zapisywany
// w postaci szesnastkowej.
//
// @param path Ścieżka do pliku, w którym ma znajdować się identyfikator.
// @return Błąd zapisu lub nil w przypadku powodzenia.
func ensureId(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}

	randomBytes := make([]byte, 32)
	if _, err = rand.Read(randomBytes); err != nil {
		return err
	}

	sum := sha256.Sum256(randomBytes)
	id := hex.EncodeToString(sum[:])

	return os.WriteFile(path, []byte(id), 0600)
}

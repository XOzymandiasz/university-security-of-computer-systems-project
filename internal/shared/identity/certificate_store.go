//@file certificate_store.go
///@brief Obsługa zapisu i odczytu certyfikatu X.509 z lokalnego magazynu.
//
///Plik zawiera funkcje pomocnicze wykorzystywane przez aplikacje klienta
///i serwera do przechowywania certyfikatu otrzymanego od TTP.

package identity

import (
	"os"
	"path/filepath"
)

// LoadCertificate wczytuje certyfikat z lokalnego magazynu aplikacji.
//
// Funkcja odczytuje certyfikat zapisany w formacie Base64 z katalogu
// wskazanego przez parametr baseDir. Certyfikat jest później używany
// podczas uwierzytelniania aplikacji z udziałem TTP.
//
// @param baseDir Katalog bazowy, w którym znajduje się plik certyfikatu.
// @return Certyfikat X.509 w formacie Base64 lub błąd odczytu.
func LoadCertificate(baseDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(baseDir, certFileName))
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// SaveCertificate zapisuje certyfikat w lokalnym magazynie aplikacji.
//
// Funkcja tworzy katalog magazynu, jeśli jeszcze nie istnieje, a następnie
// zapisuje certyfikat X.509 otrzymany z TTP w formacie Base64.
//
// @param baseDir Katalog bazowy, w którym ma zostać zapisany certyfikat.
// @param certificateBase64 Certyfikat X.509 zakodowany w formacie Base64.
// @return Błąd zapisu lub nil w przypadku powodzenia.
func SaveCertificate(baseDir string, certificateBase64 string) error {
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return err
	}

	path := filepath.Join(baseDir, certFileName)

	return os.WriteFile(path, []byte(certificateBase64), 0600)
}

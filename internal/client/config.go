package client

import (
	"fmt"
	"os"
)

// defaultBaseDir określa domyślny katalog przechowywania lokalnych danych klienta.
//
// W katalogu tym zapisywane są między innymi identyfikator klienta,
// klucze RSA, certyfikat X.509 oraz klucz sesyjny AES.
const defaultBaseDir = "/tmp/scs/client"

// Config przechowuje konfigurację aplikacji klienta.
//
// Struktura zawiera katalog bazowy lokalnej tożsamości, port lokalnego API,
// adres usługi TTP oraz adres serwera aplikacyjnego.
type Config struct {
	BaseDir    string
	Port       string
	TTPAddr    string
	ServerAddr string
}

// ConfigFromEnv wczytuje konfigurację klienta ze zmiennych środowiskowych.
//
// Funkcja wymaga ustawienia zmiennych PORT, TTP_ADDR oraz SERVER_ADDR.
// Na ich podstawie budowana jest konfiguracja aplikacji klienta.
//
// @return Konfiguracja klienta lub błąd brakującej zmiennej środowiskowej.
func ConfigFromEnv() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		return Config{}, fmt.Errorf("environment variable PORT not set")
	}

	ttpAddr := os.Getenv("TTP_ADDR")
	if ttpAddr == "" {
		return Config{}, fmt.Errorf("environment variable TTP_ADDR not set")
	}

	serverAddr := os.Getenv("SERVER_ADDR")
	if serverAddr == "" {
		return Config{}, fmt.Errorf("environment variable SERVER_ADDR not set")
	}

	return Config{
		BaseDir:    defaultBaseDir,
		Port:       port,
		TTPAddr:    ttpAddr,
		ServerAddr: serverAddr,
	}, nil
}

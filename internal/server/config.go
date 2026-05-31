package server

import (
	"fmt"
	"os"
)

// defaultBaseDir określa domyślny katalog przechowywania lokalnych danych serwera.
//
// W katalogu tym zapisywane są między innymi identyfikator serwera,
// klucze RSA, certyfikat X.509 oraz klucz sesyjny AES.
const defaultBaseDir = "/tmp/scs/server"

// messagePath określa domyślną ścieżkę zasobu wiadomości obsługiwanego przez serwer.
const messagePath = "/app/message"

// Config przechowuje konfigurację aplikacji serwera.
//
// Struktura zawiera katalog bazowy lokalnej tożsamości, ścieżkę wiadomości,
// port lokalnego API oraz adres usługi TTP.
type Config struct {
	BaseDir     string
	MessagePath string
	Port        string
	TTPAddr     string
}

// ConfigFromEnv wczytuje konfigurację serwera ze zmiennych środowiskowych.
//
// Funkcja wymaga ustawienia zmiennych PORT oraz TTP_ADDR. Na ich podstawie
// budowana jest konfiguracja aplikacji serwera.
//
// @return Konfiguracja serwera lub błąd brakującej zmiennej środowiskowej.
func ConfigFromEnv() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		return Config{}, fmt.Errorf("environment variable PORT not set")
	}

	ttpAddr := os.Getenv("TTP_ADDR")
	if ttpAddr == "" {
		return Config{}, fmt.Errorf("environment variable TTP_ADDR not set")
	}

	return Config{
		BaseDir:     defaultBaseDir,
		MessagePath: messagePath,
		Port:        port,
		TTPAddr:     ttpAddr,
	}, nil
}

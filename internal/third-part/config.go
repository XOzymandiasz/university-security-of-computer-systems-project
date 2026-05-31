package third_part

import (
	"fmt"
	"os"
)

// defaultBaseDir określa domyślny katalog przechowywania lokalnych danych TTP.
//
// W katalogu tym zapisywane są między innymi identyfikator TTP,
// klucze RSA oraz dane wykorzystywane podczas inicjalizacji protokołu.
const defaultBaseDir = "/tmp/scs/ttp"

// Config przechowuje konfigurację aplikacji TTP.
//
// Struktura zawiera katalog bazowy lokalnej tożsamości TTP oraz port,
// na którym uruchamiane jest API HTTP usługi Trusted Third Party.
type Config struct {
	BaseDir string
	Port    string
}

// ConfigFromEnv wczytuje konfigurację TTP ze zmiennych środowiskowych.
//
// Funkcja wymaga ustawienia zmiennej PORT. Na jej podstawie budowana jest
// konfiguracja aplikacji TTP, natomiast katalog bazowy ustawiany jest
// na wartość domyślną.
//
// @return Konfiguracja TTP lub błąd brakującej zmiennej środowiskowej.
func ConfigFromEnv() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		return Config{}, fmt.Errorf("environment variable PORT not set")
	}

	return Config{
		BaseDir: defaultBaseDir,
		Port:    port,
	}, nil
}

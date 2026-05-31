// Package third_part zawiera główną konfigurację i uruchamianie aplikacji TTP.
//
// Pakiet łączy konfigurację, lokalną tożsamość TTP, logikę usługi Trusted Third Party
// oraz API HTTP udostępniające operacje inicjalizacji, rejestracji i uwierzytelniania.
package third_part

import (
	"log"
	"net/http"

	"scs/internal/shared/identity"
	"scs/internal/third-part/httpapi"
	"scs/internal/third-part/ttpservice"
)

// App reprezentuje główną aplikację Trusted Third Party.
//
// Struktura przechowuje konfigurację TTP oraz serwer HTTP obsługujący endpointy
// protokołu zaufanej strony trzeciej.
type App struct {
	config Config
	api    *httpapi.Server
}

// NewAppFromEnv tworzy aplikację TTP na podstawie zmiennych środowiskowych.
//
// Funkcja wczytuje konfigurację, tworzy lokalną tożsamość TTP, inicjalizuje
// usługę odpowiedzialną za logikę rejestracji i uwierzytelniania oraz buduje
// API HTTP.
//
// @return Wskaźnik do skonfigurowanej aplikacji TTP lub błąd konfiguracji.
func NewAppFromEnv() (*App, error) {
	cfg, err := ConfigFromEnv()
	if err != nil {
		return nil, err
	}

	err = identity.EnsureIdentity(cfg.BaseDir)
	if err != nil {
		return nil, err
	}

	ttpService := ttpservice.New(cfg.BaseDir)

	api := httpapi.New(ttpService)
	return &App{
		config: cfg,
		api:    api,
	}, nil
}

// Run uruchamia API HTTP usługi TTP.
//
// Funkcja startuje serwer HTTP na porcie określonym w konfiguracji.
// Endpointy TTP obsługują pobranie publicznego klucza, rejestrację klienta
// lub serwera oraz proces uwierzytelniania.
//
// @return Błąd uruchomienia serwera HTTP lub nil po zakończeniu działania.
func (a *App) Run() error {
	addr := ":" + a.config.Port

	log.Println("ttp API listening on", addr)

	return http.ListenAndServe(addr, a.api.Handler())
}

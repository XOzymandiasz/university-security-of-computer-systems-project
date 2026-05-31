// Package server zawiera główną konfigurację i uruchamianie aplikacji serwera.
//
// Pakiet łączy konfigurację, klienta HTTP do komunikacji z TTP,
// lokalne API serwera oraz proces rejestracji serwera w zaufanej stronie trzeciej.
package server

import (
	"fmt"
	"log"
	"net/http"

	"scs/internal/server/client"
	"scs/internal/server/httpapi"
	"scs/internal/shared/identity"
	"scs/internal/shared/protocol"
)

// App reprezentuje główną aplikację serwera.
//
// Struktura przechowuje konfigurację serwera, API HTTP oraz klienta TTP
// używanego do rejestracji i uwierzytelniania z udziałem Trusted Third Party.
type App struct {
	config    Config
	api       *httpapi.Server
	ttpClient *client.Client
}

// NewAppFromEnv tworzy aplikację serwera na podstawie zmiennych środowiskowych.
//
// Funkcja wczytuje konfigurację, tworzy klienta HTTP do komunikacji z TTP,
// buduje lokalne API HTTP serwera i zwraca gotową instancję aplikacji.
//
// @return Wskaźnik do skonfigurowanej aplikacji serwera lub błąd konfiguracji.
func NewAppFromEnv() (*App, error) {
	cfg, err := ConfigFromEnv()
	if err != nil {
		return nil, err
	}

	ttpClient := client.New(cfg.TTPAddr)

	api := httpapi.New(cfg.MessagePath, cfg.BaseDir, ttpClient)

	return &App{
		config:    cfg,
		api:       api,
		ttpClient: ttpClient,
	}, nil
}

// Bootstrap przygotowuje serwer do udziału w protokole TTP.
//
// Funkcja pobiera publiczny klucz szyfrujący TTP, tworzy lokalną tożsamość
// serwera, szyfruje identyfikator serwera dla TTP, rejestruje serwer jako
// EntityRoleServer oraz zapisuje otrzymany certyfikat X.509.
//
// @return Błąd inicjalizacji lub nil w przypadku powodzenia.
func (a *App) Bootstrap() error {
	ttpPublicKey, err := a.ttpClient.Init()
	if err != nil {
		return fmt.Errorf("ttp init: %w", err)
	}

	err = identity.EnsureIdentity(a.config.BaseDir)
	if err != nil {
		return err
	}

	var data protocol.RegisterRequest
	data, err = identity.LoadRegistrationData(a.config.BaseDir)

	if err != nil {
		return fmt.Errorf("load registration data: %w", err)
	}

	var encryptedID string
	encryptedID, err = identity.EncryptWithPublicKeyBase64([]byte(data.EncryptedID), ttpPublicKey)
	if err != nil {
		return fmt.Errorf("encrypt id for ttp: %w", err)
	}

	var certificateBase64 string
	certificateBase64, err = a.ttpClient.Register(
		protocol.RegisterRequest{
			EncryptedID:   encryptedID,
			EncPublicKey:  data.EncPublicKey,
			AuthPublicKey: data.AuthPublicKey,
			Role:          protocol.EntityRoleServer,
		})
	if err != nil {
		return fmt.Errorf("ttp register: %w", err)
	}

	if err = identity.SaveCertificate(a.config.BaseDir, certificateBase64); err != nil {
		return fmt.Errorf("save certificate: %w", err)
	}

	return nil
}

// Run uruchamia API HTTP serwera.
//
// Funkcja startuje serwer HTTP na porcie określonym w konfiguracji.
// Endpointy serwera obsługują uwierzytelnianie klienta oraz szyfrowaną
// wymianę wiadomości po ustanowieniu klucza sesyjnego.
//
// @return Błąd uruchomienia serwera HTTP lub nil po zakończeniu działania.
func (a *App) Run() error {
	addr := ":" + a.config.Port

	log.Println("server API listening on", addr)

	return http.ListenAndServe(addr, a.api.Handler())
}

// Package client zawiera główną konfigurację i uruchamianie aplikacji klienta.
//
// Pakiet łączy konfigurację, klientów HTTP, przypadki użycia oraz lokalne API
// udostępniane interfejsowi użytkownika.
package client

import (
	"fmt"
	"log"
	"net/http"

	"scs/internal/client/client"
	"scs/internal/client/httpapi"
	"scs/internal/client/usecase"
	"scs/internal/shared/identity"
	"scs/internal/shared/protocol"
)

// App reprezentuje główną aplikację klienta.
//
// Struktura przechowuje konfigurację, lokalne API HTTP oraz klientów
// komunikacyjnych używanych do kontaktu z serwerem aplikacyjnym i TTP.
type App struct {
	config       Config
	api          *httpapi.Server
	serverClient *client.Client
	ttpClient    *client.Client
}

// NewAppFromEnv tworzy aplikację klienta na podstawie zmiennych środowiskowych.
//
// Funkcja wczytuje konfigurację, tworzy klientów HTTP dla serwera i TTP,
// buduje przypadki użycia oraz lokalne API HTTP używane przez interfejs użytkownika.
//
// @return Wskaźnik do skonfigurowanej aplikacji klienta lub błąd konfiguracji.
func NewAppFromEnv() (*App, error) {
	cfg, err := ConfigFromEnv()
	if err != nil {
		return nil, err
	}

	serverClient := client.New(cfg.ServerAddr, cfg.BaseDir)
	ttpClient := client.New(cfg.TTPAddr, cfg.BaseDir)

	readMessage := usecase.NewReadMessage(serverClient)
	authenticate := usecase.NewAuthenticate(cfg.BaseDir, serverClient, ttpClient)

	api := httpapi.New(readMessage, authenticate, cfg.BaseDir)

	return &App{
		config:       cfg,
		api:          api,
		serverClient: serverClient,
		ttpClient:    ttpClient,
	}, nil
}

// Bootstrap przygotowuje klienta do pracy w protokole TTP.
//
// Funkcja pobiera publiczny klucz szyfrujący TTP, tworzy lokalną tożsamość
// klienta, szyfruje identyfikator klienta kluczem publicznym TTP, rejestruje
// klienta w TTP oraz zapisuje otrzymany certyfikat X.509.
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
			Role:          protocol.EntityRoleClient,
		})

	if err != nil {
		return fmt.Errorf("ttp register: %w", err)
	}

	if err = identity.SaveCertificate(a.config.BaseDir, certificateBase64); err != nil {
		return fmt.Errorf("save certificate: %w", err)
	}

	return nil
}

// Run uruchamia lokalne API HTTP klienta.
//
// Funkcja startuje serwer HTTP na porcie określonym w konfiguracji.
// API jest wykorzystywane przez interfejs użytkownika do uwierzytelniania
// i wysyłania wiadomości do serwera.
//
// @return Błąd uruchomienia serwera HTTP lub nil po zakończeniu działania.
func (a *App) Run() error {
	addr := ":" + a.config.Port

	log.Println("client API listening on", addr)

	return http.ListenAndServe(addr, a.api.Handler())
}

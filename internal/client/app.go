package client

import (
	"fmt"
	"log"
	"net/http"
	"scs/internal/protocol"

	"scs/internal/client/httpapi"
	"scs/internal/client/serverclient"
	"scs/internal/client/usecase"
	"scs/internal/identity"
)

type App struct {
	config       Config
	api          *httpapi.Server
	serverClient *serverclient.Client
	ttpClient    *serverclient.Client
}

func NewAppFromEnv() (*App, error) {
	cfg, err := ConfigFromEnv()
	if err != nil {
		return nil, err
	}

	serverClient := serverclient.New(cfg.ServerAddr)
	ttpClient := serverclient.New(cfg.TTPAddr)

	readMessage := usecase.NewReadMessage(serverClient)
	authenticate := usecase.NewAuthenticate(cfg.BaseDir, serverClient, ttpClient)

	api := httpapi.New(readMessage, authenticate)

	return &App{
		config:       cfg,
		api:          api,
		serverClient: serverClient,
		ttpClient:    ttpClient,
	}, nil
}

func (a *App) Bootstrap() error {
	ttpPublicKey, err := a.ttpClient.Init()
	if err != nil {
		return fmt.Errorf("ttp init: %w", err)
	}

	identity.EnsureIdentity(a.config.BaseDir)

	data := identity.LoadRegistrationData(a.config.BaseDir)

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

func (a *App) Run() error {
	addr := ":" + a.config.Port

	log.Println("client API listening on", addr)

	return http.ListenAndServe(addr, a.api.Handler())
}

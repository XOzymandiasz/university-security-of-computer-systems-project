package server

import (
	"fmt"
	"log"
	"net/http"
	"scs/internal/identity"
	"scs/internal/protocol"
	"scs/internal/server/httpapi"
)

type App struct {
	config    Config
	api       *httpapi.Server
	ttpClient *client.Client
}

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

func (a *App) Run() error {
	addr := ":" + a.config.Port

	log.Println("server API listening on", addr)

	return http.ListenAndServe(addr, a.api.Handler())
}

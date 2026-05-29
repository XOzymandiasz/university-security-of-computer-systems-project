package server

import (
	"fmt"
	"log"
	"net/http"
	"scs/internal/identity"
	"scs/internal/server/httpapi"
	"scs/internal/server/serverclient"
)

type App struct {
	config    Config
	api       *httpapi.Server
	ttpClient *serverclient.Client
}

func NewAppFromEnv() (*App, error) {
	cfg, err := ConfigFromEnv()
	if err != nil {
		return nil, err
	}

	ttpClient := serverclient.New(cfg.TTPAddr)

	api := httpapi.New(cfg.MessagePath)

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
	encryptedID, err = identity.EncryptWithPublicKeyBase64([]byte(data.ID), ttpPublicKey)
	if err != nil {
		return fmt.Errorf("encrypt id for ttp: %w", err)
	}

	var certificateBase64 string
	certificateBase64, err = a.ttpClient.Register(encryptedID, data.AuthPublicKey)
	if err != nil {
		return fmt.Errorf("ttp register: %w", err)
	}

	_ = certificateBase64

	return nil
}

func (a *App) Run() error {
	addr := ":" + a.config.Port

	log.Println("server API listening on", addr)

	return http.ListenAndServe(addr, a.api.Handler())
}

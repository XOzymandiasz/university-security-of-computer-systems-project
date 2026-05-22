package client

import (
	"log"
	"net/http"
	"scs/internal/client/httpapi"
	"scs/internal/client/serverclient"
	"scs/internal/client/usecase"
)

type App struct {
	config Config
	api    *httpapi.Server
}

func NewAppFromEnv() (*App, error) {
	cfg, err := ConfigFromEnv()
	if err != nil {
		return nil, err
	}

	serverClient := serverclient.New(cfg.ServerAddr)
	readMessage := usecase.NewReadMessage(serverClient)
	healthCheck := usecase.NewHealthCheck(serverClient)

	api := httpapi.New(readMessage, healthCheck)
	return &App{
		config: cfg,
		api:    api,
	}, nil
}

func (a *App) Run() error {
	addr := ":" + a.config.Port

	log.Println("client API listening on", addr)

	return http.ListenAndServe(addr, a.api.Handler())
}

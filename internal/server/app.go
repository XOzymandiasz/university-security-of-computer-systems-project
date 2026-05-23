package server

import (
	"log"
	"net/http"
	"scs/internal/client/serverclient"
	"scs/internal/server/httpapi"
	"scs/internal/server/usecase"
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

	serverClient := serverclient.New(cfg.Port)
	healthCheck := usecase.NewHealthCheck(serverClient)

	api := httpapi.New(healthCheck)

	return &App{
		config: cfg,
		api:    api,
	}, nil
}

func (a *App) Run() error {
	addr := ":" + a.config.Port

	log.Println("server API listening on", addr)

	return http.ListenAndServe(addr, a.api.Handler())
}

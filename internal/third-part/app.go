package third_part

import (
	"log"
	"net/http"
	"scs/internal/identity"
	"scs/internal/third-part/httpapi"
	"scs/internal/third-part/ttpservice"
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

	identity.EnsureIdentity(cfg.BaseDir)

	ttpService := ttpservice.New(cfg.BaseDir)

	api := httpapi.New(ttpService)
	return &App{
		config: cfg,
		api:    api,
	}, nil
}

func Bootstrap() {

}

func (a *App) Run() error {
	addr := ":" + a.config.Port

	log.Println("ttp API listening on", addr)

	return http.ListenAndServe(addr, a.api.Handler())
}

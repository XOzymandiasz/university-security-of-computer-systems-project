package main

import (
	"crypto/rsa"
	"log"
	"scs/internal/client"
	"scs/internal/identity"
	"scs/internal/server"
	ttpClient "scs/internal/ttpClient"
)

func main() {
	cfg, err := server.ConfigFromEnv()
	if err != nil {
		log.Fatalln(err)
	}

	var app *client.App
	app, err = client.NewAppFromEnv()
	if err != nil {
		log.Fatalln(err)
	}

	if err = registerIdentity(cfg); err != nil {
		log.Fatalln(err)
	}

	if err = app.Run(); err != nil {
		log.Fatalln(err)
	}
}

func registerIdentity(cfg server.Config) error {
	addr, err := ttpClient.AddrFromEnv("TTP_ADDR")
	if err != nil {
		return err
	}

	var ttpPublicKey *rsa.PublicKey
	ttpPublicKey, err = ttpClient.Init(addr)
	if err != nil {
		return err
	}

	identity.EnsureIdentity(cfg.BaseDir)

	data := identity.LoadRegistrationData(cfg.BaseDir)

	_, err = ttpClient.Register(addr, ttpPublicKey, data)
	if err != nil {
		return err
	}

	return nil
}

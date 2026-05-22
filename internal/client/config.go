package client

import (
	"fmt"
	"os"
)

const defaultBaseDir = "/tmp/scs/client"

type Config struct {
	BaseDir    string
	Port       string
	TTPAddr    string
	ServerAddr string
}

func ConfigFromEnv() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		return Config{}, fmt.Errorf("environment variable PORT not set")
	}

	ttpAddr := os.Getenv("TTP_ADDR")
	if ttpAddr == "" {
		return Config{}, fmt.Errorf("environment variable TTP_ADDR not set")
	}

	serverAddr := os.Getenv("SERVER_ADDR")
	if serverAddr == "" {
		return Config{}, fmt.Errorf("environment variable SERVER_ADDR not set")
	}

	return Config{
		BaseDir:    defaultBaseDir,
		Port:       port,
		TTPAddr:    ttpAddr,
		ServerAddr: serverAddr,
	}, nil
}

package server

import (
	"fmt"
	"os"
)

const defaultBaseDir = "/tmp/scs/server"
const messagePath = "/app/message"

type Config struct {
	BaseDir     string
	MessagePath string
	Port        string
	TTPAddr     string
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

	return Config{
		BaseDir:     defaultBaseDir,
		MessagePath: messagePath,
		Port:        port,
		TTPAddr:     ttpAddr,
	}, nil
}

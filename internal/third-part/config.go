package third_part

import (
	"fmt"
	"os"
)

const defaultBaseDir = "/tmp/scs/ttp"

type Config struct {
	BaseDir string
	Port    string
}

func ConfigFromEnv() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		return Config{}, fmt.Errorf("environment variable PORT not set")
	}

	return Config{
		BaseDir: defaultBaseDir,
		Port:    port,
	}, nil
}

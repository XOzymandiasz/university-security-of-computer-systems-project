package identity

import (
	"encoding/base64"
	"os"
	"path/filepath"
)

const sessionFileName = "session.key"

func SaveSessionKey(baseDir string, sessionKey []byte) error {
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(sessionKey)

	return os.WriteFile(filepath.Join(baseDir, sessionFileName), []byte(encoded), 0600)
}

func LoadSessionKey(baseDir string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(baseDir, sessionFileName))
	if err != nil {
		return nil, err
	}

	return base64.StdEncoding.DecodeString(string(data))
}

func DeleteSessionKey(baseDir string) error {
	path := filepath.Join(baseDir, sessionFileName)

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

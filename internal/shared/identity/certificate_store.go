package identity

import (
	"os"
	"path/filepath"
)

func LoadCertificate(baseDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(baseDir, certFileName))
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func SaveCertificate(baseDir string, certificateBase64 string) error {
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return err
	}

	path := filepath.Join(baseDir, certFileName)

	return os.WriteFile(path, []byte(certificateBase64), 0600)
}

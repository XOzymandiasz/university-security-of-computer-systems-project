package identity

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
)

var (
	idFileName   = "id.txt"
	authFileName = "auth.key"
	encFileName  = "enc.key"
	certFileName = "cert.key"
)

func EnsureIdentity(baseDir string) error {
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return err
	}

	if err := ensureKey(filepath.Join(baseDir, authFileName)); err != nil {
		return err
	}

	if err := ensureKey(filepath.Join(baseDir, encFileName)); err != nil {
		return err
	}

	if err := ensureId(filepath.Join(baseDir, idFileName)); err != nil {
		return err
	}

	return nil
}

func ensureId(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}

	randomBytes := make([]byte, 32)
	if _, err = rand.Read(randomBytes); err != nil {
		return err
	}

	sum := sha256.Sum256(randomBytes)
	id := hex.EncodeToString(sum[:])

	return os.WriteFile(path, []byte(id), 0600)
}

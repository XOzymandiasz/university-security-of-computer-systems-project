package identity

import (
	"os"
	"path/filepath"
	"scs/internal/shared/protocol"
)

func LoadRegistrationData(baseDir string) (protocol.RegisterRequest, error) {
	idBytes, err := os.ReadFile(filepath.Join(baseDir, idFileName))
	if err != nil {
		return protocol.RegisterRequest{}, err
	}

	var authPub string
	authPub, err = loadPublicKeyBase64(filepath.Join(baseDir, authFileName))
	if err != nil {
		return protocol.RegisterRequest{}, err
	}

	var encPub string
	encPub, err = loadPublicKeyBase64(filepath.Join(baseDir, encFileName))
	if err != nil {
		return protocol.RegisterRequest{}, err
	}

	return protocol.RegisterRequest{
		EncryptedID:   string(idBytes),
		AuthPublicKey: authPub,
		EncPublicKey:  encPub,
	}, nil
}

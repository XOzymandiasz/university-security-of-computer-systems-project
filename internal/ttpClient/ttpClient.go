package ttpclient

import (
	"crypto/rsa"
	"fmt"
	"os"

	"scs/internal/identity"
	"scs/internal/protocol"
	"scs/internal/ttp"
)

func AddrFromEnv(envName string) (string, error) {
	addr := os.Getenv(envName)
	if addr == "" {
		return "", fmt.Errorf("TTP_ADDR env variable not set")
	}

	return addr, nil
}

func Init(addr string) (*rsa.PublicKey, error) {
	ttpPublicKey, err := ttp.Init(addr)
	if err != nil {
		return nil, err
	}
	fmt.Println("x1")
	if ttpPublicKey == nil {
		return nil, fmt.Errorf("TTP public key is nil")
	}

	return ttpPublicKey, nil
}

func Register(
	addr string,
	ttpPublicKey *rsa.PublicKey,
	data protocol.RegistrationData,
) (string, error) {
	if ttpPublicKey == nil {
		return "", fmt.Errorf("cannot register to TTP: public key is nil")
	}

	encryptedID, err := identity.EncryptWithPublicKeyBase64([]byte(data.ID), ttpPublicKey)
	if err != nil {
		return "", err
	}

	data.ID = encryptedID

	var certificateBase64 string
	certificateBase64, err = ttp.Register(addr, data)
	if err != nil {
		return "", err
	}

	if certificateBase64 == "" {
		return "", fmt.Errorf("empty certificate from TTP")
	}

	return certificateBase64, nil
}

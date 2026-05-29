package ttpservice

import (
	"path/filepath"
	"scs/internal/identity"
	"scs/internal/protocol"
)

type Service struct {
	baseDir string
}

func New(baseDir string) *Service {
	return &Service{
		baseDir: baseDir,
	}
}

func (s *Service) Init() (protocol.Message, error) {
	responseData := identity.LoadRegistrationData(s.baseDir)

	return protocol.Message{
		Type: "TTP_PUBLIC_KEY",
		Body: responseData.EncPublicKey,
	}, nil
}

func (s *Service) Register(reg protocol.RegisterRequest) (protocol.Message, error) {
	ttpEncPrivateKey, err := identity.LoadPrivateKey(filepath.Join(s.baseDir, "enc.key"))
	if err != nil {
		return protocol.Message{}, err
	}

	decryptedIDBytes, err := identity.DecryptWithPrivateKeyBase64(reg.EncryptedID, ttpEncPrivateKey)
	if err != nil {
		return protocol.Message{}, err
	}

	userAuthPublicKey, err := identity.ParsePublicKeyFromBase64(reg.AuthPublicKey)
	if err != nil {
		return protocol.Message{}, err
	}

	ttpAuthPrivateKey, err := identity.LoadPrivateKey(filepath.Join(s.baseDir, "auth.key"))
	if err != nil {
		return protocol.Message{}, err
	}

	certificateBase64, err := identity.CreateCertificateBase64(
		string(decryptedIDBytes),
		userAuthPublicKey,
		ttpAuthPrivateKey,
	)
	if err != nil {
		return protocol.Message{}, err
	}

	return protocol.Message{
		Type: "CERTIFICATE",
		Body: certificateBase64,
	}, nil
}

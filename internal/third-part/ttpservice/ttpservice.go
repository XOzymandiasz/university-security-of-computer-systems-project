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

func (s *Service) Init() (protocol.InitResponse, error) {
	responseData := identity.LoadRegistrationData(s.baseDir)

	return protocol.InitResponse{
		TTPEncPublicKey: responseData.EncPublicKey,
	}, nil
}

func (s *Service) Register(reg protocol.RegisterRequest) (protocol.RegisterResponse, error) {
	ttpEncPrivateKey, err := identity.LoadPrivateKey(filepath.Join(s.baseDir, "enc.key"))
	if err != nil {
		return protocol.RegisterResponse{}, err
	}

	decryptedIDBytes, err := identity.DecryptWithPrivateKeyBase64(reg.EncryptedID, ttpEncPrivateKey)
	if err != nil {
		return protocol.RegisterResponse{}, err
	}

	userAuthPublicKey, err := identity.ParsePublicKeyFromBase64(reg.AuthPublicKey)
	if err != nil {
		return protocol.RegisterResponse{}, err
	}

	ttpAuthPrivateKey, err := identity.LoadPrivateKey(filepath.Join(s.baseDir, "auth.key"))
	if err != nil {
		return protocol.RegisterResponse{}, err
	}

	certificateBase64, err := identity.CreateCertificateBase64(
		string(decryptedIDBytes),
		userAuthPublicKey,
		ttpAuthPrivateKey,
	)
	if err != nil {
		return protocol.RegisterResponse{}, err
	}

	return protocol.RegisterResponse{
		Certificate: certificateBase64,
	}, nil
}

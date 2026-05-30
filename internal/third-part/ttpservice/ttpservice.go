package ttpservice

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"path/filepath"
	"scs/internal/identity"
	"scs/internal/protocol"
)

type Service struct {
	baseDir string
	entries map[string]RegisteredEntity
}

func New(baseDir string) *Service {
	return &Service{
		baseDir: baseDir,
		entries: make(map[string]RegisteredEntity),
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

	var decryptedIDBytes []byte
	decryptedIDBytes, err = identity.DecryptWithPrivateKeyBase64(reg.EncryptedID, ttpEncPrivateKey)
	if err != nil {
		return protocol.RegisterResponse{}, err
	}

	var userAuthPublicKey *rsa.PublicKey
	userAuthPublicKey, err = identity.ParsePublicKeyFromBase64(reg.AuthPublicKey)
	if err != nil {
		return protocol.RegisterResponse{}, err
	}

	var ttpAuthPrivateKey *rsa.PrivateKey
	ttpAuthPrivateKey, err = identity.LoadPrivateKey(filepath.Join(s.baseDir, "auth.key"))
	if err != nil {
		return protocol.RegisterResponse{}, err
	}

	var certificateBase64 string
	certificateBase64, err = identity.CreateCertificateBase64(
		string(decryptedIDBytes),
		userAuthPublicKey,
		ttpAuthPrivateKey,
	)
	if err != nil {
		return protocol.RegisterResponse{}, err
	}

	entityID := string(decryptedIDBytes)

	if _, exists := s.entries[entityID]; exists {
		return protocol.RegisterResponse{}, fmt.Errorf("entity already registered")
	}

	s.entries[entityID] = RegisteredEntity{
		ID:            entityID,
		Role:          string(reg.Role),
		EncPublicKey:  reg.EncPublicKey,
		AuthPublicKey: reg.AuthPublicKey,
		Certificate:   certificateBase64,
	}

	return protocol.RegisterResponse{
		Certificate: certificateBase64,
	}, nil
}

func (s *Service) Authenticate(req protocol.AuthenticateRequest) (protocol.AuthenticateResponse, error) {
	ttpEncPrivateKey, err := identity.LoadPrivateKey(filepath.Join(s.baseDir, "enc.key"))
	if err != nil {
		return protocol.AuthenticateResponse{}, fmt.Errorf("load ttp enc private key: %w", err)
	}

	serverEntity, exists := s.entries[req.ServerID]
	if !exists {
		return protocol.AuthenticateResponse{}, fmt.Errorf("server not registered")
	}

	if serverEntity.Role != string(protocol.EntityRoleServer) {
		return protocol.AuthenticateResponse{}, fmt.Errorf("entity is not server")
	}

	var ttpAuthPrivateKey *rsa.PrivateKey
	ttpAuthPrivateKey, err = identity.LoadPrivateKey(filepath.Join(s.baseDir, "auth.key"))
	if err != nil {
		return protocol.AuthenticateResponse{}, fmt.Errorf("load ttp auth private key: %w", err)
	}

	if err = identity.ValidateCertificateBase64(
		req.ServerCertificate,
		serverEntity.ID,
		serverEntity.AuthPublicKey,
		&ttpAuthPrivateKey.PublicKey,
	); err != nil {
		return protocol.AuthenticateResponse{}, fmt.Errorf("invalid server certificate: %w", err)
	}

	var serverAuthPublicKey *rsa.PublicKey
	serverAuthPublicKey, err = identity.ParsePublicKeyFromBase64(serverEntity.AuthPublicKey)
	if err != nil {
		return protocol.AuthenticateResponse{}, fmt.Errorf("parse server auth public key: %w", err)
	}

	if err = identity.VerifySignatureBase64([]byte(req.ServerID), req.ServerSignature, serverAuthPublicKey); err != nil {
		return protocol.AuthenticateResponse{}, fmt.Errorf("invalid server signature: %w", err)
	}

	var clientPayloadBytes []byte
	clientPayloadBytes, err = identity.DecryptLargePayloadWithPrivateKeyBase64(
		req.ClientEncryptedPayload,
		ttpEncPrivateKey,
	)
	if err != nil {
		return protocol.AuthenticateResponse{}, fmt.Errorf("decrypt client payload: %w", err)
	}

	var clientPayload protocol.AuthenticateClientPayload
	if err = json.Unmarshal(clientPayloadBytes, &clientPayload); err != nil {
		return protocol.AuthenticateResponse{}, fmt.Errorf("decode client payload: %w", err)
	}

	var clientEntity RegisteredEntity
	clientEntity, exists = s.entries[clientPayload.ClientID]
	if !exists {
		return protocol.AuthenticateResponse{}, fmt.Errorf("client not registered")
	}

	if clientEntity.Role != string(protocol.EntityRoleClient) {
		return protocol.AuthenticateResponse{}, fmt.Errorf("entity is not client")
	}

	if err = identity.ValidateCertificateBase64(
		clientPayload.ClientCertificate,
		clientEntity.ID,
		clientEntity.AuthPublicKey,
		&ttpAuthPrivateKey.PublicKey,
	); err != nil {
		return protocol.AuthenticateResponse{}, fmt.Errorf("invalid client certificate: %w", err)
	}

	var clientAuthPublicKey *rsa.PublicKey
	clientAuthPublicKey, err = identity.ParsePublicKeyFromBase64(clientEntity.AuthPublicKey)
	if err != nil {
		return protocol.AuthenticateResponse{}, fmt.Errorf("parse client auth public key: %w", err)
	}

	if err = identity.VerifySignatureBase64([]byte(clientPayload.ClientID), clientPayload.ClientSignature, clientAuthPublicKey); err != nil {
		return protocol.AuthenticateResponse{}, fmt.Errorf("invalid client signature: %w", err)
	}

	var sessionKey []byte
	sessionKey, err = identity.GenerateRandomBytes(32)
	if err != nil {
		return protocol.AuthenticateResponse{}, fmt.Errorf("generate session key: %w", err)
	}

	var clientEncPublicKey *rsa.PublicKey
	clientEncPublicKey, err = identity.ParsePublicKeyFromBase64(clientEntity.EncPublicKey)
	if err != nil {
		return protocol.AuthenticateResponse{}, fmt.Errorf("parse client enc public key: %w", err)
	}

	var encryptedSessionKeyForClient string
	encryptedSessionKeyForClient, err = identity.EncryptWithPublicKeyBase64(sessionKey, clientEncPublicKey)
	if err != nil {
		return protocol.AuthenticateResponse{}, fmt.Errorf("encrypt session key for client: %w", err)
	}

	var serverEncPublicKey *rsa.PublicKey
	serverEncPublicKey, err = identity.ParsePublicKeyFromBase64(serverEntity.EncPublicKey)
	if err != nil {
		return protocol.AuthenticateResponse{}, fmt.Errorf("parse server enc public key: %w", err)
	}

	var encryptedSessionKeyForServer string
	encryptedSessionKeyForServer, err = identity.EncryptWithPublicKeyBase64(sessionKey, serverEncPublicKey)
	if err != nil {
		return protocol.AuthenticateResponse{}, fmt.Errorf("encrypt session key for server: %w", err)
	}

	return protocol.AuthenticateResponse{
		OK:                           true,
		EncryptedSessionKeyForClient: encryptedSessionKeyForClient,
		EncryptedSessionKeyForServer: encryptedSessionKeyForServer,
		Message:                      "authenticated",
	}, nil
}

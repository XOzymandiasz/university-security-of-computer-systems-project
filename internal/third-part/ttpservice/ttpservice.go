package ttpservice

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"path/filepath"
	identity2 "scs/internal/shared/identity"
	protocol2 "scs/internal/shared/protocol"
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

func (s *Service) Init() (protocol2.InitResponse, error) {
	responseData, err := identity2.LoadRegistrationData(s.baseDir)

	if err != nil {
		fmt.Println(err)
	}

	return protocol2.InitResponse{
		TTPEncPublicKey: responseData.EncPublicKey,
	}, nil
}

func (s *Service) Register(reg protocol2.RegisterRequest) (protocol2.RegisterResponse, error) {
	ttpEncPrivateKey, err := identity2.LoadPrivateKey(filepath.Join(s.baseDir, "enc.key"))
	if err != nil {
		return protocol2.RegisterResponse{}, err
	}

	var decryptedIDBytes []byte
	decryptedIDBytes, err = identity2.DecryptWithPrivateKeyBase64(reg.EncryptedID, ttpEncPrivateKey)
	if err != nil {
		return protocol2.RegisterResponse{}, err
	}

	var userAuthPublicKey *rsa.PublicKey
	userAuthPublicKey, err = identity2.ParsePublicKeyFromBase64(reg.AuthPublicKey)
	if err != nil {
		return protocol2.RegisterResponse{}, err
	}

	var ttpAuthPrivateKey *rsa.PrivateKey
	ttpAuthPrivateKey, err = identity2.LoadPrivateKey(filepath.Join(s.baseDir, "auth.key"))
	if err != nil {
		return protocol2.RegisterResponse{}, err
	}

	var certificateBase64 string
	certificateBase64, err = CreateCertificateBase64(
		string(decryptedIDBytes),
		userAuthPublicKey,
		ttpAuthPrivateKey,
	)
	if err != nil {
		return protocol2.RegisterResponse{}, err
	}

	entityID := string(decryptedIDBytes)

	if _, exists := s.entries[entityID]; exists {
		return protocol2.RegisterResponse{}, fmt.Errorf("entity already registered")
	}

	s.entries[entityID] = RegisteredEntity{
		ID:            entityID,
		Role:          string(reg.Role),
		EncPublicKey:  reg.EncPublicKey,
		AuthPublicKey: reg.AuthPublicKey,
		Certificate:   certificateBase64,
	}

	return protocol2.RegisterResponse{
		Certificate: certificateBase64,
	}, nil
}

func (s *Service) Authenticate(req protocol2.AuthenticateRequest) (protocol2.AuthenticateResponse, error) {
	ttpEncPrivateKey, err := identity2.LoadPrivateKey(filepath.Join(s.baseDir, "enc.key"))
	if err != nil {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("load ttp enc private key: %w", err)
	}

	serverEntity, exists := s.entries[req.ServerID]
	if !exists {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("server not registered")
	}

	if serverEntity.Role != string(protocol2.EntityRoleServer) {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("entity is not server")
	}

	var ttpAuthPrivateKey *rsa.PrivateKey
	ttpAuthPrivateKey, err = identity2.LoadPrivateKey(filepath.Join(s.baseDir, "auth.key"))
	if err != nil {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("load ttp auth private key: %w", err)
	}

	if err = ValidateCertificateBase64(
		req.ServerCertificate,
		serverEntity.ID,
		serverEntity.AuthPublicKey,
		&ttpAuthPrivateKey.PublicKey,
	); err != nil {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("invalid server certificate: %w", err)
	}

	var serverAuthPublicKey *rsa.PublicKey
	serverAuthPublicKey, err = identity2.ParsePublicKeyFromBase64(serverEntity.AuthPublicKey)
	if err != nil {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("parse server auth public key: %w", err)
	}

	if err = VerifySignatureBase64([]byte(req.ServerID), req.ServerSignature, serverAuthPublicKey); err != nil {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("invalid server signature: %w", err)
	}

	var clientPayloadBytes []byte
	clientPayloadBytes, err = DecryptLargePayloadWithPrivateKeyBase64(
		req.ClientEncryptedPayload,
		ttpEncPrivateKey,
	)
	if err != nil {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("decrypt client payload: %w", err)
	}

	var clientPayload protocol2.AuthenticateClientPayload
	if err = json.Unmarshal(clientPayloadBytes, &clientPayload); err != nil {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("decode client payload: %w", err)
	}

	var clientEntity RegisteredEntity
	clientEntity, exists = s.entries[clientPayload.ClientID]
	if !exists {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("client not registered")
	}

	if clientEntity.Role != string(protocol2.EntityRoleClient) {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("entity is not client")
	}

	if err = ValidateCertificateBase64(
		clientPayload.ClientCertificate,
		clientEntity.ID,
		clientEntity.AuthPublicKey,
		&ttpAuthPrivateKey.PublicKey,
	); err != nil {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("invalid client certificate: %w", err)
	}

	var clientAuthPublicKey *rsa.PublicKey
	clientAuthPublicKey, err = identity2.ParsePublicKeyFromBase64(clientEntity.AuthPublicKey)
	if err != nil {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("parse client auth public key: %w", err)
	}

	if err = VerifySignatureBase64([]byte(clientPayload.ClientID), clientPayload.ClientSignature, clientAuthPublicKey); err != nil {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("invalid client signature: %w", err)
	}

	var sessionKey []byte
	sessionKey, err = identity2.GenerateRandomBytes(32)
	if err != nil {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("generate session key: %w", err)
	}

	var clientEncPublicKey *rsa.PublicKey
	clientEncPublicKey, err = identity2.ParsePublicKeyFromBase64(clientEntity.EncPublicKey)
	if err != nil {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("parse client enc public key: %w", err)
	}

	var encryptedSessionKeyForClient string
	encryptedSessionKeyForClient, err = identity2.EncryptWithPublicKeyBase64(sessionKey, clientEncPublicKey)
	if err != nil {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("encrypt session key for client: %w", err)
	}

	var serverEncPublicKey *rsa.PublicKey
	serverEncPublicKey, err = identity2.ParsePublicKeyFromBase64(serverEntity.EncPublicKey)
	if err != nil {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("parse server enc public key: %w", err)
	}

	var encryptedSessionKeyForServer string
	encryptedSessionKeyForServer, err = identity2.EncryptWithPublicKeyBase64(sessionKey, serverEncPublicKey)
	if err != nil {
		return protocol2.AuthenticateResponse{}, fmt.Errorf("encrypt session key for server: %w", err)
	}

	return protocol2.AuthenticateResponse{
		OK:                           true,
		EncryptedSessionKeyForClient: encryptedSessionKeyForClient,
		EncryptedSessionKeyForServer: encryptedSessionKeyForServer,
		Message:                      "authenticated",
	}, nil
}

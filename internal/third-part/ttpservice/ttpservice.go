package ttpservice

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"path/filepath"

	"scs/internal/shared/identity"
	"scs/internal/shared/protocol"
)

// Service reprezentuje usługę Trusted Third Party.
//
// Struktura przechowuje katalog bazowy lokalnej tożsamości TTP oraz mapę
// zarejestrowanych podmiotów, czyli klientów i serwerów uczestniczących
// w protokole.
type Service struct {
	baseDir string
	entries map[string]RegisteredEntity
}

// New tworzy nową instancję usługi TTP.
//
// Funkcja inicjalizuje pusty rejestr podmiotów oraz zapisuje katalog bazowy,
// z którego TTP odczytuje własne klucze RSA.
// @param baseDir Katalog bazowy lokalnej tożsamości TTP.
// @return Wskaźnik do nowej instancji Service.
func New(baseDir string) *Service {
	return &Service{
		baseDir: baseDir,
		entries: make(map[string]RegisteredEntity),
	}
}

// Init zwraca publiczny klucz szyfrujący TTP.
//
// Funkcja odczytuje lokalne dane rejestracyjne TTP i zwraca publiczny klucz,
// którym klient oraz serwer mogą szyfrować dane przesyłane do zaufanej strony
// trzeciej.
//
// @return Odpowiedź inicjalizacyjna zawierająca publiczny klucz TTP lub błąd.
func (s *Service) Init() (protocol.InitResponse, error) {
	responseData, err := identity.LoadRegistrationData(s.baseDir)

	if err != nil {
		fmt.Println(err)
	}

	return protocol.InitResponse{
		TTPEncPublicKey: responseData.EncPublicKey,
	}, nil
}

// Register rejestruje klienta albo serwer w usłudze TTP.
//
// Funkcja odszyfrowuje identyfikator podmiotu kluczem prywatnym TTP,
// parsuje publiczny klucz uwierzytelniający, generuje certyfikat X.509
// i zapisuje podmiot w lokalnym rejestrze zarejestrowanych aplikacji.
//
// @param reg Żądanie rejestracji zawierające zaszyfrowany identyfikator,
// publiczne klucze RSA oraz rolę podmiotu.
// @return Odpowiedź rejestracji z certyfikatem X.509 lub błąd.
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
	certificateBase64, err = CreateCertificateBase64(
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

// Authenticate uwierzytelnia klienta i serwer z udziałem TTP.
//
// Funkcja weryfikuje, czy serwer i klient są zarejestrowani, sprawdza ich role,
// waliduje certyfikaty X.509 oraz podpisy cyfrowe. Po pozytywnej walidacji
// generuje 256-bitowy klucz sesyjny AES i szyfruje go osobno kluczami
// publicznymi klienta oraz serwera.
//
// @param req Żądanie uwierzytelnienia zawierające dane serwera i zaszyfrowany pakiet klienta.
// @return Odpowiedź z zaszyfrowanymi kluczami sesyjnymi lub błąd walidacji.
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

	if err = ValidateCertificateBase64(
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

	if err = VerifySignatureBase64([]byte(req.ServerID), req.ServerSignature, serverAuthPublicKey); err != nil {
		return protocol.AuthenticateResponse{}, fmt.Errorf("invalid server signature: %w", err)
	}

	var clientPayloadBytes []byte
	clientPayloadBytes, err = DecryptLargePayloadWithPrivateKeyBase64(
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

	if err = ValidateCertificateBase64(
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

	if err = VerifySignatureBase64([]byte(clientPayload.ClientID), clientPayload.ClientSignature, clientAuthPublicKey); err != nil {
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

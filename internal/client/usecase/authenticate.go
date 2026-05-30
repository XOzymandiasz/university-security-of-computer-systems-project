package usecase

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"path/filepath"

	"scs/internal/identity"
	"scs/internal/protocol"
)

type ServerAuthenticator interface {
	Authenticate(req protocol.ClientAuthenticateRequest) (protocol.ClientAuthenticateResponse, error)
}

type TTPInitializer interface {
	Init() (*rsa.PublicKey, error)
}

type Authenticate struct {
	baseDir      string
	serverClient ServerAuthenticator
	ttpClient    TTPInitializer
}

func NewAuthenticate(
	baseDir string,
	serverClient ServerAuthenticator,
	ttpClient TTPInitializer,
) *Authenticate {
	return &Authenticate{
		baseDir:      baseDir,
		serverClient: serverClient,
		ttpClient:    ttpClient,
	}
}

func (a *Authenticate) Authenticate() error {
	ttpPublicKey, err := a.ttpClient.Init()
	if err != nil {
		return fmt.Errorf("ttp init: %w", err)
	}

	clientData := identity.LoadRegistrationData(a.baseDir)

	clientCertificate, err := identity.LoadCertificate(a.baseDir)
	if err != nil {
		return fmt.Errorf("load client certificate: %w", err)
	}

	clientAuthPrivateKey, err := identity.LoadPrivateKey(filepath.Join(a.baseDir, "auth.key"))
	if err != nil {
		return fmt.Errorf("load client auth private key: %w", err)
	}

	clientSignature, err := identity.SignBase64([]byte(clientData.EncryptedID), clientAuthPrivateKey)
	if err != nil {
		return fmt.Errorf("sign client id: %w", err)
	}

	payload := protocol.AuthenticateClientPayload{
		ClientID:          clientData.EncryptedID,
		ClientCertificate: clientCertificate,
		ClientSignature:   clientSignature,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal client auth payload: %w", err)
	}

	encryptedPayload, err := identity.EncryptLargePayloadWithPublicKeyBase64(payloadBytes, ttpPublicKey)
	if err != nil {
		return fmt.Errorf("encrypt client auth payload: %w", err)
	}

	serverResp, err := a.serverClient.Authenticate(protocol.ClientAuthenticateRequest{
		ClientEncryptedPayload: encryptedPayload,
	})
	if err != nil {
		return fmt.Errorf("server authenticate: %w", err)
	}

	clientEncPrivateKey, err := identity.LoadPrivateKey(filepath.Join(a.baseDir, "enc.key"))
	if err != nil {
		return fmt.Errorf("load client enc private key: %w", err)
	}

	sessionKey, err := identity.DecryptWithPrivateKeyBase64(
		serverResp.EncryptedSessionKeyForClient,
		clientEncPrivateKey,
	)
	if err != nil {
		return fmt.Errorf("decrypt client session key: %w", err)
	}

	if err := identity.SaveSessionKey(a.baseDir, sessionKey); err != nil {
		return fmt.Errorf("save client session key: %w", err)
	}

	return nil
}

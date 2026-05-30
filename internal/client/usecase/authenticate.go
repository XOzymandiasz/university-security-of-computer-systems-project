package usecase

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"path/filepath"
	identity2 "scs/internal/shared/identity"
	"scs/internal/shared/protocol"
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

	clientData, err := identity2.LoadRegistrationData(a.baseDir)

	if err != nil {
		return fmt.Errorf("load registration data: %w", err)
	}

	var clientCertificate string
	clientCertificate, err = identity2.LoadCertificate(a.baseDir)
	if err != nil {
		return fmt.Errorf("load client certificate: %w", err)
	}

	var clientAuthPrivateKey *rsa.PrivateKey
	clientAuthPrivateKey, err = identity2.LoadPrivateKey(filepath.Join(a.baseDir, "auth.key"))
	if err != nil {
		return fmt.Errorf("load client auth private key: %w", err)
	}

	var clientSignature string
	clientSignature, err = identity2.SignBase64([]byte(clientData.EncryptedID), clientAuthPrivateKey)
	if err != nil {
		return fmt.Errorf("sign client id: %w", err)
	}

	payload := protocol.AuthenticateClientPayload{
		ClientID:          clientData.EncryptedID,
		ClientCertificate: clientCertificate,
		ClientSignature:   clientSignature,
	}

	var payloadBytes []byte
	payloadBytes, err = json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal client auth payload: %w", err)
	}

	var encryptedPayload string
	encryptedPayload, err = identity2.EncryptLargePayloadWithPublicKeyBase64(payloadBytes, ttpPublicKey)
	if err != nil {
		return fmt.Errorf("encrypt client auth payload: %w", err)
	}

	var serverResp protocol.ClientAuthenticateResponse
	serverResp, err = a.serverClient.Authenticate(protocol.ClientAuthenticateRequest{
		ClientEncryptedPayload: encryptedPayload,
	})
	if err != nil {
		return fmt.Errorf("server authenticate: %w", err)
	}

	var clientEncPrivateKey *rsa.PrivateKey
	clientEncPrivateKey, err = identity2.LoadPrivateKey(filepath.Join(a.baseDir, "enc.key"))
	if err != nil {
		return fmt.Errorf("load client enc private key: %w", err)
	}

	var sessionKey []byte
	sessionKey, err = identity2.DecryptWithPrivateKeyBase64(
		serverResp.EncryptedSessionKeyForClient,
		clientEncPrivateKey,
	)
	if err != nil {
		return fmt.Errorf("decrypt client session key: %w", err)
	}

	if err = identity2.SaveSessionKey(a.baseDir, sessionKey); err != nil {
		return fmt.Errorf("save client session key: %w", err)
	}

	return nil
}

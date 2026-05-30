package identity

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"
)

func CreateCertificateBase64(
	subjectID string,
	subjectPublicKey *rsa.PublicKey,
	issuerPrivateKey *rsa.PrivateKey,
) (string, error) {
	serialLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	serialNumber, err := rand.Int(rand.Reader, serialLimit)
	if err != nil {
		return "", err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,

		SignatureAlgorithm: x509.SHA256WithRSA,

		Subject: pkix.Name{
			CommonName: subjectID,
		},

		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),

		KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,

		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},

		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	issuerTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),

		SignatureAlgorithm: x509.SHA256WithRSA,

		Subject: pkix.Name{
			CommonName: "SCS TTP",
		},

		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(10 * 365 * 24 * time.Hour),

		KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,

		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	var certDER []byte
	certDER, err = x509.CreateCertificate(
		rand.Reader,
		&template,
		&issuerTemplate,
		subjectPublicKey,
		issuerPrivateKey,
	)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(certDER), nil
}

func EncryptWithPublicKeyBase64(data []byte, pub *rsa.PublicKey) (string, error) {
	hash := sha256.New()

	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, pub, data, nil)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DecryptWithPrivateKeyBase64(encoded string, privateKey *rsa.PrivateKey) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	hash := sha256.New()

	var plaintext []byte
	plaintext, err = rsa.DecryptOAEP(hash, rand.Reader, privateKey, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func ParsePublicKeyFromBase64(encoded string) (*rsa.PublicKey, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	var key any
	key, err = x509.ParsePKIXPublicKey(keyBytes)
	if err != nil {
		return nil, err
	}

	pub, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("no RSA public key")
	}

	return pub, nil
}

type HybridEncryptedPayload struct {
	EncryptedKey string `json:"encrypted_key"`
	Nonce        string `json:"nonce"`
	Ciphertext   string `json:"ciphertext"`
}

func EncryptLargePayloadWithPublicKeyBase64(data []byte, pub *rsa.PublicKey) (string, error) {
	aesKey := make([]byte, 32)

	if _, err := rand.Read(aesKey); err != nil {
		return "", err
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())

	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nil, nonce, data, nil)

	encryptedKey, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		pub,
		aesKey,
		nil,
	)
	if err != nil {
		return "", err
	}

	payload := HybridEncryptedPayload{
		EncryptedKey: base64.StdEncoding.EncodeToString(encryptedKey),
		Nonce:        base64.StdEncoding.EncodeToString(nonce),
		Ciphertext:   base64.StdEncoding.EncodeToString(ciphertext),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(payloadBytes), nil
}

func DecryptLargePayloadWithPrivateKeyBase64(encoded string, privateKey *rsa.PrivateKey) ([]byte, error) {
	payloadBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	var payload HybridEncryptedPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, err
	}

	encryptedKey, err := base64.StdEncoding.DecodeString(payload.EncryptedKey)
	if err != nil {
		return nil, err
	}

	nonce, err := base64.StdEncoding.DecodeString(payload.Nonce)
	if err != nil {
		return nil, err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(payload.Ciphertext)
	if err != nil {
		return nil, err
	}

	aesKey, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		privateKey,
		encryptedKey,
		nil,
	)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func SignBase64(data []byte, privateKey *rsa.PrivateKey) (string, error) {
	digest := sha256.Sum256(data)

	signature, err := rsa.SignPSS(
		rand.Reader,
		privateKey,
		crypto.SHA256,
		digest[:],
		nil,
	)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

func VerifySignatureBase64(data []byte, signatureBase64 string, publicKey *rsa.PublicKey) error {
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return err
	}

	digest := sha256.Sum256(data)

	return rsa.VerifyPSS(
		publicKey,
		crypto.SHA256,
		digest[:],
		signature,
		nil,
	)
}

func EncryptWithSessionKeyBase64(plaintext []byte, sessionKey []byte) (string, error) {
	if len(sessionKey) != 32 {
		return "", fmt.Errorf("invalid AES-256 session key length: %d", len(sessionKey))
	}

	block, err := aes.NewCipher(sessionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	payload := HybridEncryptedPayload{
		Nonce:      base64.StdEncoding.EncodeToString(nonce),
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(payloadBytes), nil
}

func DecryptWithSessionKeyBase64(encoded string, sessionKey []byte) ([]byte, error) {
	if len(sessionKey) != 32 {
		return nil, fmt.Errorf("invalid AES-256 session key length: %d", len(sessionKey))
	}

	payloadBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	var payload HybridEncryptedPayload
	if err = json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, err
	}

	nonce, err := base64.StdEncoding.DecodeString(payload.Nonce)
	if err != nil {
		return nil, err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(payload.Ciphertext)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(sessionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, nonce, ciphertext, nil)
}

func ParseCertificateBase64(encoded string) (*x509.Certificate, error) {
	certDER, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	return x509.ParseCertificate(certDER)
}

func ValidateCertificateBase64(
	certificateBase64 string,
	expectedSubjectID string,
	expectedPublicKeyBase64 string,
	issuerPublicKey *rsa.PublicKey,
) error {
	cert, err := ParseCertificateBase64(certificateBase64)
	if err != nil {
		return fmt.Errorf("parse certificate: %w", err)
	}

	now := time.Now()
	if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
		return fmt.Errorf("certificate expired or not yet valid")
	}

	if cert.Subject.CommonName != expectedSubjectID {
		return fmt.Errorf("invalid certificate subject: got=%s expected=%s", cert.Subject.CommonName, expectedSubjectID)
	}

	expectedPublicKey, err := ParsePublicKeyFromBase64(expectedPublicKeyBase64)
	if err != nil {
		return fmt.Errorf("parse expected public key: %w", err)
	}

	certPublicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("certificate does not contain RSA public key")
	}

	if certPublicKey.N.Cmp(expectedPublicKey.N) != 0 || certPublicKey.E != expectedPublicKey.E {
		return fmt.Errorf("certificate public key mismatch")
	}

	digest := sha256.Sum256(cert.RawTBSCertificate)

	if err = rsa.VerifyPKCS1v15(
		issuerPublicKey,
		crypto.SHA256,
		digest[:],
		cert.Signature,
	); err != nil {
		return fmt.Errorf("invalid certificate signature: %w", err)
	}

	return nil
}

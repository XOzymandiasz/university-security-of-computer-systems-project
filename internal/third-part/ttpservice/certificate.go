package ttpservice

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"fmt"
	"math/big"
	"scs/internal/shared/identity"
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

	var expectedPublicKey *rsa.PublicKey
	expectedPublicKey, err = identity.ParsePublicKeyFromBase64(expectedPublicKeyBase64)
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

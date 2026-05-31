// Package ttpservice zawiera logikę usługi Trusted Third Party.
//
// Pakiet odpowiada za tworzenie, parsowanie i walidację certyfikatów X.509
// używanych podczas rejestracji oraz uwierzytelniania klienta i serwera.
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
	"time"

	"scs/internal/shared/identity"
)

// CreateCertificateBase64 tworzy certyfikat X.509 dla klienta lub serwera.
//
// Funkcja generuje certyfikat dla podanego identyfikatora i klucza publicznego.
// Certyfikat jest podpisywany kluczem prywatnym TTP, dzięki czemu może być
// później weryfikowany podczas procesu uwierzytelniania.
//
// @param subjectID Identyfikator podmiotu, dla którego tworzony jest certyfikat.
// @param subjectPublicKey Klucz publiczny RSA podmiotu.
// @param issuerPrivateKey Klucz prywatny RSA TTP używany do podpisania certyfikatu.
// @return Certyfikat X.509 zakodowany w formacie Base64 lub błąd tworzenia.
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

// ParseCertificateBase64 parsuje certyfikat X.509 zapisany w formacie Base64.
//
// Funkcja dekoduje dane Base64 do postaci DER, a następnie odczytuje certyfikat
// przy użyciu standardowego parsera x509.
//
// @param encoded Certyfikat X.509 zakodowany w formacie Base64.
// @return Struktura certyfikatu X.509 lub błąd parsowania.
func ParseCertificateBase64(encoded string) (*x509.Certificate, error) {
	certDER, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	return x509.ParseCertificate(certDER)
}

// ValidateCertificateBase64 weryfikuje poprawność certyfikatu X.509.
//
// Funkcja sprawdza poprawność okresu ważności, zgodność identyfikatora,
// zgodność klucza publicznego oraz podpis certyfikatu wykonany przez TTP.
// Negatywny wynik walidacji oznacza, że certyfikat jest nieważny,
// nie pasuje do deklarowanej tożsamości albo został podrobiony.
//
// @param certificateBase64 Certyfikat X.509 zakodowany w formacie Base64.
// @param expectedSubjectID Oczekiwany identyfikator właściciela certyfikatu.
// @param expectedPublicKeyBase64 Oczekiwany klucz publiczny właściciela w formacie Base64.
// @param issuerPublicKey Publiczny klucz RSA TTP używany do weryfikacji podpisu.
// @return Błąd walidacji certyfikatu lub nil w przypadku powodzenia.
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

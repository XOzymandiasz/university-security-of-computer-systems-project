package identity

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"scs/internal/protocol"
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

		Subject: pkix.Name{
			CommonName: "SCS TTP",
		},

		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(10 * 365 * 24 * time.Hour),

		KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,

		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certDER, err := x509.CreateCertificate(
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

	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, privateKey, ciphertext, nil)
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

	key, err := x509.ParsePKIXPublicKey(keyBytes)
	if err != nil {
		return nil, err
	}

	pub, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("no RSA public key")
	}

	return pub, nil
}

func LoadRegistrationData(baseDir string) protocol.RegistrationData {
	idBytes, err := os.ReadFile(baseDir + "id.txt")
	if err != nil {
		log.Fatal(err)
	}
	authPub := loadPublicKeyBase64(baseDir + "auth.key")
	encPub := loadPublicKeyBase64(baseDir + "enc.key")

	return protocol.RegistrationData{
		ID:            string(idBytes),
		AuthPublicKey: authPub,
		EncPublicKey:  encPub,
	}
}

func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyPEM, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return nil, errors.New("failed to decode PEM")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func loadPublicKeyBase64(path string) string {
	keyPEM, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	block, _ := pem.Decode(keyPEM)
	if block == nil {
		log.Fatal("failed to decode PEM")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Fatal(err)
	}

	publicKey, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatal(err)
	}

	return base64.StdEncoding.EncodeToString(publicKey)
}

func EnsureIdentity(baseDir string) {
	err := os.MkdirAll(baseDir, 0700)
	if err != nil {
		log.Fatal(err)
	}

	ensureKey(baseDir + "auth.key")
	ensureKey(baseDir + "enc.key")
	ensureId(baseDir + "id.txt")

	fmt.Println("Identity ready")
}

func ensureKey(path string) {
	_, err := os.Stat(path)
	if err == nil {
		return
	}
	if !os.IsNotExist(err) {
		log.Fatal(err)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}
	keyBytes := x509.MarshalPKCS1PrivateKey(privateKey)

	pemBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyBytes}
	file, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	err = pem.Encode(file, pemBlock)
	if err != nil {
		log.Fatal(err)
	}
}

func ensureId(path string) {
	_, err := os.Stat(path)
	if err == nil {
		return
	}
	if !os.IsNotExist(err) {
		log.Fatal(err)
	}

	randomBytes := make([]byte, 32)
	_, err = rand.Read(randomBytes)
	if err != nil {
		log.Fatal(err)
	}
	sum := sha256.Sum256(randomBytes)
	id := hex.EncodeToString(sum[:])

	err = os.WriteFile(path, []byte(id), 0600)
	if err != nil {
		log.Fatal(err)
	}
}

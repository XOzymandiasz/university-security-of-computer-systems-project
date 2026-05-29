package identity

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"scs/internal/protocol"
)

var idFileName = "id.txt"
var authFileName = "auth.key"
var encFileName = "enc.key"

func LoadRegistrationData(baseDir string) protocol.RegistrationData {
	idBytes, err := os.ReadFile(filepath.Join(baseDir, idFileName))
	if err != nil {
		log.Fatal(err)
	}
	authPub := loadPublicKeyBase64(filepath.Join(baseDir, authFileName))
	encPub := loadPublicKeyBase64(filepath.Join(baseDir, encFileName))

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

	var privateKey *rsa.PrivateKey
	privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
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

	var privateKey *rsa.PrivateKey
	privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Fatal(err)
	}

	var publicKey []byte
	publicKey, err = x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
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

	ensureKey(filepath.Join(baseDir, authFileName))
	ensureKey(filepath.Join(baseDir, encFileName))
	ensureId(filepath.Join(baseDir, idFileName))

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

	var privateKey *rsa.PrivateKey
	privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}
	keyBytes := x509.MarshalPKCS1PrivateKey(privateKey)

	pemBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyBytes}
	var file *os.File
	file, err = os.Create(path)
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

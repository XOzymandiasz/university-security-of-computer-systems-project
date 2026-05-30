package identity

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"os"
)

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

func loadPublicKeyBase64(path string) (string, error) {
	keyPEM, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM")
	}

	var privateKey *rsa.PrivateKey
	privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	var publicKey []byte
	publicKey, err = x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(publicKey), nil
}

func ensureKey(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}

	var privateKey *rsa.PrivateKey
	privateKey, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
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

	return pem.Encode(file, pemBlock)
}

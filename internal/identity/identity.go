package identity

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"log"
	"os"

	"scs/internal/protocol"
)

func LoadRegistrationData(baseDir string) protocol.RegistrationData {
	idBytes, err := os.ReadFile(baseDir + "/id.txt")
	if err != nil {
		log.Fatal(err)
	}

	authPub := loadPublicKeyBase64(baseDir + "/auth.key")
	encPub := loadPublicKeyBase64(baseDir + "/enc.key")

	return protocol.RegistrationData{
		ID:            string(idBytes),
		AuthPublicKey: authPub,
		EncPublicKey:  encPub,
	}
}

func loadPublicKeyBase64(privateKeyPath string) string {
	keyPEM, err := os.ReadFile(privateKeyPath)
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

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatal(err)
	}

	return base64.StdEncoding.EncodeToString(publicKeyBytes)
}

func EnsureIdentity(baseDir string) {
	err := os.MkdirAll(baseDir, 0700)
	if err != nil {
		log.Fatal(err)
	}

	ensureKey(baseDir + "/auth.key")
	ensureKey(baseDir + "/enc.key")
	ensureId(baseDir + "/id.txt")

	fmt.Println("Server identity ready")
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

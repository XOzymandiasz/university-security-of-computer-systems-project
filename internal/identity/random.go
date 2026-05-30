package identity

import "crypto/rand"

func GenerateRandomBytes(size int) ([]byte, error) {
	data := make([]byte, size)

	if _, err := rand.Read(data); err != nil {
		return nil, err
	}

	return data, nil
}

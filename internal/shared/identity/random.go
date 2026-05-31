package identity

import "crypto/rand"

// GenerateRandomBytes generuje losowe bajty o podanej długości.
//
// Funkcja wykorzystuje kryptograficznie bezpieczny generator losowy
// z pakietu crypto/rand. Jest używana do tworzenia wartości wymagających
// wysokiej losowości, takich jak klucze sesyjne, identyfikatory lub nonce.
//
// @param size Liczba bajtów, które mają zostać wygenerowane.
// @return Wygenerowane losowe bajty lub błąd generatora losowego.
func GenerateRandomBytes(size int) ([]byte, error) {
	data := make([]byte, size)

	if _, err := rand.Read(data); err != nil {
		return nil, err
	}

	return data, nil
}

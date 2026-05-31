package ttpservice

// RegisteredEntity reprezentuje podmiot zarejestrowany w TTP.
//
// Struktura przechowuje identyfikator aplikacji, jej rolę, publiczne klucze RSA
// oraz certyfikat X.509 wydany przez TTP. Dane te są później wykorzystywane
// podczas walidacji tożsamości klienta lub serwera.
//
// @field ID Identyfikator zarejestrowanego podmiotu.
// @field Role Rola podmiotu w protokole, na przykład klient albo serwer.
// @field EncPublicKey Publiczny klucz RSA używany do szyfrowania danych dla podmiotu.
// @field AuthPublicKey Publiczny klucz RSA używany do weryfikacji podpisów podmiotu.
// @field Certificate Certyfikat X.509 wydany przez TTP.
type RegisteredEntity struct {
	ID            string
	Role          string
	EncPublicKey  string
	AuthPublicKey string
	Certificate   string
}

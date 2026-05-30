package ttpservice

type RegisteredEntity struct {
	ID            string
	Role          string
	EncPublicKey  string
	AuthPublicKey string
	Certificate   string
}

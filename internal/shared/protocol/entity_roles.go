package protocol

// EntityRole określa rolę aplikacji biorącej udział w protokole TTP.
//
// Typ jest używany do rozróżnienia, czy dana tożsamość należy do klienta,
// czy do serwera podczas rejestracji, certyfikacji i uwierzytelniania.
type EntityRole string

const (
	EntityRoleClient EntityRole = "CLIENT"
	EntityRoleServer EntityRole = "SERVER"
)

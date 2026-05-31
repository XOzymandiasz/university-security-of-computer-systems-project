// Package usecase zawiera przypadki użycia aplikacji klienta.
//
// Pakiet oddziela logikę aplikacyjną od warstwy HTTP i klientów komunikacyjnych.
package usecase

// MessageReader definiuje operację odczytu wiadomości z serwera.
//
// Interfejs pozwala przypadkowi użycia korzystać z dowolnej implementacji,
// która potrafi wysłać wiadomość do serwera i zwrócić jego odpowiedź.
type MessageReader interface {
	ReadMessage(msg string) (string, error)
}

// ReadMessage reprezentuje przypadek użycia wysłania wiadomości do serwera.
//
// Struktura opakowuje właściwy komponent komunikacyjny i udostępnia prostą
// metodę wykorzystywaną przez lokalne API HTTP klienta.
type ReadMessage struct {
	reader MessageReader
}

// NewReadMessage tworzy nowy przypadek użycia odczytu wiadomości.
//
// @param reader Komponent odpowiedzialny za komunikację z serwerem.
// @return Wskaźnik do nowej instancji ReadMessage.
func NewReadMessage(reader MessageReader) *ReadMessage {
	return &ReadMessage{
		reader: reader,
	}
}

// ReadMessage wysyła wiadomość do komponentu odpowiedzialnego za komunikację.
//
// Funkcja deleguje obsługę wiadomości do obiektu reader. W praktyce oznacza to
// zaszyfrowanie wiadomości, wysłanie jej do serwera i zwrócenie odszyfrowanej
// odpowiedzi.
//
// @param msg Jawna treść wiadomości wpisana przez użytkownika.
// @return Odpowiedź serwera lub błąd komunikacji.
func (r *ReadMessage) ReadMessage(msg string) (string, error) {
	return r.reader.ReadMessage(msg)
}

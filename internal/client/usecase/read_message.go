package usecase

type MessageReader interface {
	ReadMessage(msg string) (string, error)
}

type ReadMessage struct {
	reader MessageReader
}

func NewReadMessage(reader MessageReader) *ReadMessage {
	return &ReadMessage{
		reader: reader,
	}
}

func (r *ReadMessage) ReadMessage(msg string) (string, error) {
	return r.reader.ReadMessage(msg)
}

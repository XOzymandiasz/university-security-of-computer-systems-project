package usecase

type MessageReader interface {
	ReadMessage() (string, error)
}

type ReadMessage struct {
	reader MessageReader
}

func NewReadMessage(reader MessageReader) *ReadMessage {
	return &ReadMessage{
		reader: reader,
	}
}

func (r *ReadMessage) ReadMessage() (string, error) {
	return r.reader.ReadMessage()
}

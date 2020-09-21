package demo_gokit_kafka_grpc

type Message struct {
	ID string
	Msg string
	Username string
}

type MessageService interface {
	Send(msg Message) error
	Get(limit int) ([]Message, error)
}

// MessageRepository stores messages.
type MessageRepository interface {
	Save(msg Message) error
	Get(limit int) ([]Message, error)
}

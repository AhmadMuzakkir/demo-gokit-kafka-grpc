package inmem

import (
	"sync"

	demo "github.com/ahmadmuzakkir/demo-gokit-kafka-grpc"
)

type messageRepository struct {
	mu   sync.RWMutex
	list []demo.Message
}

func NewMessageRepository() demo.MessageRepository {
	return &messageRepository{}
}

func (c *messageRepository) Save(msg demo.Message) error {
	c.mu.Lock()
	c.list = append(c.list, msg)
	c.mu.Unlock()

	return nil
}

func (c *messageRepository) Get(limit int) ([]demo.Message, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	r := make([]demo.Message, limit)
	n := copy(r, c.list)
	if n < limit {
		r = r[:n]
	}

	return r, nil
}

type consumerRepository struct {
	mu     sync.RWMutex
	offset uint64
}

func NewConsumerRepository() demo.KafkaRepository {
	return &consumerRepository{}
}

func (c *consumerRepository) SaveOffset(partition, offset uint64) error {
	c.mu.Lock()
	c.offset = offset
	c.mu.Unlock()

	return nil
}

func (c *consumerRepository) GetOffset(partition uint64) (uint64, error) {
	c.mu.RLock()
	offset := c.offset
	c.mu.RUnlock()

	return offset, nil
}
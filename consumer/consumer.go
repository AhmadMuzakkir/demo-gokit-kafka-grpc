package consumer

import (
	"context"
	"fmt"
	"io"
	"time"

	demo "github.com/ahmadmuzakkir/demo-gokit-kafka-grpc"
	"github.com/ahmadmuzakkir/demo-gokit-kafka-grpc/proto"
	"github.com/segmentio/kafka-go"
	"go.uber.org/atomic"
)

type KafkaConsumerConfig struct {
	Addr      string
	Topic     string
	Partition uint64
}

type KafkaConsumer struct {
	addr      string
	topic     string
	partition uint64
	offset uint64

	r *kafka.Reader

	msgService   demo.MessageService
	consumerRepo demo.KafkaRepository

	hasStopped atomic.Bool
}

func NewKafkaConsumer(cfg KafkaConsumerConfig, msgService demo.MessageService, consumerRepo demo.KafkaRepository) *KafkaConsumer {
	return &KafkaConsumer{
		addr:         cfg.Addr,
		topic:        cfg.Topic,
		partition:    cfg.Partition,
		msgService:   msgService,
		consumerRepo: consumerRepo,
	}
}

// Start starts consuming the messages. This function blocks until Stop is called.
// You can call it in a goroutine.
func (k *KafkaConsumer) Start() error {
	k.r = kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{k.addr},
		Topic:     k.topic,
		Partition: int(k.partition),
		MinBytes:  10e3,
		MaxBytes:  10e6,
		MaxWait:   1 * time.Second,
	})

	// Get the offset

	offset, err := k.consumerRepo.GetOffset(k.partition)
	if err != nil {
		fmt.Printf("using 0 as offset, bcause failed get offset: %v\n", err)
		offset = 0
	} else {
		// The offset is the last message's offset.
		// Add 1 to ignore the last message.
		offset += 1
	}

	err = k.r.SetOffset(int64(offset))
	if err != nil {
		return fmt.Errorf("set offset: %w", err)
	}

	for {
		m, err := k.r.ReadMessage(context.Background())
		if err != nil {
			if err == io.EOF {
				if k.hasStopped.Load() {
					fmt.Println("consumer stopped")
					return nil
				} else {
					fmt.Println("error EOF read message. possible reason is kafka has stopped.")
				}
			} else {
				fmt.Println("error read message:", err)
			}

			return err
		}

		k.offset = uint64(m.Offset)

		data := proto.MessageData{}
		err = data.Unmarshal(m.Value)
		if err != nil {
			fmt.Printf("error unmarshal message at offset %d: %v\n", m.Offset, err)
		}

		err = k.msgService.Send(demo.Message{
			ID:       data.Id,
			Msg:      data.Msg,
			Username: data.Username,
		})

		if err != nil {
			fmt.Println("error save message:", err)
		}
	}
}

func (k *KafkaConsumer) Stop() error {
	if k.hasStopped.Load() {
		return nil
	}

	k.hasStopped.Store(true)

	var err error
	if k.r != nil {
		err = k.r.Close()
	}

	// Save the offset
	err = k.consumerRepo.SaveOffset(k.partition, k.offset)

	return err
}

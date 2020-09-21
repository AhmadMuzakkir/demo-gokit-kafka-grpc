package cmd

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)


// CreateTopic is a kafka helper func to create a topic.
func CreateTopic(kafkaAddr, topic string) error {
	fmt.Printf("Checking if topic %q exists, will create if neccesary...\n", topic)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	conn, err := kafka.DialContext(ctx, "tcp", kafkaAddr)
	if err != nil {
		return err
	}

	// Refresh brokers information.
	_, err = conn.ReadPartitions()
	if err != nil {
		return fmt.Errorf("read partitions: %w", err)
	}

	// Get the current controller.
	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("controller: %w", err)
	}
	conn.Close()

	ctrlConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return fmt.Errorf("dial controller %s:%d: %w", controller.Host, controller.Port, err)
	}

	err = ctrlConn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
	if err != nil {
		return fmt.Errorf("create topic: %w", err)
	}

	ctrlConn.Close()

	return nil
}

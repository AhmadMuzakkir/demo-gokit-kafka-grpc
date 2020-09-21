package demo_gokit_kafka_grpc

// KafkaRepository stores kafka consumer metadata.
type KafkaRepository interface {
	SaveOffset(partition, offset uint64) error
	GetOffset(partition uint64) (uint64, error)
}

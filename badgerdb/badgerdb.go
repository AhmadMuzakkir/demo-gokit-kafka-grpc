package badgerdb

import (
	"encoding/binary"
	"os"

	demo "github.com/ahmadmuzakkir/demo-gokit-kafka-grpc"
	"github.com/ahmadmuzakkir/demo-gokit-kafka-grpc/proto"
	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/options"
)

// Note: BadgerRepository uses message.pb to store the messages in the database.

var (
	keyMessage  = [1]byte{0x1}
	keyConsumer = [1]byte{0x2}
)

type BadgerRepository struct {
	db          *badger.DB
	messageRepo demo.MessageRepository
	kafkaRepo   demo.KafkaRepository
}

func NewBadgerRepository(path string) (*BadgerRepository, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, err
	}

	// Explicitly specify compression.
	// Because the default compression with CGO is zstd, and without CGO it's Snappy.
	db, err := badger.Open(badger.DefaultOptions(path).WithLogger(nullLog{}).WithCompression(options.Snappy))
	if err != nil {
		return nil, err
	}

	r := &BadgerRepository{
		messageRepo: &messageRepository{
			db:     db,
			prefix: keyMessage[:],
		},
		kafkaRepo: newConsumerRepository(db, keyConsumer[:]),
	}

	return r, nil
}

func (b *BadgerRepository) Close() error {
	return b.db.Close()
}

func (b *BadgerRepository) MessageRepository() demo.MessageRepository {
	return b.messageRepo
}

func (b *BadgerRepository) KafkaRepository() demo.KafkaRepository {
	return b.kafkaRepo
}

type messageRepository struct {
	db     *badger.DB
	prefix []byte
}

func (c *messageRepository) Save(msg demo.Message) error {
	key := make([]byte, 0, len(c.prefix)+len(msg.ID))
	key = append(key, c.prefix...)
	key = append(key, []byte(msg.ID)...)

	err := c.db.Update(func(txn *badger.Txn) error {
		p := &proto.MessageData{
			Id:       msg.ID,
			Msg:      msg.Msg,
			Username: msg.Username,
		}

		payload, err := p.Marshal()
		if err != nil {
			return err
		}

		return txn.Set(key, payload)
	})

	return err
}

func (c *messageRepository) Get(limit int) ([]demo.Message, error) {
	msgList := make([]demo.Message, 0, limit)

	err := c.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = limit

		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		count := 0

		for it.Seek(c.prefix); it.ValidForPrefix(c.prefix) && count < limit; it.Next() {
			count++

			item := it.Item()

			err := item.Value(func(v []byte) error {

				var data proto.MessageData

				err := data.Unmarshal(v)
				if err != nil {
					return err
				}

				msgList = append(msgList, demo.Message{
					ID:       data.Id,
					Msg:      data.Msg,
					Username: data.Username,
				})

				return nil
			})

			if err != nil {
				return err
			}
		}

		return nil
	})

	return msgList, err
}

type consumerRepository struct {
	db        *badger.DB
	offsetPrefix []byte
}

func newConsumerRepository(db *badger.DB, prefix []byte) *consumerRepository {
	offsetPrefix := make([]byte, 0, len(prefix)+8)
	offsetPrefix = append(offsetPrefix, prefix...)
	offsetPrefix = append(offsetPrefix, []byte("offset")...)

	r := &consumerRepository{
		db:        db,
		offsetPrefix: offsetPrefix,
	}

	return r
}

func (c *consumerRepository) SaveOffset(partition, offset uint64) error {
	key := c.getOffsetKey(partition)

	return c.db.Update(func(txn *badger.Txn) error {
		var b [8]byte

		binary.LittleEndian.PutUint64(b[:], offset)

		return txn.Set(key, b[:])
	})
}

func (c *consumerRepository) GetOffset(partition uint64) (uint64, error) {
	key := c.getOffsetKey(partition)

	var offset uint64

	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		v, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		offset = binary.LittleEndian.Uint64(v)

		return nil
	})

	return offset, err
}

func (c *consumerRepository) getOffsetKey(partition uint64) []byte {
	key := make([]byte, len(c.offsetPrefix) + 8)
	copy(key[:], c.offsetPrefix[:])

	binary.LittleEndian.PutUint64(key[len(c.offsetPrefix):], partition)

	return key
}

type nullLog struct{}

func (l nullLog) Errorf(f string, v ...interface{})   {}
func (l nullLog) Warningf(f string, v ...interface{}) {}
func (l nullLog) Infof(f string, v ...interface{})    {}
func (l nullLog) Debugf(f string, v ...interface{})   {}

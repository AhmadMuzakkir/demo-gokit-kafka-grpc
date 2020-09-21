package message

import (
	"time"

	demo "github.com/ahmadmuzakkir/demo-gokit-kafka-grpc"
	"github.com/go-kit/kit/log"
)

type loggingMiddleware struct {
	logger log.Logger
	next demo.MessageService
}

func (l loggingMiddleware) Send(msg demo.Message) error {
	var err error

	defer func(begin time.Time) {
		l.logger.Log(
			"method", "send",
			"id", msg.ID,
			"msg", msg.Msg,
			"username", msg.Username,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = l.next.Send(msg)

	return err
}

func (l loggingMiddleware) Get(limit int) ([]demo.Message, error) {
	var err error
	var messages []demo.Message

	defer func(begin time.Time) {
		l.logger.Log(
			"method", "send",
			"limit", limit,
			"response", messages,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	messages, err = l.next.Get(limit)

	return messages, err
}



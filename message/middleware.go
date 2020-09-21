package message

import (
	"time"

	demo "github.com/ahmadmuzakkir/demo-gokit-kafka-grpc"
	"github.com/go-kit/kit/log"
)

type LoggingMiddleware struct {
	logger log.Logger
	svc    demo.MessageService
}

func NewLoggingMiddleware(svc demo.MessageService, logger log.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
		svc:    svc,
	}
}

func (l *LoggingMiddleware) Send(msg demo.Message) error {
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

	err = l.svc.Send(msg)

	return err
}

func (l *LoggingMiddleware) Get(limit int) ([]demo.Message, error) {
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

	messages, err = l.svc.Get(limit)

	return messages, err
}

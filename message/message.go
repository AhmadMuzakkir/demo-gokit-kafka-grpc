package message

import (
	demo "github.com/ahmadmuzakkir/demo-gokit-kafka-grpc"
	"github.com/go-kit/kit/log"
)

type Service struct {
	repo demo.MessageRepository
}

func NewService(repo demo.MessageRepository, logger log.Logger) demo.MessageService {
	var service demo.MessageService

	service = &Service{repo: repo}

	// Wrap with logging
	service = loggingMiddleware{
		logger: logger,
		next:   service,
	}

	return service
}

func (s *Service) Send(msg demo.Message) error {
	return s.repo.Save(msg)
}

func (s *Service) Get(limit int) ([]demo.Message, error) {
	return s.repo.Get(limit)
}

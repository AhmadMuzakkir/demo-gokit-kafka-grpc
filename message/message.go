package message

import (
	demo "github.com/ahmadmuzakkir/demo-gokit-kafka-grpc"
)

type Service struct {
	repo demo.MessageRepository
}

func NewService(repo demo.MessageRepository) demo.MessageService {
	return &Service{repo: repo}
}

func (s *Service) Send(msg demo.Message) error {
	return s.repo.Save(msg)
}

func (s *Service) Get(limit int) ([]demo.Message, error) {
	return s.repo.Get(limit)
}

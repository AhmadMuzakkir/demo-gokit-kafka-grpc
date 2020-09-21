package message

import (
	"context"

	demo "github.com/ahmadmuzakkir/demo-gokit-kafka-grpc"
	"github.com/go-kit/kit/endpoint"
)

// message service endpoints.

type sendMessageRequest struct {
	id       string
	msg      string
	username string
}

type sendMessageResponse struct {
	Err string
}

func makeSendEndpoint(svc demo.MessageService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(sendMessageRequest)
		err := svc.Send(demo.Message{
			ID:       req.id,
			Msg:      req.msg,
			Username: req.username,
		})
		if err != nil {
			return sendMessageResponse{err.Error()}, nil
		}
		return sendMessageResponse{""}, nil
	}
}

type getMessageRequest struct {
	limit int
}

type getMessageResponse struct {
	err      error
	messages []demo.Message
}

func makeGetEndpoint(svc demo.MessageService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(getMessageRequest)
		messages, err := svc.Get(req.limit)
		if err != nil {
			return getMessageResponse{err: err}, nil
		}
		return getMessageResponse{messages: messages}, nil
	}
}

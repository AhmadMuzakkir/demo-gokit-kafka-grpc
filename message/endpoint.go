package message

import (
	"context"

	demo "github.com/ahmadmuzakkir/demo-gokit-kafka-grpc"
	"github.com/go-kit/kit/endpoint"
)

// message service endpoints.

type Endpoints struct {
	sendEndpoint endpoint.Endpoint
	getEndpoint  endpoint.Endpoint
}

func NewEndpoints(svc demo.MessageService) *Endpoints {
	return &Endpoints{
		sendEndpoint: MakeSendEndpoint(svc),
		getEndpoint:  MakeGetEndpoint(svc),
	}
}

func (e *Endpoints) Send(msg demo.Message) error {
	_, err := e.sendEndpoint(context.Background(), sendMessageRequest{
		id:       msg.ID,
		msg:      msg.Msg,
		username: msg.Username,
	})

	return err
}

func (e *Endpoints) Get(limit int) ([]demo.Message, error) {
	resp, err := e.getEndpoint(context.Background(), getMessageRequest{
		limit: limit,
	})
	if err != nil {
		return nil, err
	}

	res := resp.(getMessageResponse)
	return res.messages, res.err
}

type sendMessageRequest struct {
	id       string
	msg      string
	username string
}

type sendMessageResponse struct {
	Err string
}

func MakeSendEndpoint(svc demo.MessageService) endpoint.Endpoint {
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

func MakeGetEndpoint(svc demo.MessageService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(getMessageRequest)
		messages, err := svc.Get(req.limit)
		if err != nil {
			return getMessageResponse{err: err}, nil
		}
		return getMessageResponse{messages: messages}, nil
	}
}

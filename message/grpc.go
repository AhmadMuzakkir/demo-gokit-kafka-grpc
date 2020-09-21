package message

import (
	"context"

	demo "github.com/ahmadmuzakkir/demo-gokit-kafka-grpc"
	"github.com/ahmadmuzakkir/demo-gokit-kafka-grpc/proto"
	grpctransport "github.com/go-kit/kit/transport/grpc"
)

// gRPC server for message endpoints.

type grpcServer struct {
	send grpctransport.Handler
	get  grpctransport.Handler
}

func NewGRPCServer(messageService demo.MessageService) proto.MessageServer {
	return &grpcServer{
		send: grpctransport.NewServer(
			MakeSendEndpoint(messageService),
			decodeGRPCSendMessageRequest,
			encodeGRPCSendMessageResponse),
		get: grpctransport.NewServer(
			MakeGetEndpoint(messageService),
			decodeGRPCGetMessageRequest,
			encodeGRPCGetMessageResponse),
	}
}

func (g *grpcServer) Send(ctx context.Context, req *proto.MessageData) (*proto.SendMessageResponse, error) {
	_, res, err := g.send.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}

	return res.(*proto.SendMessageResponse), nil
}

func (g *grpcServer) Get(ctx context.Context, req *proto.GetMessageRequest) (*proto.GetMessageResponse, error) {
	_, res, err := g.get.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}

	return res.(*proto.GetMessageResponse), nil
}

func decodeGRPCSendMessageRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*proto.MessageData)
	return sendMessageRequest{
		id:       req.Id,
		msg:      req.Msg,
		username: req.Username,
	}, nil
}

func encodeGRPCSendMessageResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(sendMessageResponse)
	return &proto.SendMessageResponse{Err: resp.Err}, nil
}

func decodeGRPCGetMessageRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*proto.GetMessageRequest)
	return getMessageRequest{
		limit: int(req.Limit),
	}, nil
}

func encodeGRPCGetMessageResponse(_ context.Context, response interface{}) (interface{}, error) {
	res := response.(getMessageResponse)

	var m []*proto.MessageData

	for i := range res.messages {
		m = append(m, &proto.MessageData{
			Id:       res.messages[i].ID,
			Msg:      res.messages[i].Msg,
			Username: res.messages[i].Username,
		})
	}

	grpcRes := &proto.GetMessageResponse{}

	if res.err != nil {
		grpcRes.Err = res.err.Error()
	} else {
		grpcRes.Messages = m
	}

	return grpcRes, nil
}

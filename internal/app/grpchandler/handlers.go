package grpchandler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"alex-one/subscriber/internal/domain/models"
	proto "alex-one/subscriber/internal/pkg/api"
	"alex-one/subscriber/internal/pkg/logger"
)

type SubPub interface {
	Subscribe(subject string, cb models.MessageHandler) models.Subscription
	Publish(subject string, msg interface{}) error
}

type GRPCBublisherServer struct {
	proto.UnimplementedPubSubServer
	sp SubPub
}

func NewGRPCPublisherServer(sp SubPub) *GRPCBublisherServer {
	return &GRPCBublisherServer{
		sp: sp,
	}
}

func (g *GRPCBublisherServer) Publish(ctx context.Context, req *proto.PublishRequest) (*emptypb.Empty, error) {
	err := g.sp.Publish(req.Key, req.Data)
	if err != nil {
		logger.Errorf(ctx, "error  publish msg %v", err)
		return &emptypb.Empty{}, status.Error(codes.Internal, err.Error())

	}
	return &emptypb.Empty{}, status.Error(codes.OK, "Success")
}

func (g *GRPCBublisherServer) Subscribe(req *proto.SubscribeRequest, stream proto.PubSub_SubscribeServer) error {
	sub := g.sp.Subscribe(req.Key, func(msg interface{}) {
		if data, ok := msg.(string); ok {
			stream.Send(&proto.Event{Data: data})
		}
	})

	<-stream.Context().Done()
	sub.Unsubscribe()
	logger.Infof(context.Background(), "client completed")
	return nil
}

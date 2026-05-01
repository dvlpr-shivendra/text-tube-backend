package client

import (
	"context"

	pb "shared/proto"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type VideoClient struct {
	client pb.VideoServiceClient
	conn   *grpc.ClientConn
}

func NewVideoClient(addr string) (*VideoClient, error) {
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, err
	}

	return &VideoClient{
		client: pb.NewVideoServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *VideoClient) SearchChannel(ctx context.Context, req *pb.SearchChannelRequest) (*pb.SearchChannelResponse, error) {
	return c.client.SearchChannel(ctx, req)
}

func (c *VideoClient) GetChannelVideos(ctx context.Context, req *pb.GetChannelVideosRequest) (*pb.GetChannelVideosResponse, error) {
	return c.client.GetChannelVideos(ctx, req)
}

func (c *VideoClient) GetVideoDetails(ctx context.Context, req *pb.GetVideoDetailsRequest) (*pb.GetVideoDetailsResponse, error) {
	return c.client.GetVideoDetails(ctx, req)
}

func (c *VideoClient) GetVideoTranscript(ctx context.Context, req *pb.GetVideoTranscriptRequest) (*pb.GetVideoTranscriptResponse, error) {
	return c.client.GetVideoTranscript(ctx, req)
}

func (c *VideoClient) SummarizeVideo(ctx context.Context, req *pb.SummarizeVideoRequest) (*pb.SummarizeVideoResponse, error) {
	return c.client.SummarizeVideo(ctx, req)
}


func (c *VideoClient) Close() error {
	return c.conn.Close()
}

package client

import (
	"context"

	pb "shared/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthClient struct {
	client pb.AuthServiceClient
	conn   *grpc.ClientConn
}

func NewAuthClient(addr string) (*AuthClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &AuthClient{
		client: pb.NewAuthServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *AuthClient) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthSuccessResponse, error) {
	return c.client.Register(ctx, req)
}

func (c *AuthClient) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthSuccessResponse, error) {
	return c.client.Login(ctx, req)
}

func (c *AuthClient) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	return c.client.ValidateToken(ctx, req)
}

func (c *AuthClient) Close() error {
	return c.conn.Close()
}

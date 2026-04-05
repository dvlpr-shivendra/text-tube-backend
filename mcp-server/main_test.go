package main

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"google.golang.org/grpc"

	pb "shared/proto"
)

// MockVideoClient is a manual mock for pb.VideoServiceClient
type MockVideoClient struct {
	pb.VideoServiceClient
	SearchChannelFunc      func(ctx context.Context, in *pb.SearchChannelRequest, opts ...grpc.CallOption) (*pb.SearchChannelResponse, error)
	GetChannelVideosFunc   func(ctx context.Context, in *pb.GetChannelVideosRequest, opts ...grpc.CallOption) (*pb.GetChannelVideosResponse, error)
	GetVideoDetailsFunc    func(ctx context.Context, in *pb.GetVideoDetailsRequest, opts ...grpc.CallOption) (*pb.GetVideoDetailsResponse, error)
	GetVideoTranscriptFunc func(ctx context.Context, in *pb.GetVideoTranscriptRequest, opts ...grpc.CallOption) (*pb.GetVideoTranscriptResponse, error)
	SummarizeVideoFunc     func(ctx context.Context, in *pb.SummarizeVideoRequest, opts ...grpc.CallOption) (*pb.SummarizeVideoResponse, error)
}

func (m *MockVideoClient) SearchChannel(ctx context.Context, in *pb.SearchChannelRequest, opts ...grpc.CallOption) (*pb.SearchChannelResponse, error) {
	return m.SearchChannelFunc(ctx, in, opts...)
}

func (m *MockVideoClient) GetChannelVideos(ctx context.Context, in *pb.GetChannelVideosRequest, opts ...grpc.CallOption) (*pb.GetChannelVideosResponse, error) {
	return m.GetChannelVideosFunc(ctx, in, opts...)
}

func (m *MockVideoClient) GetVideoDetails(ctx context.Context, in *pb.GetVideoDetailsRequest, opts ...grpc.CallOption) (*pb.GetVideoDetailsResponse, error) {
	return m.GetVideoDetailsFunc(ctx, in, opts...)
}

func (m *MockVideoClient) GetVideoTranscript(ctx context.Context, in *pb.GetVideoTranscriptRequest, opts ...grpc.CallOption) (*pb.GetVideoTranscriptResponse, error) {
	return m.GetVideoTranscriptFunc(ctx, in, opts...)
}

func (m *MockVideoClient) SummarizeVideo(ctx context.Context, in *pb.SummarizeVideoRequest, opts ...grpc.CallOption) (*pb.SummarizeVideoResponse, error) {
	return m.SummarizeVideoFunc(ctx, in, opts...)
}

func TestSearchChannelTool(t *testing.T) {
	s := server.NewMCPServer("Test", "1.0.0")
	mock := &MockVideoClient{
		SearchChannelFunc: func(ctx context.Context, in *pb.SearchChannelRequest, opts ...grpc.CallOption) (*pb.SearchChannelResponse, error) {
			return &pb.SearchChannelResponse{
				ChannelId:    "UC123",
				ChannelTitle: "Test Channel",
				Videos:       []*pb.VideoInfo{},
			}, nil
		},
	}
	registerTools(s, mock)

	handler := s.GetTool("search_channel").Handler
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"channel_name": "test"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	if result.IsError {
		t.Fatalf("expected no error, got %v", result.Content)
	}

	found := false
	for _, content := range result.Content {
		if text, ok := mcp.AsTextContent(content); ok {
			if strings.Contains(text.Text, "Channel: Test Channel") {
				found = true
				break
			}
		}
	}
	if !found {
		t.Errorf("expected channel title in result, but got %+v", result.Content)
	}
}

func TestGetChannelVideosTool(t *testing.T) {
	s := server.NewMCPServer("Test", "1.0.0")
	mock := &MockVideoClient{
		GetChannelVideosFunc: func(ctx context.Context, in *pb.GetChannelVideosRequest, opts ...grpc.CallOption) (*pb.GetChannelVideosResponse, error) {
			if in.MaxResults != 10 {
				t.Errorf("expected default MaxResults 10, got %d", in.MaxResults)
			}
			return &pb.GetChannelVideosResponse{
				Videos: []*pb.VideoInfo{
					{VideoId: "v1", Title: "Vid 1"},
				},
			}, nil
		},
	}
	registerTools(s, mock)

	handler := s.GetTool("get_channel_videos").Handler
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"channel_id": "UC123"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, content := range result.Content {
		if text, ok := mcp.AsTextContent(content); ok {
			if strings.Contains(text.Text, "Vid 1") {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected video title in result")
	}
}

func TestGetVideoDetailsTool(t *testing.T) {
	s := server.NewMCPServer("Test", "1.0.0")
	mock := &MockVideoClient{
		GetVideoDetailsFunc: func(ctx context.Context, in *pb.GetVideoDetailsRequest, opts ...grpc.CallOption) (*pb.GetVideoDetailsResponse, error) {
			return &pb.GetVideoDetailsResponse{
				Video: &pb.VideoInfo{
					Title:        "Awesome Video",
					ChannelTitle: "Cool Channel",
					ViewCount:    1000,
				},
			}, nil
		},
	}
	registerTools(s, mock)

	handler := s.GetTool("get_video_details").Handler
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"video_id": "vid456"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, content := range result.Content {
		if text, ok := mcp.AsTextContent(content); ok {
			if strings.Contains(text.Text, "Title: Awesome Video") {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected video details in result")
	}
}

func TestGetVideoTranscriptTool(t *testing.T) {
	s := server.NewMCPServer("Test", "1.0.0")
	mock := &MockVideoClient{
		GetVideoTranscriptFunc: func(ctx context.Context, in *pb.GetVideoTranscriptRequest, opts ...grpc.CallOption) (*pb.GetVideoTranscriptResponse, error) {
			return &pb.GetVideoTranscriptResponse{
				Transcript: "Hello World",
			}, nil
		},
	}
	registerTools(s, mock)

	handler := s.GetTool("get_video_transcript").Handler
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"video_id": "vid123"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, content := range result.Content {
		if text, ok := mcp.AsTextContent(content); ok {
			if text.Text == "Hello World" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected transcript in result")
	}
}

func TestSummarizeVideoTool(t *testing.T) {
	s := server.NewMCPServer("Test", "1.0.0")
	mock := &MockVideoClient{
		SummarizeVideoFunc: func(ctx context.Context, in *pb.SummarizeVideoRequest, opts ...grpc.CallOption) (*pb.SummarizeVideoResponse, error) {
			return &pb.SummarizeVideoResponse{
				VideoId: in.VideoId,
				Summary: "This is a great summary.",
			}, nil
		},
	}
	registerTools(s, mock)

	handler := s.GetTool("summarize_video").Handler
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"video_id": "vid789"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, content := range result.Content {
		if text, ok := mcp.AsTextContent(content); ok {
			if strings.Contains(text.Text, "This is a great summary.") {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected summary in result")
	}
}

func TestToolMissingArgument(t *testing.T) {
	s := server.NewMCPServer("Test", "1.0.0")
	mock := &MockVideoClient{}
	registerTools(s, mock)

	handler := s.GetTool("search_channel").Handler
	req := mcp.CallToolRequest{} // no arguments

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	if !result.IsError {
		t.Error("expected tool to return error for missing argument")
	}
}

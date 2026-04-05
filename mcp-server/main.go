package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "shared/proto"
)

func main() {
	// Video service client
	videoServiceAddr := os.Getenv("VIDEO_SERVICE_ADDR")
	if videoServiceAddr == "" {
		videoServiceAddr = "localhost:50052"
	}

	conn, err := grpc.NewClient(videoServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	videoClient := pb.NewVideoServiceClient(conn)

	// Create MCP server
	s := server.NewMCPServer(
		"TextTube MCP Server",
		"1.0.0",
	)

	registerTools(s, videoClient)

	// Start standard I/O server
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func registerTools(s *server.MCPServer, videoClient pb.VideoServiceClient) {
	// Add tools
	// 1. Search Channel
	s.AddTool(mcp.NewTool("search_channel",
		mcp.WithDescription("Search for a YouTube channel by name and get its latest videos"),
		mcp.WithString("channel_name", mcp.Required(), mcp.Description("Name of the YouTube channel")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		channelName, err := request.RequireString("channel_name")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Missing argument: %v", err)), nil
		}

		resp, err := videoClient.SearchChannel(ctx, &pb.SearchChannelRequest{
			ChannelName: channelName,
			UserId:      "mcp-user",
		})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error searching channel: %v", err)), nil
		}

		return mcp.NewToolResultText(
			fmt.Sprintf("Channel: %s\nID: %s\nDescription: %s\nVideos found: %d",
				resp.ChannelTitle, resp.ChannelId, resp.ChannelDescription, len(resp.Videos)),
		), nil
	})

	// 2. Get Channel Videos
	s.AddTool(mcp.NewTool("get_channel_videos",
		mcp.WithDescription("Get the latest videos from a specific YouTube channel ID"),
		mcp.WithString("channel_id", mcp.Required(), mcp.Description("YouTube Channel ID")),
		mcp.WithNumber("max_results", mcp.Description("Maximum number of videos to fetch (default 10)")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		channelID, err := request.RequireString("channel_id")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Missing argument: %v", err)), nil
		}
		maxResults := int32(request.GetFloat("max_results", 10))

		resp, err := videoClient.GetChannelVideos(ctx, &pb.GetChannelVideosRequest{
			ChannelId:  channelID,
			MaxResults: maxResults,
			UserId:     "mcp-user",
		})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error getting channel videos: %v", err)), nil
		}

		var resultText string
		for _, v := range resp.Videos {
			resultText += fmt.Sprintf("- [%s] %s (Published: %s)\n", v.VideoId, v.Title, v.PublishedAt)
		}

		return mcp.NewToolResultText(resultText), nil
	})

	// 3. Get Video Details
	s.AddTool(mcp.NewTool("get_video_details",
		mcp.WithDescription("Get detailed information about a specific YouTube video"),
		mcp.WithString("video_id", mcp.Required(), mcp.Description("YouTube Video ID")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		videoID, err := request.RequireString("video_id")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Missing argument: %v", err)), nil
		}

		resp, err := videoClient.GetVideoDetails(ctx, &pb.GetVideoDetailsRequest{
			VideoId: videoID,
			UserId:  "mcp-user",
		})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error getting video details: %v", err)), nil
		}

		v := resp.Video
		resultText := fmt.Sprintf("Title: %s\nChannel: %s\nViews: %d\nLikes: %d\nPublished: %s\n\nDescription:\n%s",
			v.Title, v.ChannelTitle, v.ViewCount, v.LikeCount, v.PublishedAt, v.Description)

		return mcp.NewToolResultText(resultText), nil
	})

	// 4. Get Video Transcript
	s.AddTool(mcp.NewTool("get_video_transcript",
		mcp.WithDescription("Fetch the transcript for a given YouTube video"),
		mcp.WithString("video_id", mcp.Required(), mcp.Description("YouTube Video ID")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		videoID, err := request.RequireString("video_id")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Missing argument: %v", err)), nil
		}

		resp, err := videoClient.GetVideoTranscript(ctx, &pb.GetVideoTranscriptRequest{
			VideoId: videoID,
			UserId:  "mcp-user",
		})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error getting transcript: %v", err)), nil
		}

		return mcp.NewToolResultText(resp.Transcript), nil
	})

	// 5. Summarize Video
	s.AddTool(mcp.NewTool("summarize_video",
		mcp.WithDescription("Generate an AI summary for a YouTube video based on its transcript"),
		mcp.WithString("video_id", mcp.Required(), mcp.Description("YouTube Video ID")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		videoID, err := request.RequireString("video_id")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Missing argument: %v", err)), nil
		}

		resp, err := videoClient.SummarizeVideo(ctx, &pb.SummarizeVideoRequest{
			VideoId: videoID,
			UserId:  "mcp-user",
		})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error generating summary: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Summary for Video [%s]:\n\n%s", resp.VideoId, resp.Summary)), nil
	})
}

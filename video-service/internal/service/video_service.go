package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
	"videoservice/internal/client"
	"videoservice/internal/models"
	"videoservice/internal/repository"

	pb "shared/proto"
)

type VideoService struct {
	pb.UnimplementedVideoServiceServer
	videoRepo            *repository.VideoRepository
	youtubeClient        *client.YouTubeClient
	llmClient            LLMClient
	cacheMaxAge          time.Duration
	transcriptServiceURL string
}

func NewVideoService(videoRepo *repository.VideoRepository, youtubeClient *client.YouTubeClient, llmClient LLMClient) *VideoService {
	transcriptURL := os.Getenv("TRANSCRIPT_SERVICE_URL")
	if transcriptURL == "" {
		transcriptURL = "http://localhost:8081"
	}

	return &VideoService{
		videoRepo:            videoRepo,
		youtubeClient:        youtubeClient,
		llmClient:            llmClient,
		cacheMaxAge:          30 * time.Minute, // Cache for 30 minutes
		transcriptServiceURL: transcriptURL,
	}
}

func (s *VideoService) SearchChannel(ctx context.Context, req *pb.SearchChannelRequest) (*pb.SearchChannelResponse, error) {
	log.Printf("Searching channel: %s", req.ChannelName)
	// Try to get from cache first
	cachedChannel, err := s.videoRepo.GetCachedChannel(ctx, req.ChannelName)
	if err == nil && time.Since(cachedChannel.CachedAt) < s.cacheMaxAge {
		log.Printf("Cache hit for channel: %s", req.ChannelName)
		// Get cached videos
		videos, _ := s.videoRepo.GetCachedVideos(ctx, cachedChannel.ChannelID, s.cacheMaxAge)
		return s.buildSearchResponse(cachedChannel, videos), nil
	}

	log.Printf("Cache miss for channel: %s, searching YouTube", req.ChannelName)
	// Search YouTube
	channel, err := s.youtubeClient.SearchChannel(req.ChannelName)
	if err != nil {
		log.Printf("Error searching channel %s on YouTube: %v", req.ChannelName, err)
		return nil, err
	}

	// Get videos
	videos, err := s.youtubeClient.GetChannelVideos(channel.ChannelID, 10)
	if err != nil {
		log.Printf("Error getting videos for channel %s: %v", channel.ChannelID, err)
		return nil, err
	}

	// Cache channel and videos
	s.videoRepo.CacheChannel(ctx, channel)
	s.videoRepo.CacheVideos(ctx, videos)

	return s.buildSearchResponse(channel, videos), nil
}

func (s *VideoService) GetChannelVideos(ctx context.Context, req *pb.GetChannelVideosRequest) (*pb.GetChannelVideosResponse, error) {
	log.Printf("Getting videos for channel: %s", req.ChannelId)
	maxResults := req.MaxResults
	if maxResults <= 0 || maxResults > 50 {
		maxResults = 10
	}

	// Try cache first
	cachedVideos, err := s.videoRepo.GetCachedVideos(ctx, req.ChannelId, s.cacheMaxAge)
	if err == nil && len(cachedVideos) > 0 {
		log.Printf("Cache hit for videos of channel: %s", req.ChannelId)
		return &pb.GetChannelVideosResponse{
			Videos: s.convertVideosToProto(cachedVideos),
		}, nil
	}

	log.Printf("Cache miss for videos of channel: %s, fetching from YouTube", req.ChannelId)
	// Fetch from YouTube
	videos, err := s.youtubeClient.GetChannelVideos(req.ChannelId, int(maxResults))
	if err != nil {
		log.Printf("Error fetching videos for channel %s from YouTube: %v", req.ChannelId, err)
		return nil, err
	}

	// Cache videos
	s.videoRepo.CacheVideos(ctx, videos)

	return &pb.GetChannelVideosResponse{
		Videos: s.convertVideosToProto(videos),
	}, nil
}

func (s *VideoService) GetVideoDetails(ctx context.Context, req *pb.GetVideoDetailsRequest) (*pb.GetVideoDetailsResponse, error) {
	log.Printf("Getting video details for: %s", req.VideoId)
	// Try cache first
	cachedVideo, err := s.videoRepo.GetCachedVideo(ctx, req.VideoId, s.cacheMaxAge)
	if err == nil {
		log.Printf("Cache hit for video details: %s", req.VideoId)
		return &pb.GetVideoDetailsResponse{
			Video: s.convertVideoToProto(cachedVideo),
		}, nil
	}

	log.Printf("Cache miss for video details: %s, fetching from YouTube", req.VideoId)
	// Fetch from YouTube
	video, err := s.youtubeClient.GetVideoDetails(req.VideoId)
	if err != nil {
		log.Printf("Error fetching video details for %s from YouTube: %v", req.VideoId, err)
		return nil, err
	}

	// Cache video
	s.videoRepo.CacheVideos(ctx, []models.Video{*video})

	return &pb.GetVideoDetailsResponse{
		Video: s.convertVideoToProto(video),
	}, nil
}

func (s *VideoService) GetVideoTranscript(ctx context.Context, req *pb.GetVideoTranscriptRequest) (*pb.GetVideoTranscriptResponse, error) {
	log.Printf("Getting transcript for video: %s", req.VideoId)
	// Validate video ID format
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]{11}$`, req.VideoId)
	if !matched {
		log.Printf("Invalid video ID format: %s", req.VideoId)
		return nil, fmt.Errorf("invalid video id")
	}

	transcriptURL := fmt.Sprintf("%s/transcript?videoId=%s", s.transcriptServiceURL, req.VideoId)
	log.Printf("Fetching transcript from: %s", transcriptURL)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, transcriptURL, nil)
	if err != nil {
		log.Printf("Failed to create transcript request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		log.Printf("Error calling transcript service: %v", err)
		return nil, fmt.Errorf("failed to fetch transcript: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Transcript string `json:"transcript"`
		Error      string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Failed to decode transcript response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	if result.Error != "" {
		log.Printf("Transcript service returned error: %s", result.Error)
		return nil, fmt.Errorf("transcript service error: %s", result.Error)
	}

	log.Printf("Successfully fetched transcript for video: %s", req.VideoId)

	return &pb.GetVideoTranscriptResponse{
		Transcript: result.Transcript,
		VideoId:    req.VideoId,
	}, nil
}

func (s *VideoService) SummarizeVideo(ctx context.Context, req *pb.SummarizeVideoRequest) (*pb.SummarizeVideoResponse, error) {
	log.Printf("Summarizing video: %s for user: %s", req.VideoId, req.UserId)
	// First, fetch the transcript
	transcriptResp, err := s.GetVideoTranscript(ctx, &pb.GetVideoTranscriptRequest{
		VideoId: req.VideoId,
		UserId:  req.UserId,
	})
	if err != nil {
		log.Printf("Error getting transcript for summarization of %s: %v", req.VideoId, err)
		return nil, fmt.Errorf("failed to fetch transcript for summarization: %w", err)
	}

	transcriptResp.Transcript = "Hello, this is a test transcript."

	if transcriptResp.Transcript == "" {
		log.Printf("Empty transcript for video %s, cannot summarize", req.VideoId)
		return nil, fmt.Errorf("transcript is empty, cannot generate summary")
	}

	// Then, call LLM to summarize
	log.Printf("Calling LLM to summarize video: %s", req.VideoId)
	summary, err := s.llmClient.Summarize(ctx, transcriptResp.Transcript)
	if err != nil {
		log.Printf("Error summarizing video %s with LLM: %v", req.VideoId, err)
		return nil, fmt.Errorf("failed to generate summary: %w", err)
	}

	log.Printf("Successfully summarized video: %s", req.VideoId)

	return &pb.SummarizeVideoResponse{
		Summary: summary,
		VideoId: req.VideoId,
	}, nil
}

func cleanWhitespace(s string) string {
	// Replace multiple whitespace characters with a single space
	space := regexp.MustCompile(`\s+`)
	s = space.ReplaceAllString(s, " ")
	// Trim leading and trailing whitespace
	return strings.TrimSpace(s)
}

func (s *VideoService) buildSearchResponse(channel *models.Channel, videos []models.Video) *pb.SearchChannelResponse {
	return &pb.SearchChannelResponse{
		ChannelId:          channel.ChannelID,
		ChannelTitle:       channel.Title,
		ChannelDescription: channel.Description,
		ThumbnailUrl:       channel.Thumbnail,
		Videos:             s.convertVideosToProto(videos),
	}
}

func (s *VideoService) convertVideosToProto(videos []models.Video) []*pb.VideoInfo {
	result := make([]*pb.VideoInfo, 0, len(videos))
	for _, v := range videos {
		result = append(result, s.convertVideoToProto(&v))
	}
	return result
}

func (s *VideoService) convertVideoToProto(video *models.Video) *pb.VideoInfo {
	return &pb.VideoInfo{
		VideoId:      video.VideoID,
		Title:        video.Title,
		Description:  video.Description,
		ThumbnailUrl: video.Thumbnail,
		PublishedAt:  video.PublishedAt,
		ChannelId:    video.ChannelID,
		ChannelTitle: video.ChannelTitle,
		ViewCount:    video.ViewCount,
		LikeCount:    video.LikeCount,
	}
}

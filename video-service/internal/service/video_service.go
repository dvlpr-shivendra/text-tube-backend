package service

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
	"videoservice/internal/client"
	"videoservice/internal/models"
	"videoservice/internal/repository"

	pb "shared/proto"

	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/html"
)

type VideoService struct {
	pb.UnimplementedVideoServiceServer
	videoRepo     *repository.VideoRepository
	youtubeClient *client.YouTubeClient
	cacheMaxAge   time.Duration
}

func NewVideoService(videoRepo *repository.VideoRepository, youtubeClient *client.YouTubeClient) *VideoService {
	return &VideoService{
		videoRepo:     videoRepo,
		youtubeClient: youtubeClient,
		cacheMaxAge:   30 * time.Minute, // Cache for 30 minutes
	}
}

func (s *VideoService) SearchChannel(ctx context.Context, req *pb.SearchChannelRequest) (*pb.SearchChannelResponse, error) {
	// Try to get from cache first
	cachedChannel, err := s.videoRepo.GetCachedChannel(ctx, req.ChannelName)
	if err == nil && time.Since(cachedChannel.CachedAt) < s.cacheMaxAge {
		// Get cached videos
		videos, _ := s.videoRepo.GetCachedVideos(ctx, cachedChannel.ChannelID, s.cacheMaxAge)
		return s.buildSearchResponse(cachedChannel, videos), nil
	}

	// Search YouTube
	channel, err := s.youtubeClient.SearchChannel(req.ChannelName)
	if err != nil {
		return nil, err
	}

	// Get videos
	videos, err := s.youtubeClient.GetChannelVideos(channel.ChannelID, 10)
	if err != nil {
		return nil, err
	}

	// Cache channel and videos
	s.videoRepo.CacheChannel(ctx, channel)
	s.videoRepo.CacheVideos(ctx, videos)

	return s.buildSearchResponse(channel, videos), nil
}

func (s *VideoService) GetChannelVideos(ctx context.Context, req *pb.GetChannelVideosRequest) (*pb.GetChannelVideosResponse, error) {
	maxResults := req.MaxResults
	if maxResults <= 0 || maxResults > 50 {
		maxResults = 10
	}

	// Try cache first
	cachedVideos, err := s.videoRepo.GetCachedVideos(ctx, req.ChannelId, s.cacheMaxAge)
	if err == nil && len(cachedVideos) > 0 {
		return &pb.GetChannelVideosResponse{
			Videos: s.convertVideosToProto(cachedVideos),
		}, nil
	}

	// Fetch from YouTube
	videos, err := s.youtubeClient.GetChannelVideos(req.ChannelId, int(maxResults))
	if err != nil {
		return nil, err
	}

	// Cache videos
	s.videoRepo.CacheVideos(ctx, videos)

	return &pb.GetChannelVideosResponse{
		Videos: s.convertVideosToProto(videos),
	}, nil
}

func (s *VideoService) GetVideoDetails(ctx context.Context, req *pb.GetVideoDetailsRequest) (*pb.GetVideoDetailsResponse, error) {
	// Try cache first
	cachedVideo, err := s.videoRepo.GetCachedVideo(ctx, req.VideoId, s.cacheMaxAge)
	if err == nil {
		return &pb.GetVideoDetailsResponse{
			Video: s.convertVideoToProto(cachedVideo),
		}, nil
	}

	// Fetch from YouTube
	video, err := s.youtubeClient.GetVideoDetails(req.VideoId)
	if err != nil {
		return nil, err
	}

	// Cache video
	s.videoRepo.CacheVideos(ctx, []models.Video{*video})

	return &pb.GetVideoDetailsResponse{
		Video: s.convertVideoToProto(video),
	}, nil
}

func (s *VideoService) GetVideoTranscript(ctx context.Context, req *pb.GetVideoTranscriptRequest) (*pb.GetVideoTranscriptResponse, error) {
	transcriptURL := fmt.Sprintf("https://youtubetotranscript.com/transcript?v=%s", req.VideoId)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", transcriptURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch transcript: status %d", resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	transcript := extractTranscript(doc)
	transcript = cleanWhitespace(transcript)

	return &pb.GetVideoTranscriptResponse{
		Transcript: transcript,
		VideoId:    req.VideoId,
	}, nil
}

func extractTranscript(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "div" {
		for _, attr := range n.Attr {
			if attr.Key == "id" && attr.Val == "transcript" {
				return getTextContent(n)
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := extractTranscript(c); result != "" {
			return result
		}
	}

	return ""
}

func getTextContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}

	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += getTextContent(c)
	}
	return text
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

// Handle cache miss as not found
func isCacheMiss(err error) bool {
	return err == mongo.ErrNoDocuments
}

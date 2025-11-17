package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"videoservice/internal/models"
)

type YouTubeClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewYouTubeClient(apiKey string) *YouTubeClient {
	return &YouTubeClient{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

type SearchResponse struct {
	Items []struct {
		ID struct {
			ChannelID string `json:"channelId"`
		} `json:"id"`
		Snippet struct {
			ChannelID   string `json:"channelId"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Thumbnails  struct {
				Default struct {
					URL string `json:"url"`
				} `json:"default"`
			} `json:"thumbnails"`
		} `json:"snippet"`
	} `json:"items"`
}

type VideosResponse struct {
	Items []struct {
		ID struct {
			VideoID string `json:"videoId"`
		} `json:"id"`
		Snippet struct {
			PublishedAt  string `json:"publishedAt"`
			ChannelID    string `json:"channelId"`
			Title        string `json:"title"`
			Description  string `json:"description"`
			ChannelTitle string `json:"channelTitle"`
			Thumbnails   struct {
				Default struct {
					URL string `json:"url"`
				} `json:"default"`
			} `json:"thumbnails"`
		} `json:"snippet"`
	} `json:"items"`
}

type VideoDetailsResponse struct {
	Items []struct {
		ID      string `json:"id"`
		Snippet struct {
			PublishedAt  string `json:"publishedAt"`
			ChannelID    string `json:"channelId"`
			Title        string `json:"title"`
			Description  string `json:"description"`
			ChannelTitle string `json:"channelTitle"`
			Thumbnails   struct {
				Default struct {
					URL string `json:"url"`
				} `json:"default"`
			} `json:"thumbnails"`
		} `json:"snippet"`
		Statistics struct {
			ViewCount string `json:"viewCount"`
			LikeCount string `json:"likeCount"`
		} `json:"statistics"`
	} `json:"items"`
}

type ChannelResponse struct {
	Items []struct {
		ID      string `json:"id"`
		Snippet struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Thumbnails  struct {
				Default struct {
					URL string `json:"url"`
				} `json:"default"`
			} `json:"thumbnails"`
		} `json:"snippet"`
	} `json:"items"`
}

type CaptionsListResponse struct {
	Items []struct {
		ID      string `json:"id"`
		Snippet struct {
			VideoID  string `json:"videoId"`
			Language string `json:"language"`
			Name     string `json:"name"`
		} `json:"snippet"`
	} `json:"items"`
}

func (c *YouTubeClient) GetChannelByHandle(handle string) (*models.Channel, error) {
	apiURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/channels?part=snippet&forHandle=%s&key=%s", handle, c.apiKey)

	resp, err := c.httpClient.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("YouTube API error: %s - %s", resp.Status, string(body))
	}

	var channelResp ChannelResponse
	if err := json.NewDecoder(resp.Body).Decode(&channelResp); err != nil {
		return nil, err
	}

	if len(channelResp.Items) == 0 {
		return nil, fmt.Errorf("channel not found")
	}

	item := channelResp.Items[0]
	return &models.Channel{
		ChannelID:   item.ID,
		Title:       item.Snippet.Title,
		Description: item.Snippet.Description,
		Thumbnail:   item.Snippet.Thumbnails.Default.URL,
	}, nil
}

func (c *YouTubeClient) SearchChannel(channelName string) (*models.Channel, error) {
	// Check if channelName is a URL like https://www.youtube.com/@VeronicaExplains
	if strings.HasPrefix(channelName, "http://") || strings.HasPrefix(channelName, "https://") {
		u, err := url.Parse(channelName)
		if err == nil {
			path := strings.TrimPrefix(u.Path, "/@")
			if path != "" {
				// Treat as a handle (e.g., @VeronicaExplains)
				channelHandle := strings.TrimPrefix(path, "@")
				return c.GetChannelByHandle(channelHandle)
			}
		}
	}

	searchURL := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/search?key=%s&q=%s&type=channel&part=snippet&maxResults=1",
		c.apiKey,
		url.QueryEscape(channelName),
	)

	resp, err := c.httpClient.Get(searchURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("YouTube API error: %s - %s", resp.Status, string(body))
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	if len(searchResp.Items) == 0 {
		return nil, fmt.Errorf("channel not found")
	}

	item := searchResp.Items[0]
	return &models.Channel{
		ChannelID:   item.Snippet.ChannelID,
		Title:       item.Snippet.Title,
		Description: item.Snippet.Description,
		Thumbnail:   item.Snippet.Thumbnails.Default.URL,
	}, nil
}

func (c *YouTubeClient) GetChannelVideos(channelID string, maxResults int) ([]models.Video, error) {
	videosURL := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/search?key=%s&channelId=%s&part=snippet,id&order=date&maxResults=%d&type=video",
		c.apiKey,
		channelID,
		maxResults,
	)

	resp, err := c.httpClient.Get(videosURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("YouTube API error: %s - %s", resp.Status, string(body))
	}

	var videosResp VideosResponse
	if err := json.NewDecoder(resp.Body).Decode(&videosResp); err != nil {
		return nil, err
	}

	videos := make([]models.Video, 0, len(videosResp.Items))
	for _, item := range videosResp.Items {
		if item.ID.VideoID != "" {
			videos = append(videos, models.Video{
				VideoID:      item.ID.VideoID,
				Title:        item.Snippet.Title,
				Description:  item.Snippet.Description,
				Thumbnail:    item.Snippet.Thumbnails.Default.URL,
				PublishedAt:  item.Snippet.PublishedAt,
				ChannelID:    item.Snippet.ChannelID,
				ChannelTitle: item.Snippet.ChannelTitle,
			})
		}
	}

	return videos, nil
}

func (c *YouTubeClient) GetVideoDetails(videoID string) (*models.Video, error) {
	detailsURL := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/videos?key=%s&id=%s&part=snippet,statistics",
		c.apiKey,
		videoID,
	)

	resp, err := c.httpClient.Get(detailsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("YouTube API error: %s - %s", resp.Status, string(body))
	}

	var detailsResp VideoDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&detailsResp); err != nil {
		return nil, err
	}

	if len(detailsResp.Items) == 0 {
		return nil, fmt.Errorf("video not found")
	}

	item := detailsResp.Items[0]

	var viewCount, likeCount int64
	fmt.Sscanf(item.Statistics.ViewCount, "%d", &viewCount)
	fmt.Sscanf(item.Statistics.LikeCount, "%d", &likeCount)

	return &models.Video{
		VideoID:      item.ID,
		Title:        item.Snippet.Title,
		Description:  item.Snippet.Description,
		Thumbnail:    item.Snippet.Thumbnails.Default.URL,
		PublishedAt:  item.Snippet.PublishedAt,
		ChannelID:    item.Snippet.ChannelID,
		ChannelTitle: item.Snippet.ChannelTitle,
		ViewCount:    viewCount,
		LikeCount:    likeCount,
	}, nil
}

func (c *YouTubeClient) GetVideoTranscript(videoID string) (string, error) {
	// First, list available captions
	captionsURL := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/captions?key=%s&videoId=%s&part=snippet",
		c.apiKey,
		videoID,
	)

	resp, err := c.httpClient.Get(captionsURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("YouTube API error: %s - %s", resp.Status, string(body))
	}

	var captionsResp CaptionsListResponse
	if err := json.NewDecoder(resp.Body).Decode(&captionsResp); err != nil {
		return "", err
	}

	if len(captionsResp.Items) == 0 {
		return "", fmt.Errorf("no captions found for video")
	}

	// Get the first available caption track
	captionID := captionsResp.Items[0].ID

	// Download the caption track
	downloadURL := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/captions/%s?key=%s&tfmt=srt",
		captionID,
		c.apiKey,
	)

	resp, err = c.httpClient.Get(downloadURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("YouTube API error: %s - %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse SRT format and extract text
	transcript := parseSRTTranscript(string(body))
	return transcript, nil
}

func parseSRTTranscript(srt string) string {
	lines := strings.Split(srt, "\n")
	var transcript []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip empty lines, sequence numbers, and timecodes
		if trimmed == "" || isSequenceNumber(trimmed) || isTimecode(trimmed) {
			continue
		}
		if trimmed != "" {
			transcript = append(transcript, trimmed)
		}
	}

	return strings.Join(transcript, " ")
}

func isSequenceNumber(s string) bool {
	// Check if string is just a number
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return s != ""
}

func isTimecode(s string) bool {
	// Check if string looks like "00:00:00,000 --> 00:00:05,000"
	return strings.Contains(s, "-->")
}

package handler

import (
	"encoding/json"
	"gateway/internal/client"
	"log"
	"net/http"
	"strconv"

	pb "shared/proto"

	"github.com/gorilla/mux"
)

type VideoHandler struct {
	videoClient *client.VideoClient
}

func NewVideoHandler(videoClient *client.VideoClient) *VideoHandler {
	return &VideoHandler{videoClient: videoClient}
}

type VideoThumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type VideoSummary struct {
	VideoID      string         `json:"video_id"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	ThumbnailURL string         `json:"thumbnail_url"`
	PublishedAt  string         `json:"published_at"`
	ChannelID    string         `json:"channel_id"`
	ChannelTitle string         `json:"channel_title"`
}

type SearchChannelResponse struct {
	ChannelID          string         `json:"channel_id"`
	ChannelTitle       string         `json:"channel_title"`
	ChannelDescription string         `json:"channel_description"`
	ThumbnailURL       string         `json:"thumbnail_url"`
	Videos             []VideoSummary `json:"videos"`
	NextPageToken      string         `json:"next_page_token"`
}

type GetChannelVideosResponse struct {
	Videos        []VideoSummary `json:"videos"`
	NextPageToken string         `json:"next_page_token"`
}

type VideoDetailsResponse struct {
	Video VideoSummary `json:"video"`
}

type TranscriptLine struct {
	Text      string  `json:"text"`
	StartTime float64 `json:"start_time"`
	Duration  float64 `json:"duration"`
}

type TranscriptResponse struct {
	VideoID    string           `json:"video_id"`
	Transcript []TranscriptLine `json:"transcript"`
}

type SummarizeResponse struct {
	VideoID string `json:"video_id"`
	Summary string `json:"summary"`
}

func (h *VideoHandler) sendJSONError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// SearchChannel godoc
// @Summary Search for a YouTube channel
// @Description Search for a channel by name and return its details and recent videos
// @Tags videos
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param channel query string true "Channel Name"
// @Success 200 {object} SearchChannelResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/videos/search [get]
func (h *VideoHandler) SearchChannel(w http.ResponseWriter, r *http.Request) {
	channelName := r.URL.Query().Get("channel")
	if channelName == "" {
		h.sendJSONError(w, "channel parameter is required", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(string)

	resp, err := h.videoClient.SearchChannel(r.Context(), &pb.SearchChannelRequest{
		ChannelName: channelName,
		UserId:      userID,
	})
	if err != nil {
		log.Printf("SearchChannel failure: %v", err)
		h.sendJSONError(w, "Failed to search channel", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetChannelVideos godoc
// @Summary Get videos from a channel
// @Description Get a list of videos from a specific channel ID
// @Tags videos
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param channelId path string true "Channel ID"
// @Param max_results query int false "Max Results" default(10)
// @Param page_token query string false "Page Token"
// @Success 200 {object} GetChannelVideosResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/videos/channel/{channelId} [get]
func (h *VideoHandler) GetChannelVideos(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID := vars["channelId"]

	userID := r.Context().Value("user_id").(string)

	maxResults := int32(10)
	if mr := r.URL.Query().Get("max_results"); mr != "" {
		if val, err := strconv.Atoi(mr); err == nil {
			maxResults = int32(val)
		}
	}

	pageToken := r.URL.Query().Get("page_token")

	resp, err := h.videoClient.GetChannelVideos(r.Context(), &pb.GetChannelVideosRequest{
		ChannelId:  channelID,
		UserId:     userID,
		MaxResults: maxResults,
		PageToken:  pageToken,
	})
	if err != nil {
		log.Printf("GetChannelVideos failure: %v", err)
		h.sendJSONError(w, "Failed to get channel videos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetVideoDetails godoc
// @Summary Get video details
// @Description Get detailed information about a specific video
// @Tags videos
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param videoId path string true "Video ID"
// @Success 200 {object} VideoDetailsResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/videos/{videoId} [get]
func (h *VideoHandler) GetVideoDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	videoID := vars["videoId"]

	userID := r.Context().Value("user_id").(string)

	resp, err := h.videoClient.GetVideoDetails(r.Context(), &pb.GetVideoDetailsRequest{
		VideoId: videoID,
		UserId:  userID,
	})
	if err != nil {
		log.Printf("GetVideoDetails failure: %v", err)
		h.sendJSONError(w, "Failed to get video details", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetVideoTranscript godoc
// @Summary Get video transcript
// @Description Get the transcript of a specific video
// @Tags videos
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param videoId path string true "Video ID"
// @Success 200 {object} TranscriptResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/videos/{videoId}/transcript [get]
func (h *VideoHandler) GetVideoTranscript(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	videoID := vars["videoId"]

	userID := r.Context().Value("user_id").(string)

	resp, err := h.videoClient.GetVideoTranscript(r.Context(), &pb.GetVideoTranscriptRequest{
		VideoId: videoID,
		UserId:  userID,
	})
	if err != nil {
		log.Printf("GetVideoTranscript failure: %v", err)
		h.sendJSONError(w, "Failed to get video transcript", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// SummarizeVideo godoc
// @Summary Summarize a video
// @Description Generate a summary for a specific video
// @Tags videos
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param videoId path string true "Video ID"
// @Success 200 {object} SummarizeResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/videos/{videoId}/summarize [get]
func (h *VideoHandler) SummarizeVideo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	videoID := vars["videoId"]

	userID := r.Context().Value("user_id").(string)

	resp, err := h.videoClient.SummarizeVideo(r.Context(), &pb.SummarizeVideoRequest{
		VideoId: videoID,
		UserId:  userID,
	})
	if err != nil {
		log.Printf("SummarizeVideo failure: %v", err)
		h.sendJSONError(w, "Failed to summarize video", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}


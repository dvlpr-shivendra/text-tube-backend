package handler

import (
	"encoding/json"
	"gateway/internal/client"
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

func (h *VideoHandler) SearchChannel(w http.ResponseWriter, r *http.Request) {
	channelName := r.URL.Query().Get("channel")
	if channelName == "" {
		http.Error(w, "channel parameter is required", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(string)

	resp, err := h.videoClient.SearchChannel(r.Context(), &pb.SearchChannelRequest{
		ChannelName: channelName,
		UserId:      userID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

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

	resp, err := h.videoClient.GetChannelVideos(r.Context(), &pb.GetChannelVideosRequest{
		ChannelId:  channelID,
		UserId:     userID,
		MaxResults: maxResults,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *VideoHandler) GetVideoDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	videoID := vars["videoId"]

	userID := r.Context().Value("user_id").(string)

	resp, err := h.videoClient.GetVideoDetails(r.Context(), &pb.GetVideoDetailsRequest{
		VideoId: videoID,
		UserId:  userID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *VideoHandler) GetVideoTranscript(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	videoID := vars["videoId"]

	userID := r.Context().Value("user_id").(string)

	resp, err := h.videoClient.GetVideoTranscript(r.Context(), &pb.GetVideoTranscriptRequest{
		VideoId: videoID,
		UserId:  userID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

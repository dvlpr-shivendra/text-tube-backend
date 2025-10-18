package main

import (
	"gateway/internal/client"
	"gateway/internal/handler"
	"gateway/internal/middleware"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = "8080"
	}

	authServiceAddr := os.Getenv("AUTH_SERVICE_ADDR")
	if authServiceAddr == "" {
		authServiceAddr = "localhost:50051"
	}

	videoServiceAddr := os.Getenv("VIDEO_SERVICE_ADDR")
	if videoServiceAddr == "" {
		videoServiceAddr = "localhost:50052"
	}

	authClient, err := client.NewAuthClient(authServiceAddr)
	if err != nil {
		log.Fatalf("Failed to connect to auth service: %v", err)
	}
	defer authClient.Close()

	videoClient, err := client.NewVideoClient(videoServiceAddr)
	if err != nil {
		log.Fatalf("Failed to connect to video service: %v", err)
	}
	defer videoClient.Close()

	h := handler.NewHandler(authClient)
	vh := handler.NewVideoHandler(videoClient)
	m := middleware.NewMiddleware(authClient)

	r := mux.NewRouter()

	// Public routes
	r.HandleFunc("/health", h.HealthCheck).Methods("GET")
	r.HandleFunc("/api/auth/register", h.Register).Methods("POST")
	r.HandleFunc("/api/auth/login", h.Login).Methods("POST")

	// Protected routes
	protected := r.PathPrefix("/api").Subrouter()
	protected.Use(m.AuthMiddleware)
	protected.HandleFunc("/profile", h.GetProfile).Methods("GET")

	// Video routes (protected)
	protected.HandleFunc("/videos/search", vh.SearchChannel).Methods("GET")
	protected.HandleFunc("/videos/channel/{channelId}", vh.GetChannelVideos).Methods("GET")
	protected.HandleFunc("/videos/{videoId}", vh.GetVideoDetails).Methods("GET")
	protected.HandleFunc("/videos/{videoId}/transcript", vh.GetVideoTranscript).Methods("GET") // NEW

	log.Printf("🚀 Gateway starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

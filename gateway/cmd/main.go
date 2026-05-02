package main

import (
	"gateway/internal/client"
	"gateway/internal/handler"
	"gateway/internal/middleware"
	"log"
	"net/http"
	"os"
	"context"
	"time"

	"shared/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"github.com/gorilla/mux"
	_ "gateway/docs" // This is where swag will generate the docs
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title TextTube Gateway API
// @version 1.0
// @description This is the API gateway for the TextTube project.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @query.collection.format multi

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = "8080"
	}

	collectorAddr := os.Getenv("OTEL_COLLECTOR_ADDR")
	if collectorAddr == "" {
		collectorAddr = "localhost:4317"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize Telemetry
	shutdown, err := telemetry.InitTelemetry(ctx, "gateway", collectorAddr)
	if err != nil {
		log.Printf("Failed to initialize telemetry: %v", err)
	} else {
		defer shutdown()
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
	ssrh := handler.NewSSRHandler(authClient, videoClient)
	m := middleware.NewMiddleware(authClient)

	r := mux.NewRouter()

	// Public routes
	r.HandleFunc("/health", h.HealthCheck).Methods("GET")
	r.HandleFunc("/api/auth/register", h.Register).Methods("POST")
	r.HandleFunc("/api/auth/login", h.Login).Methods("POST")

	// Swagger UI
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// SSR Public routes
	r.HandleFunc("/login", ssrh.ShowLogin).Methods("GET")
	r.HandleFunc("/login", ssrh.Login).Methods("POST")
	r.HandleFunc("/register", ssrh.ShowRegister).Methods("GET")
	r.HandleFunc("/register", ssrh.Register).Methods("POST")
	r.HandleFunc("/logout", ssrh.Logout).Methods("GET")

	// SSR Protected routes
	ssr := r.PathPrefix("/").Subrouter()
	ssr.Use(m.SSRAuthMiddleware)
	ssr.HandleFunc("/", ssrh.Home).Methods("GET")
	ssr.HandleFunc("/video/{videoId}", ssrh.VideoDetail).Methods("GET")
	ssr.HandleFunc("/video/{videoId}/summarize", ssrh.Summarize).Methods("POST")

	// Protected JSON routes
	protected := r.PathPrefix("/api").Subrouter()
	protected.Use(m.AuthMiddleware)
	protected.HandleFunc("/profile", h.GetProfile).Methods("GET")

	// Video routes (protected)
	protected.HandleFunc("/videos/search", vh.SearchChannel).Methods("GET")
	protected.HandleFunc("/videos/channel/{channelId}", vh.GetChannelVideos).Methods("GET")
	protected.HandleFunc("/videos/{videoId}", vh.GetVideoDetails).Methods("GET")
	protected.HandleFunc("/videos/{videoId}/transcript", vh.GetVideoTranscript).Methods("GET")
	protected.HandleFunc("/videos/{videoId}/summarize", vh.SummarizeVideo).Methods("GET")

	// Wrap router with CORS and OpenTelemetry middleware
	otelHandler := otelhttp.NewHandler(r, "gateway")
	handler := CORSMiddleware(otelHandler)

	log.Printf("🚀 Gateway starting on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // allow any origin
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

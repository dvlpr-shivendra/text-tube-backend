package handler

import (
	"encoding/json"
	"gateway/internal/client"
	"log"
	"net/http"

	pb "shared/proto"
)

type Handler struct {
	authClient *client.AuthClient
}

func NewHandler(authClient *client.AuthClient) *Handler {
	return &Handler{authClient: authClient}
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ProfileResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

// HealthCheck godoc
// @Summary Show the status of server.
// @Description get the status of server.
// @Tags root
// @Accept  json
// @Produce  json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "gateway"})
}

func (h *Handler) sendJSONError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// Register godoc
// @Summary Register a new user
// @Description Create a new account with username, email, and password
// @Tags auth
// @Accept  json
// @Produce  json
// @Param request body RegisterRequest true "Registration Info"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Register failure: invalid request body: %v", err)
		h.sendJSONError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	resp, err := h.authClient.Register(r.Context(), &pb.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		log.Printf("Register failure: %v", err)
		h.sendJSONError(w, "Registration failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":    resp.Token,
		"user_id":  resp.UserId,
		"username": resp.Username,
	})
}

// Login godoc
// @Summary Log in a user
// @Description Authenticate user with email and password and return a token
// @Tags auth
// @Accept  json
// @Produce  json
// @Param request body LoginRequest true "Login Credentials"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Login failure: invalid request body: %v", err)
		h.sendJSONError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	resp, err := h.authClient.Login(r.Context(), &pb.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		log.Printf("Login failure for email %s: %v", req.Email, err)
		h.sendJSONError(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":    resp.Token,
		"user_id":  resp.UserId,
		"username": resp.Username,
	})
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get current user's profile information using token
// @Tags profile
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} ProfileResponse
// @Router /api/profile [get]
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	username := r.Context().Value("username").(string)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":  userID,
		"username": username,
	})
}

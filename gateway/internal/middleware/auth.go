package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"gateway/internal/client"
	pb "shared/proto"
)

type Middleware struct {
	authClient *client.AuthClient
}

func NewMiddleware(authClient *client.AuthClient) *Middleware {
	return &Middleware{authClient: authClient}
}

func (m *Middleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Println("Auth failure: missing Authorization header")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			log.Println("Auth failure: invalid authorization format")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
			return
		}

		resp, err := m.authClient.ValidateToken(r.Context(), &pb.ValidateTokenRequest{
			Token: token,
		})

		if err != nil || !resp.Valid {
			if err != nil {
				log.Printf("Auth failure: token validation failed: %v", err)
			} else {
				log.Println("Auth failure: token is invalid")
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", resp.UserId)
		ctx = context.WithValue(ctx, "username", resp.Username)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

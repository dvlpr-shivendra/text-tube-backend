package middleware

import (
	"context"
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
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		resp, err := m.authClient.ValidateToken(r.Context(), &pb.ValidateTokenRequest{
			Token: token,
		})

		if err != nil || !resp.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", resp.UserId)
		ctx = context.WithValue(ctx, "username", resp.Username)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

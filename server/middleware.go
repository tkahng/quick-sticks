package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type contextKey string // Define a custom type for context keys to avoid collisions

const playerIDkey contextKey = "player_id"

func getPlayerIDFromContext(ctx context.Context) string {
	playerID, _ := ctx.Value(playerIDkey).(string)
	return playerID
}

func withPlayerID(ctx context.Context, playerID string) context.Context {
	return context.WithValue(ctx, playerIDkey, playerID)
}

func Cors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // Or specific origin, or "*" for all (with caveats)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Authorization, X-CSRF-Token")
		w.Header().Set("Access-Control-Expose-Headers", "Link")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		h.ServeHTTP(w, r)
	})
}

func PlayerID(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var playerID string
		playerIDCookie, err := r.Cookie("player_id")
		if err != nil {
			slog.Error("Failed to get player ID cookie", slog.Any("error", err))
			http.Error(w, "Failed to get player ID cookie", http.StatusInternalServerError)
			return
		}
		if playerIDCookie.Value == "" {
			playerID = generatePlayerID()
		} else {
			playerID = playerIDCookie.Value
		}
		// nolint:exhaustruct
		cookie := &http.Cookie{
			Name:     "my_session_cookie",
			Value:    playerID,
			Expires:  time.Now().Add(24 * time.Hour), // Cookie expires in 24 hours
			Path:     "/",                            // Valid for all paths
			HttpOnly: true,                           // Accessible only via HTTP(S), not JavaScript
			Secure:   true,                           // Only sent over HTTPS
			SameSite: http.SameSiteStrictMode,        // Strict SameSite policy
		}

		// Set the cookie on the response writer
		http.SetCookie(w, cookie)

		r = r.WithContext(withPlayerID(r.Context(), playerID))

		// Call the next handler in the chain
		h.ServeHTTP(w, r)
	})
}

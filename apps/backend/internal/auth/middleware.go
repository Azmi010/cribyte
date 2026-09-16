package auth

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Azmi010/my-drive/apps/backend/internal/session"
)

func Middleware(sessions session.SessionStore, svc *Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_id")
			if err != nil {
				http.Error(w, `{"error":"not authenticated"}`, http.StatusUnauthorized)
				return
			}

			sess, err := sessions.Get(cookie.Value)
			if err != nil || sess == nil {
				http.Error(w, `{"error":"not authenticated"}`, http.StatusUnauthorized)
				return
			}

			user, err := svc.GetUserByID(r.Context(), sess.UserID)
			if err != nil {
				slog.Error("failed to get user from session", "user_id", sess.UserID, "error", err)
				http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), UserKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

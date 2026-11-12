package middleware

import (
	handler "backend/internal/handlers"
	"backend/internal/repository"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"
)

func Middleware(repo *repository.UserRepository, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("session_id")
		if err != nil || cookie.Value == "" {
			fmt.Println("error middleware::", err)
			if errors.Is(err, http.ErrNoCookie) {
				handler.SendError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}
			handler.SendError(w, http.StatusBadRequest, "Bad Request")
			return
		}
		session, err := repo.GetSession(cookie.Value)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				handler.SendError(w, http.StatusUnauthorized, "Unauthorized")
				fmt.Println("error session error", err)
				return
			}
			handler.SendError(w, http.StatusInternalServerError, "Bad Request")
			return
		}
		if time.Now().After(session.ExpiresAt) {
			_ = repo.DeleteSession(cookie.Value)
			handler.SendError(w, http.StatusUnauthorized, "session expired")
			return
		}
		ctx := context.WithValue(r.Context(), "user_id", session.User_id)
		ctx = context.WithValue(ctx, "session_id", cookie.Value)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

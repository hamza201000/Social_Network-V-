package middleware

import (
	handler "backend/internal/handlers"
	"backend/internal/repository"
	"context"
	"database/sql"
	"errors"
	"net/http"
)

func Middleware(repo *repository.UserRepository, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Cookie, err := r.Cookie("session_id")
		cookie, err := r.Cookie("session_id")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				handler.SendError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}
			handler.SendError(w, http.StatusBadRequest, "Bad Request")
			return
		}
		session, err := repo.GetSession(cookie.Name)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				handler.SendError(w, http.StatusUnauthorized, "Unauthorized")
			}
			handler.SendError(w, http.StatusInternalServerError, "Bad Request")
		}
		ctx:=context.WithValue(r.Context(), "user_id", session.User_id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

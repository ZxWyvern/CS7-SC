package httpapi

import (
	"context"
	"errors"
	"net/http"

	"cs7sc/server/internal/store"
)

type ctxKey struct{}

func userFromContext(ctx context.Context) (store.User, bool) {
	user, ok := ctx.Value(ctxKey{}).(store.User)
	return user, ok
}

func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := s.currentUser(r)
		if errors.Is(err, store.ErrNotFound) {
			s.clearSessionCookie(w)
			writeError(w, http.StatusUnauthorized, "Belum masuk.")
			return
		}
		if err != nil {
			s.logger.Error("membaca sesi", "error", err)
			writeError(w, http.StatusInternalServerError, "Terjadi kesalahan pada server.")
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, user)))
	})
}

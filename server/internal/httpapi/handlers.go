package httpapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"cs7sc/server/internal/auth"
	"cs7sc/server/internal/store"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type publicUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

func toPublicUser(user store.User) publicUser {
	return publicUser{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Role:  user.Role,
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if !isJSON(r) {
		writeError(w, http.StatusUnsupportedMediaType, "Content-Type harus application/json.")
		return
	}

	var request loginRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "Permintaan tidak valid.")
		return
	}

	ip := s.clientIP(r)
	if allowed, retryAfter := s.ipLimiter.Allow(ip); !allowed {
		s.tooManyRequests(w, retryAfter)
		return
	}

	email := auth.NormalizeEmail(request.Email)
	if auth.ValidateEmail(email) != nil || len(request.Password) == 0 || len(request.Password) > auth.MaxPasswordLength {
		writeError(w, http.StatusUnauthorized, "Email atau kata sandi salah.")
		return
	}

	accountKey := ip + "|" + email
	if allowed, retryAfter := s.accountLimiter.Allow(accountKey); !allowed {
		s.tooManyRequests(w, retryAfter)
		return
	}

	user, err := s.store.UserByEmail(r.Context(), email)
	if errors.Is(err, store.ErrNotFound) {
		_, _ = s.verifyPassword(request.Password, s.dummyHash)
		s.logger.Info("login gagal", "ip", ip)
		writeError(w, http.StatusUnauthorized, "Email atau kata sandi salah.")
		return
	}
	if err != nil {
		s.logger.Error("membaca pengguna", "error", err)
		writeError(w, http.StatusInternalServerError, "Terjadi kesalahan pada server.")
		return
	}

	valid, err := s.verifyPassword(request.Password, user.PasswordHash)
	if err != nil {
		s.logger.Error("verifikasi kata sandi", "error", err)
		writeError(w, http.StatusInternalServerError, "Terjadi kesalahan pada server.")
		return
	}
	if !valid {
		s.logger.Info("login gagal", "ip", ip)
		writeError(w, http.StatusUnauthorized, "Email atau kata sandi salah.")
		return
	}

	s.ipLimiter.Reset(ip)
	s.accountLimiter.Reset(accountKey)

	token, tokenHash, err := auth.NewSessionToken()
	if err != nil {
		s.logger.Error("membuat token sesi", "error", err)
		writeError(w, http.StatusInternalServerError, "Terjadi kesalahan pada server.")
		return
	}

	expiresAt := s.now().Add(s.cfg.SessionTTL)
	if err := s.store.CreateSession(
		r.Context(),
		user.ID,
		tokenHash,
		expiresAt,
		ip,
		r.UserAgent(),
	); err != nil {
		s.logger.Error("menyimpan sesi", "error", err)
		writeError(w, http.StatusInternalServerError, "Terjadi kesalahan pada server.")
		return
	}

	s.setSessionCookie(w, token, expiresAt)
	writeJSON(w, http.StatusOK, map[string]any{
		"user": toPublicUser(user),
	})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(s.cfg.CookieName); err == nil && len(cookie.Value) <= 256 {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.store.DeleteSession(ctx, auth.HashToken(cookie.Value)); err != nil {
			s.logger.Error("menghapus sesi", "error", err)
		}
	}

	s.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
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

	writeJSON(w, http.StatusOK, map[string]any{
		"user": toPublicUser(user),
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	w.Header().Set("Cache-Control", "no-store")
	if err := s.store.Ping(ctx); err != nil {
		s.logger.Error("ping basis data", "error", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) currentUser(r *http.Request) (store.User, error) {
	cookie, err := r.Cookie(s.cfg.CookieName)
	if err != nil || len(cookie.Value) < 20 || len(cookie.Value) > 256 {
		return store.User{}, store.ErrNotFound
	}

	tokenHash := auth.HashToken(cookie.Value)
	user, err := s.store.UserBySession(r.Context(), tokenHash, s.now())
	if err != nil {
		return store.User{}, err
	}

	_ = s.store.TouchSession(r.Context(), tokenHash, s.now())
	return user, nil
}

func (s *Server) setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cfg.CookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   int(s.cfg.SessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   s.cfg.CookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (s *Server) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cfg.CookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(1, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.cfg.CookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
}

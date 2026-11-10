package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"cs7sc/server/internal/auth"
	"cs7sc/server/internal/config"
	"cs7sc/server/internal/ratelimit"
	"cs7sc/server/internal/store"
)

const maxBodyBytes = 64 << 10

const hashWaitTimeout = 5 * time.Second

type SessionStore interface {
	UserByEmail(ctx context.Context, email string) (store.User, error)
	CreateSession(ctx context.Context, userID string, tokenHash []byte, expiresAt time.Time, ip, userAgent string) error
	UserBySession(ctx context.Context, tokenHash []byte, now time.Time) (store.User, error)
	TouchSession(ctx context.Context, tokenHash []byte, now time.Time) error
	DeleteSession(ctx context.Context, tokenHash []byte) error
	Ping(ctx context.Context) error
}

type Server struct {
	cfg            config.Config
	store          SessionStore
	hasher         *auth.Hasher
	logger         *slog.Logger
	ipLimiter      *ratelimit.Limiter
	accountLimiter *ratelimit.Limiter
	emailLimiter   *ratelimit.Limiter
	dummyHash      string
	hashSlots      chan struct{}
	sleep          func(time.Duration)
	handler        http.Handler
	now            func() time.Time
}

func New(cfg config.Config, st SessionStore, hasher *auth.Hasher, logger *slog.Logger) (*Server, error) {
	dummyRaw := make([]byte, 24)
	if _, err := rand.Read(dummyRaw); err != nil {
		return nil, err
	}
	dummyHash, err := hasher.Hash("dummy-" + base64.RawURLEncoding.EncodeToString(dummyRaw))
	if err != nil {
		return nil, err
	}

	server := &Server{
		cfg:            cfg,
		store:          st,
		hasher:         hasher,
		logger:         logger,
		ipLimiter:      ratelimit.New(30, 15*time.Minute),
		accountLimiter: ratelimit.New(8, 15*time.Minute),
		emailLimiter:   ratelimit.New(30, 15*time.Minute),
		dummyHash:      dummyHash,
		hashSlots:      make(chan struct{}, 2),
		sleep:          time.Sleep,
		now:            time.Now,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/login", server.handleLogin)
	mux.HandleFunc("POST /api/auth/logout", server.handleLogout)
	mux.Handle("GET /api/auth/me", server.requireAuth(http.HandlerFunc(server.handleMe)))
	mux.HandleFunc("GET /api/health", server.handleHealth)

	server.handler = server.withRecovery(
		server.withLogging(
			server.withSecurityHeaders(
				server.withSameOrigin(mux),
			),
		),
	)

	return server, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

func (s *Server) StartLimiterSweeper(ctx context.Context) {
	s.ipLimiter.StartSweeping(ctx, 15*time.Minute)
	s.accountLimiter.StartSweeping(ctx, 15*time.Minute)
	s.emailLimiter.StartSweeping(ctx, 15*time.Minute)
}

func (s *Server) withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				s.logger.Error("panic", "error", recovered, "stack", string(debug.Stack()))
				writeError(w, http.StatusInternalServerError, "Terjadi kesalahan pada server.")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(body)
}

func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		recorder := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(recorder, r)

		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		s.logger.Info("permintaan",
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"durasi_ms", time.Since(started).Milliseconds(),
			"ip", s.clientIP(r),
		)
	})
}

func (s *Server) withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		if strings.HasPrefix(r.URL.Path, "/api/auth/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) withSameOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			origin := r.Header.Get("Origin")
			if origin != "" && origin != s.cfg.AppOrigin {
				writeError(w, http.StatusForbidden, "Asal permintaan tidak diizinkan.")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) clientIP(r *http.Request) string {
	if hops := s.cfg.TrustedProxyHops; hops > 0 {
		parts := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
		if index := len(parts) - hops; index >= 0 {
			if ip := net.ParseIP(strings.TrimSpace(parts[index])); ip != nil {
				return maskIPv6To64(ip)
			}
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func maskIPv6To64(ip net.IP) string {
	if ip.To4() == nil {
		if v6 := ip.To16(); len(v6) == net.IPv6len {
			masked := make([]byte, net.IPv6len)
			copy(masked, v6)
			clear(masked[8:])
			return net.IP(masked).String()
		}
	}
	return ip.String()
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	if r.Body == nil {
		return errors.New("body kosong")
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("hanya satu objek JSON diizinkan")
	}
	return nil
}

func isJSON(r *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	return err == nil && mediaType == "application/json"
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func retryAfterSeconds(duration time.Duration) int {
	seconds := int(duration / time.Second)
	if duration%time.Second != 0 {
		seconds++
	}
	if seconds < 1 {
		seconds = 1
	}
	return seconds
}

func (s *Server) tooManyRequests(w http.ResponseWriter, retryAfter time.Duration) {
	w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds(retryAfter)))
	writeError(w, http.StatusTooManyRequests, "Terlalu banyak percobaan. Coba lagi nanti.")
}

func (s *Server) verifyPassword(ctx context.Context, password, hash string) (bool, error) {
	verifyCtx, cancel := context.WithTimeout(ctx, hashWaitTimeout)
	defer cancel()

	select {
	case s.hashSlots <- struct{}{}:
		defer func() { <-s.hashSlots }()
		return s.hasher.Verify(password, hash)
	case <-verifyCtx.Done():
		return false, verifyCtx.Err()
	}
}

func hashBusyError(err error) bool {
	return errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)
}

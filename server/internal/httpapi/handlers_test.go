package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cs7sc/server/internal/auth"
	"cs7sc/server/internal/config"
	"cs7sc/server/internal/store"
)

type fakeSession struct {
	userID    string
	expiresAt time.Time
	deleted   bool
}

type fakeStore struct {
	users    map[string]store.User
	sessions map[string]*fakeSession
	pingErr  error
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		users:    make(map[string]store.User),
		sessions: make(map[string]*fakeSession),
	}
}

func (f *fakeStore) UserByEmail(_ context.Context, email string) (store.User, error) {
	user, ok := f.users[email]
	if !ok {
		return store.User{}, store.ErrNotFound
	}
	return user, nil
}

func (f *fakeStore) CreateSession(_ context.Context, userID string, tokenHash []byte, expiresAt time.Time, _, _ string) error {
	f.sessions[string(tokenHash)] = &fakeSession{userID: userID, expiresAt: expiresAt}
	return nil
}

func (f *fakeStore) UserBySession(_ context.Context, tokenHash []byte, now time.Time) (store.User, error) {
	session, ok := f.sessions[string(tokenHash)]
	if !ok || session.deleted || !session.expiresAt.After(now) {
		return store.User{}, store.ErrNotFound
	}
	for _, user := range f.users {
		if user.ID == session.userID {
			return user, nil
		}
	}
	return store.User{}, store.ErrNotFound
}

func (f *fakeStore) TouchSession(context.Context, []byte, time.Time) error { return nil }

func (f *fakeStore) DeleteSession(_ context.Context, tokenHash []byte) error {
	if session, ok := f.sessions[string(tokenHash)]; ok {
		session.deleted = true
	}
	return nil
}

func (f *fakeStore) Ping(context.Context) error { return f.pingErr }

func testServer(t *testing.T) (*Server, *fakeStore, *auth.Hasher) {
	t.Helper()

	cfg := config.Config{
		AppOrigin:        "http://localhost:4321",
		CookieName:       "cs7_session",
		SessionTTL:       time.Hour,
		CookieSecure:     false,
		ArgonMemory:      1024,
		ArgonIterations:  1,
		ArgonParallelism: 1,
	}
	fake := newFakeStore()
	hasher := auth.NewHasher(cfg.ArgonMemory, cfg.ArgonIterations, cfg.ArgonParallelism)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	server, err := New(cfg, fake, hasher, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	server.now = func() time.Time { return time.Unix(1000, 0) }
	return server, fake, hasher
}

func addUser(t *testing.T, fake *fakeStore, hasher *auth.Hasher, password string) store.User {
	t.Helper()

	hash, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	user := store.User{
		ID:           "00000000-0000-0000-0000-000000000001",
		Email:        "siswa@contoh.id",
		Name:         "Siswa",
		Role:         "member",
		PasswordHash: hash,
	}
	fake.users[user.Email] = user
	return user
}

func performJSON(server *Server, method, path string, body any, cookie *http.Cookie, origin string) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != nil {
		encoded, _ := json.Marshal(body)
		reader = bytes.NewReader(encoded)
	}

	request := httptest.NewRequest(method, path, reader)
	request.Header.Set("Content-Type", "application/json")
	if origin != "" {
		request.Header.Set("Origin", origin)
	}
	if cookie != nil {
		request.AddCookie(cookie)
	}

	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	return response
}

func TestLoginSuccessSetsCookie(t *testing.T) {
	server, fake, hasher := testServer(t)
	user := addUser(t, fake, hasher, "kata-sandi-uji")

	response := performJSON(server, http.MethodPost, "/api/auth/login", map[string]string{
		"email":    "  SISWA@CONTOH.ID ",
		"password": "kata-sandi-uji",
	}, nil, "http://localhost:4321")

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}

	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "cs7_session" {
		t.Fatalf("cookie sesi tidak diset: %+v", cookies)
	}
	if !cookies[0].HttpOnly {
		t.Fatal("cookie harus HttpOnly")
	}
	if cookies[0].SameSite != http.SameSiteStrictMode {
		t.Fatal("cookie harus SameSite Strict")
	}
	if len(fake.sessions) != 1 {
		t.Fatalf("sesi tersimpan = %d, mau 1", len(fake.sessions))
	}

	var payload struct {
		User publicUser `json:"user"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.User.ID != user.ID || payload.User.Email != user.Email {
		t.Fatalf("user response salah: %+v", payload.User)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	server, fake, hasher := testServer(t)
	addUser(t, fake, hasher, "kata-sandi-uji")

	response := performJSON(server, http.MethodPost, "/api/auth/login", map[string]string{
		"email":    "siswa@contoh.id",
		"password": "salah",
	}, nil, "http://localhost:4321")

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if len(response.Result().Cookies()) != 0 {
		t.Fatal("login gagal tidak boleh menyetel cookie")
	}
}

func TestLoginUnknownEmail(t *testing.T) {
	server, _, _ := testServer(t)

	response := performJSON(server, http.MethodPost, "/api/auth/login", map[string]string{
		"email":    "tidak-ada@contoh.id",
		"password": "kata-sandi-uji",
	}, nil, "http://localhost:4321")

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestLoginRejectsCrossOrigin(t *testing.T) {
	server, fake, hasher := testServer(t)
	addUser(t, fake, hasher, "kata-sandi-uji")

	response := performJSON(server, http.MethodPost, "/api/auth/login", map[string]string{
		"email":    "siswa@contoh.id",
		"password": "kata-sandi-uji",
	}, nil, "https://situs-jahat.example")

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestLoginRequiresJSON(t *testing.T) {
	server, _, _ := testServer(t)

	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString("email=a"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)

	if response.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestLoginRateLimit(t *testing.T) {
	server, fake, hasher := testServer(t)
	addUser(t, fake, hasher, "kata-sandi-uji")

	var response *httptest.ResponseRecorder
	for range 9 {
		response = performJSON(server, http.MethodPost, "/api/auth/login", map[string]string{
			"email":    "siswa@contoh.id",
			"password": "salah",
		}, nil, "http://localhost:4321")
	}

	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, mau 429", response.Code)
	}
	if response.Header().Get("Retry-After") == "" {
		t.Fatal("Retry-After harus ada")
	}
}

func TestMeAndLogout(t *testing.T) {
	server, fake, hasher := testServer(t)
	addUser(t, fake, hasher, "kata-sandi-uji")

	login := performJSON(server, http.MethodPost, "/api/auth/login", map[string]string{
		"email":    "siswa@contoh.id",
		"password": "kata-sandi-uji",
	}, nil, "http://localhost:4321")
	cookie := login.Result().Cookies()[0]

	me := performJSON(server, http.MethodGet, "/api/auth/me", nil, cookie, "")
	if me.Code != http.StatusOK {
		t.Fatalf("me status = %d, body = %s", me.Code, me.Body.String())
	}

	logout := performJSON(server, http.MethodPost, "/api/auth/logout", nil, cookie, "http://localhost:4321")
	if logout.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d", logout.Code)
	}

	meAfter := performJSON(server, http.MethodGet, "/api/auth/me", nil, cookie, "")
	if meAfter.Code != http.StatusUnauthorized {
		t.Fatalf("me setelah logout = %d, mau 401", meAfter.Code)
	}
}

func TestMeWithoutCookie(t *testing.T) {
	server, _, _ := testServer(t)

	response := performJSON(server, http.MethodGet, "/api/auth/me", nil, nil, "")
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestHealth(t *testing.T) {
	server, fake, _ := testServer(t)

	response := performJSON(server, http.MethodGet, "/api/health", nil, nil, "")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}

	fake.pingErr = errors.New("db mati")
	response = performJSON(server, http.MethodGet, "/api/health", nil, nil, "")
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, mau 503", response.Code)
	}
}

func TestLoginDelay(t *testing.T) {
	cases := []struct {
		attempts int
		want     time.Duration
	}{
		{attempts: 0, want: 0},
		{attempts: 1, want: 0},
		{attempts: 8, want: 0},
		{attempts: 9, want: 250 * time.Millisecond},
		{attempts: 10, want: 500 * time.Millisecond},
		{attempts: 11, want: 1000 * time.Millisecond},
		{attempts: 12, want: 2 * time.Second},
		{attempts: 50, want: 2 * time.Second},
	}
	for _, tc := range cases {
		if got := loginDelay(tc.attempts); got != tc.want {
			t.Fatalf("loginDelay(%d) = %v, mau %v", tc.attempts, got, tc.want)
		}
	}
}

func performJSONRemote(server *Server, method, path string, body any, origin, remoteAddr string) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != nil {
		encoded, _ := json.Marshal(body)
		reader = bytes.NewReader(encoded)
	}

	request := httptest.NewRequest(method, path, reader)
	request.Header.Set("Content-Type", "application/json")
	if origin != "" {
		request.Header.Set("Origin", origin)
	}
	request.RemoteAddr = remoteAddr

	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	return response
}

func TestLoginThrottlesPerEmailAcrossIPs(t *testing.T) {
	server, fake, hasher := testServer(t)
	addUser(t, fake, hasher, "kata-sandi-uji")

	var delays []time.Duration
	server.sleep = func(d time.Duration) { delays = append(delays, d) }

	for attempt := range 10 {
		remoteAddr := fmt.Sprintf("192.0.2.%d:44000", attempt+1)
		response := performJSONRemote(server, http.MethodPost, "/api/auth/login", map[string]string{
			"email":    "siswa@contoh.id",
			"password": "salah",
		}, "http://localhost:4321", remoteAddr)

		if response.Code != http.StatusUnauthorized {
			t.Fatalf("percobaan %d: status = %d, mau 401", attempt+1, response.Code)
		}
	}

	if len(delays) != 2 {
		t.Fatalf("tunda tercatat = %d, mau 2", len(delays))
	}
	if delays[0] != 250*time.Millisecond || delays[1] != 500*time.Millisecond {
		t.Fatalf("tunda = %v dan %v, mau 250ms dan 500ms", delays[0], delays[1])
	}
}

func TestVerifyPasswordGivesUpWhenSlotsFull(t *testing.T) {
	server, _, _ := testServer(t)
	server.hashSlots <- struct{}{}
	server.hashSlots <- struct{}{}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := server.verifyPassword(ctx, "kata-sandi-uji", server.dummyHash); !hashBusyError(err) {
		t.Fatalf("context batal harus terkirim sebagai busy, hasil = %v", err)
	}
}

func TestHashBusyError(t *testing.T) {
	if hashBusyError(errors.New("format hash tidak valid")) {
		t.Fatal("error biasa bukan busy")
	}
	if !hashBusyError(fmt.Errorf("dibungkus: %w", context.DeadlineExceeded)) {
		t.Fatal("DeadlineExceeded harus dianggap busy")
	}
}

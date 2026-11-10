package store

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const testPasswordHash = "$argon2id$v=19$m=1024,t=2,p=1$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

const updateAnotherPasswordHash = "$argon2id$v=19$m=1024,t=2,p=1$BBBBBBBBBBBBBBBBBBBBBB$BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"

func openTestStore(t *testing.T) *Store {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL tidak diisi, lewati uji integrasi")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := applyMigrations(ctx, url); err != nil {
		t.Fatalf("migrasi: %v", err)
	}

	database, err := Open(ctx, url)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(database.Close)
	return database
}

func applyMigrations(ctx context.Context, url string) error {
	files, err := filepath.Glob("../../migrations/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(files)

	poolConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		return err
	}
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return err
	}
	defer pool.Close()

	for _, file := range files {
		raw, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		for _, statement := range strings.Split(string(raw), ";") {
			statement = strings.TrimSpace(statement)
			if statement == "" {
				continue
			}
			if _, err := pool.Exec(ctx, statement); err != nil {
				return fmt.Errorf("%s: %w", file, err)
			}
		}
	}
	return nil
}

func testTokenHash(t *testing.T) []byte {
	t.Helper()

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		t.Fatal(err)
	}
	return raw
}

func createTestUser(t *testing.T, database *Store, email string) {
	t.Helper()

	_, err := database.CreateUser(context.Background(), email, "Siswa Uji", testPasswordHash, "member")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
}

func userIDByEmail(t *testing.T, database *Store, email string) string {
	t.Helper()

	var id string
	err := database.pool.QueryRow(context.Background(), `select id::text from users where email = $1`, email).Scan(&id)
	if err != nil {
		t.Fatalf("cari user: %v", err)
	}
	return id
}

func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s-%d@contoh.id", prefix, time.Now().UnixNano())
}

func TestUpdatePasswordRevokesSessions(t *testing.T) {
	database := openTestStore(t)
	ctx := context.Background()
	email := uniqueEmail("revoke")
	createTestUser(t, database, email)

	oldHash := testTokenHash(t)
	if err := database.CreateSession(ctx, userIDByEmail(t, database, email), oldHash, time.Now().Add(time.Hour), "", ""); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	if err := database.UpdatePassword(ctx, email, updateAnotherPasswordHash); err != nil {
		t.Fatalf("UpdatePassword: %v", err)
	}

	if _, err := database.UserBySession(ctx, oldHash, time.Now()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("sesi lama harus dicabut, hasil = %v", err)
	}

	freshHash := testTokenHash(t)
	if err := database.CreateSession(ctx, userIDByEmail(t, database, email), freshHash, time.Now().Add(time.Hour), "", ""); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if _, err := database.UserBySession(ctx, freshHash, time.Now()); err != nil {
		t.Fatalf("sesi setelah ganti kata sandi harus sah: %v", err)
	}
}

func TestUserBySessionRejectsOlderThanPasswordChange(t *testing.T) {
	database := openTestStore(t)
	ctx := context.Background()
	email := uniqueEmail("lama")
	createTestUser(t, database, email)

	if err := database.UpdatePassword(ctx, email, updateAnotherPasswordHash); err != nil {
		t.Fatalf("UpdatePassword: %v", err)
	}

	oldHash := testTokenHash(t)
	const seedOldQuery = `
		insert into sessions (token_hash, user_id, created_at, last_seen_at, expires_at)
		select $1, u.id, u.password_changed_at - interval '10 minutes', now(), now() + interval '1 hour'
		from users u
		where u.email = $2`
	if _, err := database.pool.Exec(ctx, seedOldQuery, oldHash, email); err != nil {
		t.Fatalf("tanam sesi lama: %v", err)
	}
	if _, err := database.UserBySession(ctx, oldHash, time.Now()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("sesi lebih tua dari password_changed_at harus ditolak, hasil = %v", err)
	}

	freshHash := testTokenHash(t)
	const seedFreshQuery = `
		insert into sessions (token_hash, user_id, created_at, last_seen_at, expires_at)
		select $1, u.id, u.password_changed_at + interval '10 minutes', now(), now() + interval '1 hour'
		from users u
		where u.email = $2`
	if _, err := database.pool.Exec(ctx, seedFreshQuery, freshHash, email); err != nil {
		t.Fatalf("tanam sesi baru: %v", err)
	}
	if _, err := database.UserBySession(ctx, freshHash, time.Now()); err != nil {
		t.Fatalf("sesi lebih baru dari password_changed_at harus sah: %v", err)
	}
}

func TestUpdatePasswordUnknownEmail(t *testing.T) {
	database := openTestStore(t)

	if err := database.UpdatePassword(context.Background(), uniqueEmail("tidak-ada"), updateAnotherPasswordHash); !errors.Is(err, ErrNotFound) {
		t.Fatalf("email tidak ada harus ErrNotFound, hasil = %v", err)
	}
}

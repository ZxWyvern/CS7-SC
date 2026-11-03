package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound = errors.New("data tidak ditemukan")
	ErrConflict = errors.New("data sudah ada")
)

type User struct {
	ID           string
	Email        string
	Name         string
	Role         string
	PasswordHash string
}

type Store struct {
	pool *pgxpool.Pool
}

func Open(ctx context.Context, databaseURL string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	cfg.MaxConns = 8
	cfg.MinConns = 1
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 15 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *Store) UserByEmail(ctx context.Context, email string) (User, error) {
	const query = `
		select id::text, email, name, role, password_hash
		from users
		where email = $1`

	var user User
	err := s.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.PasswordHash,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return user, err
}

func (s *Store) CreateUser(ctx context.Context, email, name, passwordHash, role string) (User, error) {
	const query = `
		insert into users (email, name, password_hash, role)
		values ($1, $2, $3, $4)
		returning id::text, email, name, role, password_hash`

	var user User
	err := s.pool.QueryRow(ctx, query, email, name, passwordHash, role).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.PasswordHash,
	)
	if isUniqueViolation(err) {
		return User{}, ErrConflict
	}
	return user, err
}

func (s *Store) UpdatePassword(ctx context.Context, email, passwordHash string) error {
	const query = `
		update users
		set password_hash = $2, password_changed_at = now()
		where email = $1`

	result, err := s.pool.Exec(ctx, query, email, passwordHash)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) CreateSession(
	ctx context.Context,
	userID string,
	tokenHash []byte,
	expiresAt time.Time,
	ip string,
	userAgent string,
) error {
	const query = `
		insert into sessions (token_hash, user_id, expires_at, ip, user_agent)
		values ($1, $2::uuid, $3, nullif($4, ''), nullif($5, ''))`

	_, err := s.pool.Exec(ctx, query, tokenHash, userID, expiresAt, ip, truncate(userAgent, 512))
	return err
}

func (s *Store) UserBySession(ctx context.Context, tokenHash []byte, now time.Time) (User, error) {
	const query = `
		select u.id::text, u.email, u.name, u.role, u.password_hash
		from sessions s
		join users u on u.id = s.user_id
		where s.token_hash = $1 and s.expires_at > $2`

	var user User
	err := s.pool.QueryRow(ctx, query, tokenHash, now).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.PasswordHash,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return user, err
}

func (s *Store) TouchSession(ctx context.Context, tokenHash []byte, now time.Time) error {
	const query = `update sessions set last_seen_at = $2 where token_hash = $1`
	_, err := s.pool.Exec(ctx, query, tokenHash, now)
	return err
}

func (s *Store) DeleteSession(ctx context.Context, tokenHash []byte) error {
	const query = `delete from sessions where token_hash = $1`
	_, err := s.pool.Exec(ctx, query, tokenHash)
	return err
}

func (s *Store) DeleteExpiredSessions(ctx context.Context, now time.Time) (int64, error) {
	const query = `delete from sessions where expires_at <= $1`
	result, err := s.pool.Exec(ctx, query, now)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max]
}

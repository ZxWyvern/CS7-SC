package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"cs7sc/server/internal/auth"
	"cs7sc/server/internal/config"
	"cs7sc/server/internal/store"
)

func runCommand(command string, args []string, logger *slog.Logger) int {
	switch command {
	case "create-user":
		return runCreateUser(args, logger)
	case "set-password":
		return runSetPassword(args, logger)
	default:
		fmt.Fprintf(os.Stderr, "perintah tidak dikenal: %s\n", command)
		fmt.Fprintln(os.Stderr, "perintah tersedia: create-user, set-password")
		return 2
	}
}

func runCreateUser(args []string, logger *slog.Logger) int {
	flags := flag.NewFlagSet("create-user", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	emailFlag := flags.String("email", "", "email pengguna")
	nameFlag := flags.String("name", "", "nama pengguna")
	roleFlag := flags.String("role", "member", "peran: admin atau member")
	generateFlag := flags.Bool("generate", false, "buat kata sandi acak dan tampilkan sekali")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	email := auth.NormalizeEmail(*emailFlag)
	name := strings.TrimSpace(*nameFlag)
	role := strings.TrimSpace(*roleFlag)

	if err := auth.ValidateEmail(email); err != nil {
		fmt.Fprintf(os.Stderr, "email tidak valid: %v\n", err)
		return 2
	}
	if err := auth.ValidateName(name); err != nil {
		fmt.Fprintf(os.Stderr, "nama tidak valid: %v\n", err)
		return 2
	}
	if role != "admin" && role != "member" {
		fmt.Fprintln(os.Stderr, "role harus admin atau member")
		return 2
	}

	password, err := resolvePassword(*generateFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "kata sandi: %v\n", err)
		return 2
	}
	if err := auth.ValidatePassword(password); err != nil {
		fmt.Fprintf(os.Stderr, "kata sandi tidak valid: %v\n", err)
		return 2
	}

	cfg, db, cancel, err := openDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "basis data: %v\n", err)
		return 1
	}
	defer cancel()
	defer db.Close()

	hasher := auth.NewHasher(cfg.ArgonMemory, cfg.ArgonIterations, cfg.ArgonParallelism)
	hash, err := hasher.Hash(password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hash kata sandi: %v\n", err)
		return 1
	}

	user, err := db.CreateUser(context.Background(), email, name, hash, role)
	if errors.Is(err, store.ErrConflict) {
		fmt.Fprintf(os.Stderr, "email sudah terdaftar: %s\n", email)
		return 1
	}
	if err != nil {
		logger.Error("membuat pengguna", "error", err)
		return 1
	}

	fmt.Printf("Akun dibuat: %s (%s, %s)\n", user.Email, user.Name, user.Role)
	if *generateFlag {
		fmt.Printf("Kata sandi sementara: %s\n", password)
		fmt.Println("Simpan kata sandi ini sekarang. Tidak akan ditampilkan lagi.")
	}
	return 0
}

func runSetPassword(args []string, logger *slog.Logger) int {
	flags := flag.NewFlagSet("set-password", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	emailFlag := flags.String("email", "", "email pengguna")
	generateFlag := flags.Bool("generate", false, "buat kata sandi acak dan tampilkan sekali")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	email := auth.NormalizeEmail(*emailFlag)
	if err := auth.ValidateEmail(email); err != nil {
		fmt.Fprintf(os.Stderr, "email tidak valid: %v\n", err)
		return 2
	}

	password, err := resolvePassword(*generateFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "kata sandi: %v\n", err)
		return 2
	}
	if err := auth.ValidatePassword(password); err != nil {
		fmt.Fprintf(os.Stderr, "kata sandi tidak valid: %v\n", err)
		return 2
	}

	cfg, db, cancel, err := openDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "basis data: %v\n", err)
		return 1
	}
	defer cancel()
	defer db.Close()

	hasher := auth.NewHasher(cfg.ArgonMemory, cfg.ArgonIterations, cfg.ArgonParallelism)
	hash, err := hasher.Hash(password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hash kata sandi: %v\n", err)
		return 1
	}

	if err := db.UpdatePassword(context.Background(), email, hash); errors.Is(err, store.ErrNotFound) {
		fmt.Fprintf(os.Stderr, "email tidak ditemukan: %s\n", email)
		return 1
	} else if err != nil {
		logger.Error("mengubah kata sandi", "error", err)
		return 1
	}

	fmt.Printf("Kata sandi diperbarui: %s\n", email)
	if *generateFlag {
		fmt.Printf("Kata sandi sementara: %s\n", password)
		fmt.Println("Simpan kata sandi ini sekarang. Tidak akan ditampilkan lagi.")
	}
	return 0
}

func resolvePassword(generate bool) (string, error) {
	if generate {
		raw := make([]byte, 18)
		if _, err := rand.Read(raw); err != nil {
			return "", err
		}
		return base64.RawURLEncoding.EncodeToString(raw), nil
	}

	data, err := io.ReadAll(io.LimitReader(os.Stdin, 1024))
	if err != nil {
		return "", err
	}
	password := strings.TrimRight(string(data), "\r\n")
	if password == "" {
		return "", errors.New("kata sandi kosong")
	}
	return password, nil
}

func openDatabase() (config.Config, *store.Store, context.CancelFunc, error) {
	cfg, err := config.Load()
	if err != nil {
		return config.Config{}, nil, nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	db, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		cancel()
		return config.Config{}, nil, nil, err
	}

	return cfg, db, cancel, nil
}

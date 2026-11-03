package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Hasher struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

func NewHasher(memory, iterations uint32, parallelism uint8) *Hasher {
	if memory == 0 {
		memory = 19456
	}
	if iterations == 0 {
		iterations = 2
	}
	if parallelism == 0 {
		parallelism = 1
	}
	return &Hasher{
		memory:      memory,
		iterations:  iterations,
		parallelism: parallelism,
		saltLength:  16,
		keyLength:   32,
	}
}

func (h *Hasher) Hash(password string) (string, error) {
	salt := make([]byte, h.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("membuat salt: %w", err)
	}

	key := argon2.IDKey([]byte(password), salt, h.iterations, h.memory, h.parallelism, h.keyLength)
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.memory,
		h.iterations,
		h.parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

func (h *Hasher) Verify(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return false, errors.New("format hash tidak valid")
	}

	version, err := strconv.Atoi(strings.TrimPrefix(parts[2], "v="))
	if err != nil || version != argon2.Version {
		return false, errors.New("versi argon2 tidak didukung")
	}

	params := strings.Split(parts[3], ",")
	if len(params) != 3 {
		return false, errors.New("parameter argon2 tidak valid")
	}

	memory, err := parseUint(params[0], "m=")
	if err != nil {
		return false, err
	}
	iterations, err := parseUint(params[1], "t=")
	if err != nil {
		return false, err
	}
	parallelism, err := parseUint(params[2], "p=")
	if err != nil {
		return false, err
	}

	if memory > 128*1024 || iterations > 10 || parallelism == 0 || parallelism > 16 {
		return false, errors.New("parameter argon2 di luar batas")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) < 8 {
		return false, errors.New("salt tidak valid")
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(expected) < 16 {
		return false, errors.New("hash tidak valid")
	}

	actual := argon2.IDKey(
		[]byte(password),
		salt,
		uint32(iterations),
		uint32(memory),
		uint8(parallelism),
		uint32(len(expected)),
	)
	return subtle.ConstantTimeCompare(actual, expected) == 1, nil
}

func parseUint(value, prefix string) (uint64, error) {
	if !strings.HasPrefix(value, prefix) {
		return 0, errors.New("parameter argon2 tidak valid")
	}
	number, err := strconv.ParseUint(strings.TrimPrefix(value, prefix), 10, 32)
	if err != nil {
		return 0, errors.New("parameter argon2 tidak valid")
	}
	return number, nil
}

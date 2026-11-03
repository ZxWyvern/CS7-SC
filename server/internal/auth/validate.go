package auth

import (
	"errors"
	"strings"
	"unicode/utf8"
)

const (
	MaxEmailLength    = 254
	MinPasswordLength = 8
	MaxPasswordLength = 128
	MaxNameLength     = 80
)

func NormalizeEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func ValidateEmail(value string) error {
	if value == "" || len(value) > MaxEmailLength {
		return errors.New("email tidak valid")
	}
	if strings.ContainsAny(value, " \t\r\n") {
		return errors.New("email tidak valid")
	}
	at := strings.LastIndexByte(value, '@')
	if at <= 0 || at == len(value)-1 {
		return errors.New("email tidak valid")
	}
	domain := value[at+1:]
	if !strings.Contains(domain, ".") || strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return errors.New("email tidak valid")
	}
	return nil
}

func ValidatePassword(value string) error {
	if utf8.RuneCountInString(value) < MinPasswordLength {
		return errors.New("kata sandi minimal 8 karakter")
	}
	if len(value) > MaxPasswordLength {
		return errors.New("kata sandi maksimal 128 karakter")
	}
	return nil
}

func ValidateName(value string) error {
	name := strings.TrimSpace(value)
	if name == "" {
		return errors.New("nama wajib diisi")
	}
	if utf8.RuneCountInString(name) > MaxNameLength {
		return errors.New("nama maksimal 80 karakter")
	}
	return nil
}

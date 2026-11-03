package auth

import (
	"encoding/base64"
	"testing"
)

func TestNewSessionToken(t *testing.T) {
	token, hash, err := NewSessionToken()
	if err != nil {
		t.Fatalf("NewSessionToken: %v", err)
	}

	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		t.Fatalf("token bukan base64url: %v", err)
	}
	if len(raw) != TokenBytes {
		t.Fatalf("panjang token = %d, mau %d", len(raw), TokenBytes)
	}
	if len(hash) != 32 {
		t.Fatalf("panjang hash = %d, mau 32", len(hash))
	}
	if string(hash) == token {
		t.Fatal("hash tidak boleh sama dengan token")
	}
}

func TestHashTokenIsDeterministic(t *testing.T) {
	token := "contoh-token"
	if string(HashToken(token)) != string(HashToken(token)) {
		t.Fatal("hash token harus deterministik")
	}
	if string(HashToken(token)) == string(HashToken(token+"x")) {
		t.Fatal("token berbeda harus menghasilkan hash berbeda")
	}
}

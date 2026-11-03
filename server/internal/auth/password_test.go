package auth

import (
	"strings"
	"testing"
)

func TestHashAndVerify(t *testing.T) {
	hasher := NewHasher(1024, 1, 1)

	hash, err := hasher.Hash("kata-sandi-uji")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Fatalf("hash tidak memakai argon2id: %s", hash)
	}

	ok, err := hasher.Verify("kata-sandi-uji", hash)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !ok {
		t.Fatal("kata sandi benar harus lolos")
	}

	ok, err = hasher.Verify("kata-sandi-salah", hash)
	if err != nil {
		t.Fatalf("Verify salah: %v", err)
	}
	if ok {
		t.Fatal("kata sandi salah tidak boleh lolos")
	}
}

func TestHashUsesUniqueSalt(t *testing.T) {
	hasher := NewHasher(1024, 1, 1)

	first, err := hasher.Hash("sama")
	if err != nil {
		t.Fatal(err)
	}
	second, err := hasher.Hash("sama")
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("dua hash tidak boleh identik")
	}
}

func TestVerifyRejectsInvalidHash(t *testing.T) {
	hasher := NewHasher(1024, 1, 1)

	cases := []string{
		"",
		"$argon2id$v=19$m=1024,t=1,p=1$pendek$pendek",
		"$argon2i$v=19$m=1024,t=1,p=1$c2FsdHNhbHRzYWx0c2FsdA$aGFzaGhhc2hoYXNoaGFzaA",
		"$argon2id$v=19$m=99999999,t=1,p=1$c2FsdHNhbHRzYWx0c2FsdA$aGFzaGhhc2hoYXNoaGFzaA",
	}
	for _, encoded := range cases {
		if _, err := hasher.Verify("apa saja", encoded); err == nil {
			t.Fatalf("hash tidak valid seharusnya ditolak: %q", encoded)
		}
	}
}

func TestNormalizeEmail(t *testing.T) {
	if got := NormalizeEmail("  Siswa@Contoh.ID "); got != "siswa@contoh.id" {
		t.Fatalf("NormalizeEmail = %q", got)
	}
}

func TestValidateEmail(t *testing.T) {
	valid := []string{"siswa@contoh.id", "a.b+c@contoh.co.id"}
	for _, email := range valid {
		if err := ValidateEmail(email); err != nil {
			t.Fatalf("%q seharusnya valid: %v", email, err)
		}
	}

	invalid := []string{"", "siswa", "siswa@", "@contoh.id", "a b@contoh.id", "a@contoh"}
	for _, email := range invalid {
		if err := ValidateEmail(email); err == nil {
			t.Fatalf("%q seharusnya tidak valid", email)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	if err := ValidatePassword("12345678"); err != nil {
		t.Fatalf("8 karakter seharusnya valid: %v", err)
	}
	if err := ValidatePassword("1234567"); err == nil {
		t.Fatal("kurang dari 8 karakter seharusnya ditolak")
	}
	if err := ValidatePassword(strings.Repeat("a", 129)); err == nil {
		t.Fatal("lebih dari 128 karakter seharusnya ditolak")
	}
}

func TestValidateName(t *testing.T) {
	if err := ValidateName("  Pandu  "); err != nil {
		t.Fatalf("nama valid ditolak: %v", err)
	}
	if err := ValidateName("   "); err == nil {
		t.Fatal("nama kosong seharusnya ditolak")
	}
}

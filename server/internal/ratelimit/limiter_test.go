package ratelimit

import (
	"testing"
	"time"
)

func TestLimiterBlocksAfterLimit(t *testing.T) {
	now := time.Unix(0, 0)
	limiter := New(2, time.Minute)
	limiter.now = func() time.Time { return now }

	if ok, _ := limiter.Allow("ip"); !ok {
		t.Fatal("percobaan pertama harus lolos")
	}
	if ok, _ := limiter.Allow("ip"); !ok {
		t.Fatal("percobaan kedua harus lolos")
	}

	ok, retry := limiter.Allow("ip")
	if ok {
		t.Fatal("percobaan ketiga harus diblokir")
	}
	if retry <= 0 || retry > time.Minute {
		t.Fatalf("retryAfter tidak masuk akal: %v", retry)
	}
}

func TestLimiterResetsAfterWindow(t *testing.T) {
	now := time.Unix(0, 0)
	limiter := New(1, time.Minute)
	limiter.now = func() time.Time { return now }

	if ok, _ := limiter.Allow("ip"); !ok {
		t.Fatal("percobaan pertama harus lolos")
	}
	if ok, _ := limiter.Allow("ip"); ok {
		t.Fatal("percobaan kedua harus diblokir")
	}

	now = now.Add(time.Minute)
	if ok, _ := limiter.Allow("ip"); !ok {
		t.Fatal("setelah jendela habis harus lolos lagi")
	}
}

func TestLimiterReset(t *testing.T) {
	limiter := New(1, time.Minute)

	if ok, _ := limiter.Allow("ip"); !ok {
		t.Fatal("percobaan pertama harus lolos")
	}
	limiter.Reset("ip")
	if ok, _ := limiter.Allow("ip"); !ok {
		t.Fatal("setelah reset harus lolos")
	}
}

func TestLimiterSeparatesKeys(t *testing.T) {
	limiter := New(1, time.Minute)

	if ok, _ := limiter.Allow("a"); !ok {
		t.Fatal("key a harus lolos")
	}
	if ok, _ := limiter.Allow("b"); !ok {
		t.Fatal("key b harus lolos")
	}
}

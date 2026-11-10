package ratelimit

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func newTestLimiter(t *testing.T, limit int, window time.Duration) (*Limiter, *time.Time) {
	t.Helper()

	limiter := New(limit, window)
	current := time.Unix(0, 0)
	limiter.now = func() time.Time { return current }
	return limiter, &current
}

func TestLimiterBlocksAfterLimit(t *testing.T) {
	limiter, _ := newTestLimiter(t, 2, time.Minute)

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
	limiter, current := newTestLimiter(t, 1, time.Minute)

	if ok, _ := limiter.Allow("ip"); !ok {
		t.Fatal("percobaan pertama harus lolos")
	}
	if ok, _ := limiter.Allow("ip"); ok {
		t.Fatal("percobaan kedua harus diblokir")
	}

	*current = current.Add(time.Minute)
	if ok, _ := limiter.Allow("ip"); !ok {
		t.Fatal("setelah jendela habis harus lolos lagi")
	}
}

func TestLimiterSeparatesKeys(t *testing.T) {
	limiter, _ := newTestLimiter(t, 1, time.Minute)

	if ok, _ := limiter.Allow("a"); !ok {
		t.Fatal("key a harus lolos")
	}
	if ok, _ := limiter.Allow("b"); !ok {
		t.Fatal("key b harus lolos")
	}
}

func TestLimiterDeniesNewKeysWhenFull(t *testing.T) {
	limiter, _ := newTestLimiter(t, 5, time.Minute)

	for i := range maxEntries {
		limiter.Allow(keyForIndex(i))
	}
	if ok, _ := limiter.Allow("baru"); ok {
		t.Fatal("key baru saat penuh harus ditolak")
	}
	if ok, _ := limiter.Allow(keyForIndex(0)); !ok {
		t.Fatal("key yang sudah ada harus tetap berjalan")
	}
}

func TestLimiterSweepFreesCapacity(t *testing.T) {
	limiter, current := newTestLimiter(t, 5, time.Minute)

	for i := range maxEntries {
		limiter.Allow(keyForIndex(i))
	}
	if ok, _ := limiter.Allow("baru"); ok {
		t.Fatal("key baru saat penuh harus ditolak")
	}

	limiter.Sweep(current.Add(2*time.Minute + time.Second))
	if ok, _ := limiter.Allow("baru"); !ok {
		t.Fatal("key baru harus diterima setelah sweep")
	}
}

func TestLimiterCountTracksCurrentWindow(t *testing.T) {
	limiter, current := newTestLimiter(t, 5, time.Minute)

	if got := limiter.Count("ip"); got != 0 {
		t.Fatalf("Count awal = %d, mau 0", got)
	}
	limiter.Allow("ip")
	limiter.Allow("ip")
	if got := limiter.Count("ip"); got != 2 {
		t.Fatalf("Count = %d, mau 2", got)
	}

	*current = current.Add(time.Minute + time.Second)
	if got := limiter.Count("ip"); got != 0 {
		t.Fatalf("Count lintas jendela = %d, mau 0", got)
	}
}

func TestLimiterConcurrentAllow(t *testing.T) {
	const (
		workers    = 64
		attempts   = 40
		separator  = 8
		limit      = 3
	)
	limiter, _ := newTestLimiter(t, limit, time.Minute)

	var allowed [separator]atomic.Int64
	var done sync.WaitGroup
	done.Add(workers)
	for worker := range workers {
		go func() {
			defer done.Done()
			for range attempts {
				ok, _ := limiter.Allow("kunci-" + string(rune('a'+worker%separator)))
				if ok {
					allowed[worker%separator].Add(1)
				}
			}
		}()
	}
	done.Wait()

	for i := range separator {
		if got := allowed[i].Load(); got != limit {
			t.Fatalf("kunci %c: lolos = %d, mau %d", rune('a'+i), got, limit)
		}
	}
}

func keyForIndex(i int) string {
	return fmt.Sprintf("kunci-%06d", i)
}

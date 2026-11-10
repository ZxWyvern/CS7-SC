package ratelimit

import (
	"context"
	"sync"
	"time"
)

const maxEntries = 20000

type entry struct {
	count       int
	windowStart time.Time
}

type Limiter struct {
	mu      sync.Mutex
	entries map[string]entry
	limit   int
	window  time.Duration
	now     func() time.Time
}

func New(limit int, window time.Duration) *Limiter {
	if limit < 1 {
		limit = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	return &Limiter{
		entries: make(map[string]entry),
		limit:   limit,
		window:  window,
		now:     time.Now,
	}
}

func (l *Limiter) Allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	current, exists := l.entries[key]

	if !exists {
		if len(l.entries) >= maxEntries {
			return false, l.window
		}
		l.entries[key] = entry{count: 1, windowStart: now}
		return true, 0
	}

	if now.Sub(current.windowStart) >= l.window {
		l.entries[key] = entry{count: 1, windowStart: now}
		return true, 0
	}

	if current.count >= l.limit {
		return false, l.window - now.Sub(current.windowStart)
	}

	current.count++
	l.entries[key] = current
	return true, 0
}

func (l *Limiter) Count(key string) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	current, exists := l.entries[key]
	if !exists || now.Sub(current.windowStart) >= l.window {
		return 0
	}
	return current.count
}

func (l *Limiter) Sweep(now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for key, value := range l.entries {
		if now.Sub(value.windowStart) >= 2*l.window {
			delete(l.entries, key)
		}
	}
}

func (l *Limiter) StartSweeping(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = l.window
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				l.Sweep(l.now())
			}
		}
	}()
}

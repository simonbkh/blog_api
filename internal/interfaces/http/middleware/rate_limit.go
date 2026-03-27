package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type bucket struct {
	count     int
	windowEnd time.Time
}

type fixedWindowLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	store  map[string]bucket
}

func newFixedWindowLimiter(limit int, window time.Duration) *fixedWindowLimiter {
	return &fixedWindowLimiter{
		limit:  limit,
		window: window,
		store:  make(map[string]bucket),
	}
}

func (l *fixedWindowLimiter) allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.store[key]
	if !ok || now.After(b.windowEnd) {
		l.store[key] = bucket{count: 1, windowEnd: now.Add(l.window)}
		return true
	}
	if b.count >= l.limit {
		return false
	}
	b.count++
	l.store[key] = b
	return true
}

func (l *fixedWindowLimiter) cleanup(now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for key, b := range l.store {
		if now.After(b.windowEnd) {
			delete(l.store, key)
		}
	}
}

func RateLimit(ipPerMinute, userPerMinute int, identityExtractor func(*http.Request) (string, bool)) func(http.Handler) http.Handler {
	ipLimiter := newFixedWindowLimiter(ipPerMinute, time.Minute)
	userLimiter := newFixedWindowLimiter(userPerMinute, time.Minute)

	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()
		for t := range ticker.C {
			ipLimiter.cleanup(t)
			userLimiter.cleanup(t)
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			now := time.Now()
			if !ipLimiter.allow(clientIP(r), now) {
				http.Error(w, "too many requests (ip)", http.StatusTooManyRequests)
				return
			}
			if userKey, ok := identityExtractor(r); ok {
				if !userLimiter.allow(userKey, now) {
					http.Error(w, "too many requests (user)", http.StatusTooManyRequests)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

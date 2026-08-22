package ratelimit

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
	rate     rate.Limit
	burst    int
}

func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    burst,
	}
}

func (rl *RateLimiter) GetLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	l, exists := rl.limiters[key]
	if !exists {
		l = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[key] = l
	}
	return l
}

func (rl *RateLimiter) Cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			rl.mu.Lock()
			for k, l := range rl.limiters {
				if l.Allow() {
					delete(rl.limiters, k)
				}
			}
			rl.mu.Unlock()
		}
	}()
}

type RateLimitMiddleware struct {
	limiter *RateLimiter
}

func NewRateLimitMiddleware(limiter *RateLimiter) *RateLimitMiddleware {
	return &RateLimitMiddleware{limiter: limiter}
}

func (m *RateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.RemoteAddr
		if !m.limiter.GetLimiter(key).Allow() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"code":"C0006","info":"请求过于频繁"}`))
			return
		}
		next(w, r)
	}
}

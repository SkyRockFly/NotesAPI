package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiterParameters struct {
	Requests        int
	Period          time.Duration
	Burst           int
	VisitorTTL      time.Duration
	CleanupInterval time.Duration
}

type RateLimiter struct {
	mu            sync.Mutex
	visitor       map[string]*visitor
	limit         rate.Limit
	burst         int
	visitorTTL    time.Duration
	cleanInterval time.Duration
	keyFn         KeyFunc
}

type KeyFunc func(r *http.Request) (string, error)

func NewRateLimiter(ctx context.Context, keyFn KeyFunc, param RateLimiterParameters) (*RateLimiter, error) {
	limit := rate.Limit(float64(param.Requests) / param.Period.Seconds())
	rl := &RateLimiter{
		mu:            sync.Mutex{},
		visitor:       make(map[string]*visitor),
		limit:         limit,
		burst:         param.Burst,
		visitorTTL:    param.VisitorTTL,
		cleanInterval: param.CleanupInterval,
		keyFn:         keyFn,
	}

	go rl.cleanupLoop(ctx)

	return rl, nil
}

func (rl *RateLimiter) RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key, err := rl.keyFn(r)
		if err != nil {
			writeRateError(w, http.StatusInternalServerError, "service error")
			return
		}

		visit := rl.getVisitor(key)

		if !visit.limiter.Allow() {
			writeRateError(w, http.StatusTooManyRequests, "too many requests")
			return
		}

		next.ServeHTTP(w, r)
	}
}

func (rl *RateLimiter) getVisitor(ip string) *visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	visit, ok := rl.visitor[ip]
	if !ok {
		visit = &visitor{
			limiter: rate.NewLimiter(rl.limit, rl.burst),
		}

		rl.visitor[ip] = visit
	}

	visit.lastSeen = time.Now().UTC()
	return visit
}

func (rl *RateLimiter) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(rl.cleanInterval)
	defer ticker.Stop()

	for {
		select {
		case now := <-ticker.C:
			rl.cleaupVisitor(now)

		case <-ctx.Done():
			return
		}
	}
}

func (rl *RateLimiter) cleaupVisitor(now time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for ip, v := range rl.visitor {
		if now.Sub(v.lastSeen) >= rl.visitorTTL {
			delete(rl.visitor, ip)
		}
	}
}

func writeRateError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = fmt.Fprintf(w, `{"error":%q}`, message)
}

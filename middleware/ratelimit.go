package middleware

import (
    "LunaMFT/config"
    "golang.org/x/time/rate"
    "net/http"
    "sync"
    "time"
)

var (
    limitersMutex sync.RWMutex
    limiters      = make(map[string]*rate.Limiter)
)

func getOrCreateLimiter(ip string) *rate.Limiter {
    limitersMutex.RLock()
    limiter, exists := limiters[ip]
    limitersMutex.RUnlock()

    if !exists {
        cfg, _ := config.LoadConfig()
        limitersMutex.Lock()
        limiter, exists = limiters[ip]
        if !exists {
            // Create a new rate limiter that allows X requests per minute with a burst of Y
            limiter = rate.NewLimiter(rate.Every(time.Minute/time.Duration(cfg.RateLimit)), 10)
            limiters[ip] = limiter
        }
        limitersMutex.Unlock()
    }

    return limiter
}

// limits the number of requests from a single IP
func RateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get the client's IP address
        ip := r.RemoteAddr
        // For reverse proxies, you might use:
        // ip = r.Header.Get("X-Forwarded-For")
        
        limiter := getOrCreateLimiter(ip)
        
        if !limiter.Allow() {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
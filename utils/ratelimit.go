package utils

import (
	"net/http"
)

func RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement ratelimit
		next(w, r)
	}
}

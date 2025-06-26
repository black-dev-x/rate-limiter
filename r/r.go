package r

import (
	"fmt"
	"net/http"
	"time"
)

type RateLimiter struct {
	defaultRate RequestRate
	apiKeys     map[string]RequestRate
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		defaultRate: RequestRate{Requests: 20, Per: "1s", BlockDuration: "5m"},
		apiKeys:     make(map[string]RequestRate),
	}
}

func (rl *RateLimiter) SetDefaultRate(rate RequestRate) {
	rl.defaultRate = rate
}

func (rl *RateLimiter) AddApiKey(apiKey string, rate RequestRate) {
	rl.apiKeys[apiKey] = rate
}

func (rl *RateLimiter) RegisterMux(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("API_KEY")
		if rate, exists := rl.apiKeys[apiKey]; exists {
			// Use the special rate for this API key
			fmt.Printf("Using special rate for API key %s: %+v\n", apiKey, rate)
		} else {
			// Use the default rate
			fmt.Printf("Using default rate: %+v\n", rl.defaultRate)
		}

		mux.ServeHTTP(w, r)
	})
}

type RequestRate struct {
	Requests      int
	Per           string
	BlockDuration string
}

type AccessControl struct {
	From     string
	Count    int
	Limit    int
	Requests []time.Time
}

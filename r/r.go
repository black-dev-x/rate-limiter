package r

import (
	"net"
	"net/http"
	"time"
)

type RateLimiterConnector struct {
	Limiter RateLimiterInterface
	Type    int
}

const (
	MEMORY = iota
	REDIS
)

func NewRateLimiterConnector(connectorType int) *RateLimiterConnector {
	return &RateLimiterConnector{Limiter: newRateLimiter(connectorType)}
}

func newRateLimiter(connectorType int) RateLimiterInterface {
	if connectorType == MEMORY {
		return NewMemoryLimiter()
	} else if connectorType == REDIS {
		return NewRedisLimiter()
	}
	return nil
}

func (rl *RateLimiterConnector) AddApiKey(apiKey string, rate RequestRate) {
	rl.Limiter.AddApiKey(apiKey, rate)
}

func (rl *RateLimiterConnector) SetDefaultRate(rate RequestRate) {
	rl.Limiter.SetDefaultRate(rate)
}

func (rl *RateLimiterConnector) GetOriginAndRate(r *http.Request) (string, RequestRate) {
	var rate RequestRate
	var origin string
	apiKey := r.Header.Get("API_KEY")
	if rateFound, exists := rl.Limiter.GetApiKeyRate(apiKey); exists {
		rate = rateFound
		origin = apiKey
	} else {
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		origin = host
		rate = rl.Limiter.GetDefaultRate()
	}
	return origin, rate
}

func (rl *RateLimiterConnector) RegisterMux(mux *http.ServeMux) http.Handler {
	limiter := rl.Limiter
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin, rate := rl.GetOriginAndRate(r)
		now := time.Now().Unix()
		if blockedUntil, exists := limiter.BlockedUntil(origin); exists && now < blockedUntil {
			http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
			return
		}
		wasBlocked := limiter.AddUsage(origin, rate)

		if wasBlocked {
			http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
			return
		}

		mux.ServeHTTP(w, r)
	})
}

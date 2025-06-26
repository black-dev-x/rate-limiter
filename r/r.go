package r

import (
	"net/http"
	"time"
)

type RateLimiter struct {
	defaultRate  RequestRate
	apiKeys      map[string]RequestRate
	usage        map[string][]int64
	blockedUntil map[string]int64
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
		now := time.Now().Unix()
		var rate RequestRate
		var origin string
		if rateFound, exists := rl.apiKeys[apiKey]; exists {
			rate = rateFound
			origin = apiKey
		} else {
			origin = r.RemoteAddr
			rate = rl.defaultRate
		}
		println("Rate for origin:", origin)
		if blockedUntil, exists := rl.blockedUntil[origin]; exists && now < blockedUntil {
			http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
			return
		}
		if len(rl.usage[origin]) < rate.Requests {
			rl.usage[origin] = append(rl.usage[origin], now)
		} else {
			oldestRequestTime := rl.usage[origin][0]
			timeGap := now - oldestRequestTime
			duration, _ := time.ParseDuration(rate.Per)
			if timeGap < int64(duration) {
				block, _ := time.ParseDuration(rate.BlockDuration)
				rl.blockedUntil[origin] = now + int64(block)
				http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
				return
			}
		}
		rl.usage[origin] = append(rl.usage[origin], now)

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

package r

import (
	"net"
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
		defaultRate:  RequestRate{Requests: 20, Per: "1s", BlockDuration: "5m"},
		apiKeys:      make(map[string]RequestRate),
		usage:        make(map[string][]int64),
		blockedUntil: make(map[string]int64),
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
			host, _, _ := net.SplitHostPort(r.RemoteAddr)
			origin = host
			rate = rl.defaultRate
		}
		if blockedUntil, exists := rl.blockedUntil[origin]; exists && now < blockedUntil {
			http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
			return
		}
		wasBlocked := rl.AddUsage(origin, rate)
		rl.usage[origin] = append(rl.usage[origin], now)

		if wasBlocked {
			http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
			return
		}

		mux.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) AddUsage(origin string, rate RequestRate) bool {
	currentTime := time.Now().Unix()
	rl.usage[origin] = append(rl.usage[origin], currentTime)
	oldestRequestTime := rl.usage[origin][0]
	if len(rl.usage[origin]) > rate.Requests {
		rl.usage[origin] = rl.usage[origin][1:]
		duration, _ := time.ParseDuration(rate.Per)
		timeToCheck := time.Unix(oldestRequestTime, int64(duration)).Unix()
		if currentTime < timeToCheck {
			block, _ := time.ParseDuration(rate.BlockDuration)
			blockedUntil := time.Unix(currentTime, int64(block)).Unix()
			rl.blockedUntil[origin] = blockedUntil
			return true
		}
	}
	return false
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

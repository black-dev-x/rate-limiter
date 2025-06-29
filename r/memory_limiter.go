package r

import "time"

type MemoryRateLimiter struct {
	defaultRate  RequestRate
	apiKeys      map[string]RequestRate
	usage        map[string][]int64
	blockedUntil map[string]int64
}

func NewMemoryLimiter() RateLimiterInterface {
	return &MemoryRateLimiter{
		apiKeys:      make(map[string]RequestRate),
		usage:        make(map[string][]int64),
		blockedUntil: make(map[string]int64),
	}
}

func (rl *MemoryRateLimiter) GetDefaultRate() RequestRate {
	return rl.defaultRate
}

func (rl *MemoryRateLimiter) SetDefaultRate(rate RequestRate) {
	println("Setting default rate:", rate.Requests, "requests per", rate.Per, "with block duration", rate.BlockDuration)
	rl.defaultRate = rate
}

func (rl *MemoryRateLimiter) AddApiKey(apiKey string, rate RequestRate) {
	rl.apiKeys[apiKey] = rate
}

func (rl *MemoryRateLimiter) GetApiKeyRate(apiKey string) (RequestRate, bool) {
	rate, exists := rl.apiKeys[apiKey]
	return rate, exists
}

func (rl *MemoryRateLimiter) BlockedUntil(origin string) (int64, bool) {
	blockedUntil, exists := rl.blockedUntil[origin]
	return blockedUntil, exists
}

func (rl *MemoryRateLimiter) AddUsage(origin string, rate RequestRate) bool {
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

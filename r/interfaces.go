package r

type RateLimiterInterface interface {
	SetDefaultRate(rate RequestRate)
	GetDefaultRate() RequestRate

	AddApiKey(apiKey string, rate RequestRate)
	GetApiKeyRate(apiKey string) (RequestRate, bool)

	AddUsage(origin string, rate RequestRate) bool
	BlockedUntil(origin string) (int64, bool)
}

type RequestRate struct {
	Requests      int    `redis:"requests"`
	Per           string `redis:"per"`
	BlockDuration string `redis:"block_duration"`
}

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
	Requests      int
	Per           string
	BlockDuration string
}

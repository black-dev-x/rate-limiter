package r

import (
	"context"
	"fmt"
	"time"

	"github.com/black-dev-x/rate-limiter/env"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

type RedisRateLimiter struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisLimiter() RateLimiterInterface {
	_ = godotenv.Load()
	addr := env.String("REDIS_ADDR", "localhost:6379")
	password := env.String("REDIS_PASSWORD", "")
	db := env.Int("REDIS_DB", 0)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisRateLimiter{
		client: client,
		ctx:    context.Background(),
	}
}

const defaultRate = "default_rate"

func (rl *RedisRateLimiter) SetDefaultRate(rate RequestRate) {
	rl.client.Set(rl.ctx, defaultRate, rate, 0)
	rl.client.HSet(rl.ctx, defaultRate, map[string]interface{}{
		"requests":       rate.Requests,
		"per":            rate.Per,
		"block_duration": rate.BlockDuration,
	})
}

func (rl *RedisRateLimiter) GetDefaultRate() RequestRate {
	var rate RequestRate
	rl.client.HGetAll(rl.ctx, defaultRate).Scan(&rate)
	println("Default rate:", rate.Requests, "requests per", rate.Per, "with block duration", rate.BlockDuration)
	return rate
}

func (rl *RedisRateLimiter) AddApiKey(apiKey string, rate RequestRate) {
	rl.client.HSet(rl.ctx, "api_keys:"+apiKey, map[string]interface{}{
		"requests":       rate.Requests,
		"per":            rate.Per,
		"block_duration": rate.BlockDuration,
	})
}

func (rl *RedisRateLimiter) GetApiKeyRate(apiKey string) (RequestRate, bool) {
	var rate RequestRate
	rl.client.HGetAll(rl.ctx, "api_keys:"+apiKey).Scan(&rate)
	if rate.Requests == 0 {
		fmt.Println("Rate for API key not found:", apiKey)
		return RequestRate{}, false
	}
	return rate, true
}

func (rl *RedisRateLimiter) AddUsage(origin string, rate RequestRate) bool {
	println("Adding usage for origin:", origin, "with rate:", rate.Requests)
	now := time.Now().Unix()
	key := "usage:" + origin
	rl.client.RPush(rl.ctx, key, now)
	usages, err := rl.client.LRange(rl.ctx, key, 0, -1).Result()
	if err != nil || len(usages) <= rate.Requests {
		return false
	}
	oldest, _ := rl.client.LPop(rl.ctx, key).Int64()
	duration, _ := time.ParseDuration(rate.Per)
	if now-oldest < int64(duration.Seconds()) {
		block, _ := time.ParseDuration(rate.BlockDuration)
		blockedUntil := now + int64(block.Seconds())
		rl.client.Set(rl.ctx, "blocked:"+origin, blockedUntil, block)
		return true
	}
	return false
}

func (rl *RedisRateLimiter) BlockedUntil(origin string) (int64, bool) {
	ctx := context.Background()
	blockedUntil, err := rl.client.Get(ctx, "blocked:"+origin).Int64()
	if err != nil {
		return 0, false
	}
	return blockedUntil, true
}

package main

import (
	"fmt"
	"net/http"

	"github.com/black-dev-x/rate-limiter/env"
	"github.com/black-dev-x/rate-limiter/r"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	requests := env.Int("REQUESTS", 1)
	per := env.String("PER", "20s")
	blockDuration := env.String("BLOCK_DURATION", "20s")

	specialRequests := env.Int("SPECIAL_REQUESTS", 1)
	specialPer := env.String("SPECIAL_PER", "10s")
	specialBlockDuration := env.String("SPECIAL_BLOCK_DURATION", "1m")

	rateLimiter := r.NewRateLimiterConnector(r.MEMORY)

	defaultRate := r.RequestRate{Requests: requests, Per: per, BlockDuration: blockDuration}
	rateLimiter.SetDefaultRate(defaultRate)

	specialRate := r.RequestRate{Requests: specialRequests, Per: specialPer, BlockDuration: specialBlockDuration}
	rateLimiter.AddApiKey("123abc", specialRate)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	})

	wrappedMux := rateLimiter.RegisterMux(mux)

	http.ListenAndServe(":8080", wrappedMux)
}

package main

import (
	"fmt"
	"net/http"

	"github.com/black-dev-x/rate-limiter/r"
)

func main() {
	rateLimiter := r.NewRateLimiterConnector(r.MEMORY)

	defaultRate := r.RequestRate{Requests: 1, Per: "20s", BlockDuration: "20s"}
	rateLimiter.SetDefaultRate(defaultRate)

	specialRate := r.RequestRate{Requests: 1, Per: "10s", BlockDuration: "1m"}
	rateLimiter.AddApiKey("123abc", specialRate)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	})

	wrappedMux := rateLimiter.RegisterMux(mux)

	http.ListenAndServe(":8080", wrappedMux)
}

package main

import (
	"fmt"
	"net/http"

	"github.com/black-dev-x/rate-limiter/r"
)

func main() {
	rateLimiter := r.NewRateLimiter()

	defaultRate := r.RequestRate{Requests: 2, Per: "1s", BlockDuration: "5m"}
	rateLimiter.SetDefaultRate(defaultRate)

	specialRate := r.RequestRate{Requests: 10, Per: "1s", BlockDuration: "1m"}
	rateLimiter.AddApiKey("123abc", specialRate)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	})

	wrappedMux := rateLimiter.RegisterMux(mux)
	http.ListenAndServe(":8080", wrappedMux)
}

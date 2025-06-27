package main

import (
	"io"
	"net/http"
	"sync"
	"testing"
)

func TestRateLimiter_Blocking(t *testing.T) {
	var wg sync.WaitGroup
	blocked := 0
	requests := 20
	url := "http://localhost:8080/"

	wg.Add(requests)
	for i := 0; i < requests; i++ {
		go func() {
			defer wg.Done()
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				t.Errorf("Failed to create request: %v", err)
				return
			}
			req.Header.Set("API_KEY", "123abc")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Errorf("Request failed: %v", err)
				return
			}
			defer resp.Body.Close()
			_, _ = io.ReadAll(resp.Body)
			if resp.StatusCode == http.StatusTooManyRequests {
				blocked++
			}
		}()
	}
	wg.Wait()

	if blocked == 0 {
		t.Error("Expected some requests to be blocked, but none were.")
	} else {
		t.Logf("%d requests were blocked as expected.", blocked)
	}
}

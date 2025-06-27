package env

import (
	"fmt"
	"os"
)

func String(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func Int(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		var v int
		_, err := fmt.Sscanf(value, "%d", &v)
		if err == nil {
			return v
		}
	}
	return fallback
}

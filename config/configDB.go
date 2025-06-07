package config

import (
	"os"
)

func GetEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		if key == "DB_WITH_SSL" {
			if value == "true" || value == "require" {
				return "require"
			} else {
				return "disable"
			}
		}
		return value
	}
	return ""
}
package env

import (
	"log"
	"os"
)

func MustGetenv(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		log.Panicf("Missing required environment variable: %s", key)
	}
	return val
}

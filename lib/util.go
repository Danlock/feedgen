package lib

import (
	"log"
	"os"
)

func GetEnvOrWarn(k string) (v string) {
	if v = os.Getenv(k); v == "" {
		log.Printf("Missing env %s", k)
	}
	return v
}

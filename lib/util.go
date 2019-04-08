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

// SentinelError is for special errors that require different behaviour, such as io.EOF. This type allows them to be made const. They should not be used for passing information.
type SentinelError string

func (s SentinelError) Error() string {
	return string(s)
}

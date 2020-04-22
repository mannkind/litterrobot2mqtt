package lib

import (
	"time"
)

// Opts - Options for the litter robot lib
type Opts struct {
	Local             bool
	Email             string
	Password          string
	APIKey            string
	APILookupInterval time.Duration
	token             string
	userID            string
}

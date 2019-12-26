package main

import (
	"time"
)

type sourceOpts struct {
	globalOpts
	Local          bool          `env:"LITTERROBOT_LOCAL" envDefault:"false"`
	Email          string        `env:"LITTERROBOT_EMAIL"`
	Password       string        `env:"LITTERROBOT_PASSWORD"`
	APIKey         string        `env:"LITTERROBOT_APIKEY" envDefault:"Gmdfw5Cq3F3Mk6xvvO0inHATJeoDv6C3KfwfOuh0"`
	LookupInterval time.Duration `env:"LITTERROBOT_LOOKUPINTERVAL" envDefault:"37s"`
}

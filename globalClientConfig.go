package main

type globalClientConfig struct {
	KnownRobots knownRobots `env:"LITTERROBOT_KNOWN"`
}

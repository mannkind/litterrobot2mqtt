package main

type globalOpts struct {
	KnownRobots knownRobotMapping `env:"LITTERROBOT_KNOWN"`
}

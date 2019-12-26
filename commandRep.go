package main

const (
	cmdUnknown int64 = iota
	cmdCycle
	cmdWait
	cmdPowerOn
	cmdPowerOff
	cmdPanelLockOn
	cmdPanelLockOff
	cmdNightLightOn
	cmdNightLightOff
)

type commandRep struct {
	command       int64
	litterRobotID string
}

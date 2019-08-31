package main

import "github.com/mannkind/twomqtt"

type knownRobots = map[string]string

const (
	cmdUnknown twomqtt.Command = iota
	cmdCycle
	cmdWait
	cmdPowerOn
	cmdPowerOff
	cmdPanelLockOn
	cmdPanelLockOff
	cmdNightLightOn
	cmdNightLightOff
)

type litterRobotState struct {
	LitterRobotID             string `mqtt:",ignore" mqttDiscoveryType:",ignore"`
	LitterRobotSerial         string `mqtt:",ignore" mqttDiscoveryType:",ignore"`
	PowerStatus               string `mqttDiscoveryType:"sensor"`
	UnitStatus                string `mqttDiscoveryType:"sensor"`
	UnitStatusRaw             string `mqttDiscoveryType:"sensor"`
	CycleCount                string `mqttDiscoveryType:"sensor"`
	CycleCapacity             string `mqttDiscoveryType:"sensor"`
	CyclesAfterDrawerFull     string `mqttDiscoveryType:"sensor"`
	DFICycleCount             string `mqttDiscoveryType:"sensor"`
	CleanCycleWaitTimeMinutes string `mqttDiscoveryType:"sensor"`
	DidNotifyOffline          bool   `mqttDiscoveryType:"binary_sensor"`
	DFITriggered              bool   `mqttDiscoveryType:"binary_sensor"`
	SleepModeActive           bool   `mqttDiscoveryType:"binary_sensor"`
	Power                     bool   `mqttDiscoveryType:"switch"`
	Cycle                     bool   `mqttDiscoveryType:"switch"`
	PanelLockActive           bool   `mqttDiscoveryType:"switch"`
	NightLightActive          bool   `mqttDiscoveryType:"switch"`
}

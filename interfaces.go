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

type litterRobotState struct {
	litterRobotID             string `mqtt:",ignore" mqttDiscoveryType:",ignore"`
	litterRobotSerial         string `mqtt:",ignore" mqttDiscoveryType:",ignore"`
	powerStatus               string `mqttDiscoveryType:"sensor"`
	unitStatus                string `mqttDiscoveryType:"sensor"`
	unitStatusRaw             string `mqttDiscoveryType:"sensor"`
	cycleCount                string `mqttDiscoveryType:"sensor"`
	cycleCapacity             string `mqttDiscoveryType:"sensor"`
	cyclesAfterDrawerFull     string `mqttDiscoveryType:"sensor"`
	dfiCycleCount             string `mqttDiscoveryType:"sensor"`
	cleanCycleWaitTimeMinutes string `mqttDiscoveryType:"sensor"`
	didNotifyOffline          bool   `mqttDiscoveryType:"binary_sensor"`
	dfiTriggered              bool   `mqttDiscoveryType:"binary_sensor"`
	sleepModeActive           bool   `mqttDiscoveryType:"binary_sensor"`
	power                     bool   `mqttDiscoveryType:"switch"`
	cycle                     bool   `mqttDiscoveryType:"switch"`
	panelLockActive           bool   `mqttDiscoveryType:"switch"`
	nightLightActive          bool   `mqttDiscoveryType:"switch"`
}

type event struct {
	version int64
	data    litterRobotState
}

type observer interface {
	receiveState(event)
	receiveCommand(int64, event)
}

type publisher interface {
	register(observer)
}

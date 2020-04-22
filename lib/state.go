package lib

// State - The state of the litter robot
type State struct {
	LitterRobotID             string
	LitterRobotSerial         string
	PowerStatus               string
	UnitStatus                string
	CycleCount                string
	CycleCapacity             string
	CyclesAfterDrawerFull     string
	DFICycleCount             string
	CleanCycleWaitTimeMinutes string
	NameOrIP                  string
	PanelLockActive           bool
	NightLightActive          bool
	DidNotifyOffline          bool
	DFITriggered              bool
	SleepModeActive           bool
}

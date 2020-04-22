package lib

type apiRobotResponse struct {
	PowerStatus               string      `json:"powerStatus"`
	SleepModeStartTime        string      `json:"sleepModeStartTime"`
	LastSeen                  string      `json:"lastSeen"`
	SleepModeEndTime          string      `json:"sleepModeEndTime"`
	AutoOfflineDisabled       bool        `json:"autoOfflineDisabled"`
	SetupDate                 string      `json:"setupDate"`
	DFICycleCount             string      `json:"DFICycleCount"`
	CleanCycleWaitTimeMinutes string      `json:"cleanCycleWaitTimeMinutes"`
	UnitStatus                string      `json:"unitStatus"`
	IsOnboarded               bool        `json:"isOnboarded"`
	DeviceType                string      `json:"deviceType"`
	LitterRobotNickname       string      `json:"litterRobotNickname"`
	CycleCount                string      `json:"cycleCount"`
	PanelLockActive           string      `json:"panelLockActive"`
	CyclesAfterDrawerFull     string      `json:"cyclesAfterDrawerFull"`
	LitterRobotSerial         string      `json:"litterRobotSerial"`
	CycleCapacity             interface{} `json:"cycleCapacity"` // String when full, otherwise Number? Unclear what the pattern is
	LitterRobotID             string      `json:"litterRobotId"`
	NightLightActive          string      `json:"nightLightActive"`
	DidNotifyOffline          bool        `json:"didNotifyOffline"`
	IsDFITriggered            string      `json:"isDFITriggered"`
	SleepModeActive           string      `json:"sleepModeActive"`
}

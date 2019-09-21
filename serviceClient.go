package main

import (
	"strings"

	"github.com/mannkind/litterrobot"
	log "github.com/sirupsen/logrus"
)

type serviceClient struct {
	serviceClientConfig
	stateUpdateChan   stateChannel
	commandUpdateChan commandChannel
	lrStateChan       chan litterrobot.State
	lr                *litterrobot.Client
	state             litterRobotState
}

func newServiceClient(serviceClientCfg serviceClientConfig, stateUpdateChan stateChannel, commandUpdateChan commandChannel) *serviceClient {
	lrStateChan := make(chan litterrobot.State, 100)

	c := serviceClient{
		serviceClientConfig: serviceClientCfg,
		stateUpdateChan:     stateUpdateChan,
		commandUpdateChan:   commandUpdateChan,
		lrStateChan:         lrStateChan,
		lr: litterrobot.NewClient(litterrobot.Opts{
			Local:             serviceClientCfg.Local,
			Email:             serviceClientCfg.Email,
			Password:          serviceClientCfg.Password,
			APIKey:            serviceClientCfg.APIKey,
			APILookupInterval: serviceClientCfg.LookupInterval,
		}, lrStateChan),
	}

	redactedPassword := ""
	if len(c.Password) > 0 {
		redactedPassword = "<REDACTED>"
	}

	log.WithFields(log.Fields{
		"LitterRobot.Local":             c.Local,
		"LitterRobot.KnownRobots":       c.KnownRobots,
		"LitterRobotAPI.Email":          c.Email,
		"LitterRobotAPI.Password":       redactedPassword,
		"LitterRobotAPI.LookupInterval": c.LookupInterval,
	}).Info("Service Environmental Settings")

	return &c
}

func (c *serviceClient) receiveLitterRobotState(info litterrobot.State) {
	log.WithFields(log.Fields{
		"info": info,
	}).Info("Receiving state from Litter Robot")

	obj, err := c.adapt(info)
	if err != nil {
		log.Error("Unable to receive state from Litter Robot")
		return
	}

	c.stateUpdateChan <- obj

	log.Info("Finished receiving state from Litter Robot")
}

func (c *serviceClient) receive() {
	for {
		select {
		case info := <-c.lrStateChan:
			c.receiveLitterRobotState(info)
		case info := <-c.commandUpdateChan:
			c.receiveCommand(info.Command, info.State)
		}
	}
}

func (c *serviceClient) receiveCommand(cmd int64, info litterRobotState) {
	log.WithFields(log.Fields{
		"cmd":  cmd,
		"info": info,
	}).Info("Receiving command to handle")

	if nameOrIP, ok := c.KnownRobots[info.LitterRobotID]; ok {
		info.NameOrIP = nameOrIP
	}

	if cmd == cmdCycle {
		c.lr.Cycle(info.LitterRobotID, info.NameOrIP)
	} else if cmd == cmdPowerOn {
		c.lr.PowerOn(info.LitterRobotID, info.NameOrIP)
	} else if cmd == cmdPowerOff {
		c.lr.PowerOff(info.LitterRobotID, info.NameOrIP)
	} else if cmd == cmdNightLightOn {
		c.lr.NightLightOn(info.LitterRobotID, info.NameOrIP)
	} else if cmd == cmdNightLightOff {
		c.lr.NightLightOff(info.LitterRobotID, info.NameOrIP)
	} else if cmd == cmdPanelLockOn {
		c.lr.PanelLockOn(info.LitterRobotID, info.NameOrIP)
	} else if cmd == cmdPanelLockOff {
		c.lr.PanelLockOff(info.LitterRobotID, info.NameOrIP)
	}

	log.Info("Finished receiving command to handle")
}

func (c *serviceClient) adapt(info litterrobot.State) (litterRobotState, error) {
	log.WithFields(log.Fields{
		"state": info,
	}).Info("Adapting state information")

	power := false
	if !info.SleepModeActive && info.UnitStatus != "OFF" {
		power = true
	}

	cycle := false
	if !info.SleepModeActive && strings.HasPrefix(info.UnitStatus, "CC") {
		cycle = true
	}

	obj := litterRobotState{
		LitterRobotID:             info.LitterRobotID,
		LitterRobotSerial:         info.LitterRobotSerial,
		NameOrIP:                  info.NameOrIP,
		PowerStatus:               info.PowerStatus,
		UnitStatus:                info.UnitStatus,
		UnitStatusRaw:             info.UnitStatus,
		CycleCount:                info.CycleCount,
		CycleCapacity:             info.CycleCapacity,
		CyclesAfterDrawerFull:     info.CyclesAfterDrawerFull,
		DFICycleCount:             info.DFICycleCount,
		CleanCycleWaitTimeMinutes: info.CleanCycleWaitTimeMinutes,
		Power:                     power,
		Cycle:                     cycle,
		PanelLockActive:           info.PanelLockActive,
		NightLightActive:          info.NightLightActive,
		SleepModeActive:           info.SleepModeActive,
		DFITriggered:              info.DFITriggered,
	}

	log.Info("Finished adapting state information")
	return obj, nil
}

func (c *serviceClient) run() {
	c.lr.Run()
	go c.receive()
}

package main

import (
	"strings"

	"github.com/mannkind/litterrobot"
	log "github.com/sirupsen/logrus"
)

type source struct {
	config         sourceOpts
	incoming       <-chan commandRep
	outgoing       chan<- sourceRep
	source         *litterrobot.Client
	sourceIncoming <-chan litterrobot.State
}

func newSource(config sourceOpts, incoming <-chan commandRep, outgoing chan<- sourceRep) *source {
	sourceIncoming := make(chan litterrobot.State, 100)

	c := source{
		config:         config,
		incoming:       incoming,
		outgoing:       outgoing,
		sourceIncoming: sourceIncoming,
		source: litterrobot.NewClient(litterrobot.Opts{
			Local:             config.Local,
			Email:             config.Email,
			Password:          config.Password,
			APIKey:            config.APIKey,
			APILookupInterval: config.LookupInterval,
		}, sourceIncoming),
	}

	return &c
}

func (c *source) run() {
	// Log service settings
	c.logSettings()

	// Run
	c.source.Run()
	go c.read()
}

func (c *source) logSettings() {
	redactedPassword := ""
	if len(c.config.Password) > 0 {
		redactedPassword = "<REDACTED>"
	}

	log.WithFields(log.Fields{
		"LitterRobot.Local":             c.config.Local,
		"LitterRobot.KnownRobots":       c.config.KnownRobots,
		"LitterRobotAPI.Email":          c.config.Email,
		"LitterRobotAPI.Password":       redactedPassword,
		"LitterRobotAPI.LookupInterval": c.config.LookupInterval,
	}).Info("Service Environmental Settings")
}

func (c *source) read() {
	for {
		select {
		case info := <-c.sourceIncoming:
			c.sourceState(info)
		case info := <-c.incoming:
			c.command(info)
		}
	}
}

func (c *source) sourceState(info litterrobot.State) {
	log.WithFields(log.Fields{
		"info": info,
	}).Info("Receiving state from Litter Robot")

	c.outgoing <- c.adapt(info)

	log.Info("Finished receiving state from Litter Robot")
}

func (c *source) command(info commandRep) {
	log.WithFields(log.Fields{
		"cmd":  info.command,
		"info": info.litterRobotID,
	}).Info("Receiving command to handle")

	nameOrIP, ok := c.config.KnownRobots[info.litterRobotID]
	if !ok {
		return
	}

	if info.command == cmdCycle {
		c.source.Cycle(info.litterRobotID, nameOrIP)
	} else if info.command == cmdPowerOn {
		c.source.PowerOn(info.litterRobotID, nameOrIP)
	} else if info.command == cmdPowerOff {
		c.source.PowerOff(info.litterRobotID, nameOrIP)
	} else if info.command == cmdNightLightOn {
		c.source.NightLightOn(info.litterRobotID, nameOrIP)
	} else if info.command == cmdNightLightOff {
		c.source.NightLightOff(info.litterRobotID, nameOrIP)
	} else if info.command == cmdPanelLockOn {
		c.source.PanelLockOn(info.litterRobotID, nameOrIP)
	} else if info.command == cmdPanelLockOff {
		c.source.PanelLockOff(info.litterRobotID, nameOrIP)
	}

	log.Info("Finished receiving command to handle")
}

func (c *source) adapt(info litterrobot.State) sourceRep {
	power := false
	if !info.SleepModeActive && info.UnitStatus != "OFF" {
		power = true
	}

	cycle := false
	if !info.SleepModeActive && strings.HasPrefix(info.UnitStatus, "CC") {
		cycle = true
	}

	return sourceRep{
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
}

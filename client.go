package main

import (
	"strings"
	"time"

	"github.com/mannkind/litterrobot"
	log "github.com/sirupsen/logrus"
)

type client struct {
	observers      map[observer]struct{}
	lr             *litterrobot.Client
	state          litterRobotState
	lookupInterval time.Duration
}

func newClient(config *config) *client {
	c := client{
		observers:      map[observer]struct{}{},
		lr:             litterrobot.NewClient(config.Local, config.Email, config.Password, config.APIKey),
		lookupInterval: config.LookupInterval,
	}

	return &c
}

func (c *client) run() {
	c.lr.Register(c)
	go c.lr.Run()
	go c.loop()
}

func (c *client) register(l observer) {
	c.observers[l] = struct{}{}
}

func (c *client) sendState(e event) {
	log.WithFields(log.Fields{
		"Event": e,
	}).Debug("Sending event to observers")
	for o := range c.observers {
		o.receiveState(e)
	}
}

func (c *client) ReceiveState(e litterrobot.Event) {
	log.WithFields(log.Fields{
		"Event": e,
	}).Debug("Received state from Litter Robot client")

	power := false
	if !e.Data.SleepModeActive && e.Data.UnitStatus != "OFF" {
		power = true
	}

	cycle := false
	if !e.Data.SleepModeActive && strings.HasPrefix(e.Data.UnitStatus, "CC") {
		cycle = true
	}

	c.sendState(event{
		version: 1,
		data: litterRobotState{
			litterRobotID:     e.Data.LitterRobotID,
			litterRobotSerial: e.Data.LitterRobotSerial,

			powerStatus:               e.Data.PowerStatus,
			unitStatus:                e.Data.UnitStatus,
			unitStatusRaw:             e.Data.UnitStatus,
			cycleCount:                e.Data.CycleCount,
			cycleCapacity:             e.Data.CycleCapacity,
			cyclesAfterDrawerFull:     e.Data.CyclesAfterDrawerFull,
			dfiCycleCount:             e.Data.DFICycleCount,
			cleanCycleWaitTimeMinutes: e.Data.CleanCycleWaitTimeMinutes,

			power:            power,
			cycle:            cycle,
			panelLockActive:  e.Data.PanelLockActive,
			nightLightActive: e.Data.NightLightActive,
			sleepModeActive:  e.Data.SleepModeActive,
			dfiTriggered:     e.Data.DFITriggered,
		},
	})
}

func (c *client) receiveState(e event) {}

func (c *client) receiveCommand(cmd int64, e event) {
	log.WithFields(log.Fields{
		"Event": e,
	}).Debug("Received command from observer")

	if cmd == cmdCycle {
		log.Debug("Sending Cycle command to Litter Robot client")
		c.lr.Cycle(e.data.litterRobotID)
	} else if cmd == cmdPowerOn {
		log.Debug("Sending Power On command to Litter Robot client")
		c.lr.PowerOn(e.data.litterRobotID)
	} else if cmd == cmdPowerOff {
		log.Debug("Sending Power Off command to Litter Robot client")
		c.lr.PowerOff(e.data.litterRobotID)
	} else if cmd == cmdNightLightOn {
		log.Debug("Sending Night Light On command to Litter Robot client")
		c.lr.NightLightOn(e.data.litterRobotID)
	} else if cmd == cmdNightLightOff {
		log.Debug("Sending Night Light Off command to Litter Robot client")
		c.lr.NightLightOff(e.data.litterRobotID)
	} else if cmd == cmdPanelLockOn {
		log.Debug("Sending Panel Lock On command to Litter Robot client")
		c.lr.PanelLockOn(e.data.litterRobotID)
	} else if cmd == cmdPanelLockOff {
		log.Debug("Sending Panel Lock Off command to Litter Robot client")
		c.lr.PanelLockOff(e.data.litterRobotID)
	}
}

func (c *client) loop() {
	for {
		c.lr.PublishStates()
		time.Sleep(c.lookupInterval)
	}
}

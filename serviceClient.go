package main

import (
	"reflect"
	"strings"
	"time"

	"github.com/mannkind/litterrobot"
	"github.com/mannkind/twomqtt"
	log "github.com/sirupsen/logrus"
)

type serviceClient struct {
	twomqtt.Observer
	twomqtt.Publisher
	serviceClientConfig
	observers map[twomqtt.Observer]struct{}
	lr        *litterrobot.Client
	state     litterRobotState
}

func newServiceClient(serviceClientCfg serviceClientConfig) *serviceClient {
	c := serviceClient{
		serviceClientConfig: serviceClientCfg,
		observers:           map[twomqtt.Observer]struct{}{},
		lr:                  litterrobot.NewClient(serviceClientCfg.Local, serviceClientCfg.Email, serviceClientCfg.Password, serviceClientCfg.APIKey),
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

func (c *serviceClient) Register(l twomqtt.Observer) {
	c.observers[l] = struct{}{}
}

func (c *serviceClient) ReceiveLitterRobotState(e litterrobot.LREvent) {
	if e.Type != reflect.TypeOf(litterrobot.State{}) {
		msg := "Unexpected event type; skipping"
		log.WithFields(log.Fields{
			"type": e.Type,
		}).Error(msg)
		return
	}

	info := e.Payload.(litterrobot.State)

	log.WithFields(log.Fields{
		"info": info,
	}).Info("Receiving state from Litter Robot")

	power := false
	if !info.SleepModeActive && info.UnitStatus != "OFF" {
		power = true
	}

	cycle := false
	if !info.SleepModeActive && strings.HasPrefix(info.UnitStatus, "CC") {
		cycle = true
	}

	event, err := c.adapt(litterRobotState{
		LitterRobotID:             info.LitterRobotID,
		LitterRobotSerial:         info.LitterRobotSerial,
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
	})
	if err != nil {
		log.Error("Unable to receive state from Litter Robot")
		return
	}

	log.Info("Finished receiving state from Litter Robot")
	c.sendState(event)
}

func (c *serviceClient) ReceiveState(e twomqtt.Event) {}
func (c *serviceClient) ReceiveCommand(cmd twomqtt.Command, e twomqtt.Event) {
	if e.Type != reflect.TypeOf(litterRobotState{}) {
		msg := "Unexpected event type; skipping"
		log.WithFields(log.Fields{
			"type": e.Type,
		}).Error(msg)
		return
	}

	info := e.Payload.(litterRobotState)

	log.WithFields(log.Fields{
		"cmd":  cmd,
		"info": info,
	}).Info("Receiving command to handle")

	if cmd == cmdCycle {
		c.lr.Cycle(info.LitterRobotID)
	} else if cmd == cmdPowerOn {
		c.lr.PowerOn(info.LitterRobotID)
	} else if cmd == cmdPowerOff {
		c.lr.PowerOff(info.LitterRobotID)
	} else if cmd == cmdNightLightOn {
		c.lr.NightLightOn(info.LitterRobotID)
	} else if cmd == cmdNightLightOff {
		c.lr.NightLightOff(info.LitterRobotID)
	} else if cmd == cmdPanelLockOn {
		c.lr.PanelLockOn(info.LitterRobotID)
	} else if cmd == cmdPanelLockOff {
		c.lr.PanelLockOff(info.LitterRobotID)
	}

	log.Info("Finished receiving command to handle")
}

func (c *serviceClient) sendState(e twomqtt.Event) {
	log.WithFields(log.Fields{
		"Event": e,
	}).Debug("Sending event to observers")

	for o := range c.observers {
		o.ReceiveState(e)
	}

	log.Debug("Finished sending event to observers")
}

func (c *serviceClient) adapt(obj litterRobotState) (twomqtt.Event, error) {
	log.WithFields(log.Fields{
		"state": obj,
	}).Info("Adapting state information")

	event := twomqtt.Event{
		Type:    reflect.TypeOf(obj),
		Payload: obj,
	}

	log.Info("Finished adapting state information")
	return event, nil
}

func (c *serviceClient) run() {
	c.lr.Register(c)
	go c.lr.Run()
	go c.loop()
}

func (c *serviceClient) loop() {
	for {
		c.lr.PublishStates()
		time.Sleep(c.LookupInterval)
	}
}

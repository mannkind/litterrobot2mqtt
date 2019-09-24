package main

import (
	"reflect"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mannkind/twomqtt"
	log "github.com/sirupsen/logrus"
)

var statusMapping = map[string]string{
	"RDY": "Ready",
	"Rdy": "Ready",
	"OFF": "Off",
	"P":   "Paused",
	"BR":  "Bonnet removed",
	"DFS": "Drawer is full; will no longer cycle",
	"DF1": "Drawer is full; will cycle twice more",
	"DF2": "Drawer is full; will cycle once more",
	"CST": "Cat sensor triggered",
	"CSI": "Cat interrupted",
	"CSF": "Cat sensor full",
	"CCP": "Cycle processing",
	"CCC": "Cycle complete",
}

type mqttClient struct {
	mqttClientConfig
	*twomqtt.MQTTProxy
	stateUpdateChan   stateChannel
	commandUpdateChan commandChannel
	lastState         map[string]litterRobotState
}

func newMQTTClient(mqttClientCfg mqttClientConfig, client *twomqtt.MQTTProxy, stateUpdateChan stateChannel, commandUpdateChan commandChannel) *mqttClient {
	c := mqttClient{
		mqttClientConfig:  mqttClientCfg,
		MQTTProxy:         client,
		stateUpdateChan:   stateUpdateChan,
		commandUpdateChan: commandUpdateChan,
		lastState:         map[string]litterRobotState{},
	}

	c.Initialize(
		c.onConnect,
		c.onDisconnect,
	)

	c.LogSettings()

	// Setup last known states for known robots
	for id := range c.KnownRobots {
		c.lastState[id] = litterRobotState{
			LitterRobotID: id,
		}
	}

	return &c
}

func (c *mqttClient) run() {
	c.Run()
	go c.receive()
}

func (c *mqttClient) onConnect(client mqtt.Client) {
	log.Info("Finished connecting to MQTT")
	c.Publish(c.AvailabilityTopic(), "online")
	c.subscribe()
	c.publishDiscovery()
}

func (c *mqttClient) onDisconnect(client mqtt.Client, err error) {
	log.WithFields(log.Fields{
		"error": err,
	}).Error("Disconnected from MQTT")
}

func (c *mqttClient) subscribe() {
	// Subscribe to topics
	subscriptions := map[string]mqtt.MessageHandler{}
	for serial := range c.KnownRobots {
		subscriptions[c.CommandTopic(serial, "power")] = c.commandPower
		subscriptions[c.CommandTopic(serial, "cycle")] = c.commandCycle
		subscriptions[c.CommandTopic(serial, "nightlightactive")] = c.commandNightLight
		subscriptions[c.CommandTopic(serial, "panellockactive")] = c.commandPanelLock
	}

	for topic, handler := range subscriptions {
		llog := log.WithFields(log.Fields{
			"topic": topic,
		})

		llog.Info("Subscribing to a new topic")
		if token := c.Client.Subscribe(topic, 0, handler); !token.Wait() || token.Error() != nil {
			llog.WithFields(log.Fields{
				"error": token.Error(),
			}).Error("Error subscribing to topic")
		}
	}
}

func (c *mqttClient) receive() {
	for info := range c.stateUpdateChan {
		c.receiveState(info)
	}
}

func (c *mqttClient) receiveState(info litterRobotState) {
	log.WithFields(log.Fields{
		"info": info,
	}).Info("Information Received")

	c.updateState(info)
}

func (c *mqttClient) updateState(info litterRobotState) {
	obj := reflect.ValueOf(info)
	for i := 0; i < obj.NumField(); i++ {
		field := obj.Type().Field(i)
		val := obj.Field(i)
		sensor := strings.ToLower(field.Name)
		sensorOverride, sensorIgnored := twomqtt.MQTTOverride(field)
		_, sensorTypeIgnored := twomqtt.MQTTDiscoveryOverride(field)

		// Skip any fields tagged as ignored
		if sensorIgnored || sensorTypeIgnored {
			continue
		}

		// Override sensor name
		if sensorOverride != "" {
			sensor = sensorOverride
		}

		topic := c.StateTopic(info.LitterRobotID, sensor)
		payload := ""

		switch val.Kind() {
		case reflect.Bool:
			payload = "OFF"
			if val.Bool() {
				payload = "ON"
			}
		case reflect.String:
			payload = val.String()
			if payloadOverride, ok := statusMapping[payload]; ok && sensor == "unitstatus" {
				payload = payloadOverride
			}
		}

		if payload == "" {
			continue
		}

		c.Publish(topic, payload)
	}

	if _, ok := c.KnownRobots[info.LitterRobotID]; !ok {
		log.WithFields(log.Fields{
			"identifier": info.LitterRobotID,
			"name":       info.NameOrIP,
		}).Warn("NEW LITTER ROBOT FOUND")
	}
	c.lastState[info.LitterRobotID] = info
}

func (c *mqttClient) publishDiscovery() {
	if !c.Discovery {
		return
	}

	log.Info("Publishing MQTT Discovery")

	for litterRobotID := range c.KnownRobots {
		log.WithFields(log.Fields{
			"robot": litterRobotID,
		}).Debug("Iterating through robots")

		obj := reflect.ValueOf(litterRobotState{})
		for i := 0; i < obj.NumField(); i++ {
			field := obj.Type().Field(i)
			sensor := strings.ToLower(field.Name)
			sensorOverride, sensorIgnored := twomqtt.MQTTOverride(field)
			sensorType, sensorTypeIgnored := twomqtt.MQTTDiscoveryOverride(field)

			// Skip any fields tagged as ignored
			if sensorIgnored || sensorTypeIgnored {
				continue
			}

			// Override sensor name
			if sensorOverride != "" {
				sensor = sensorOverride
			}

			mqd := c.NewMQTTDiscovery(litterRobotID, sensor, sensorType)

			if sensorType == "switch" {
				mqd.CommandTopic = c.CommandTopic(litterRobotID, sensor)
			}

			if sensor == "cleancyclewaittimeminutes" {
				mqd.UnitOfMeasurement = "min"
			}

			mqd.Device.Name = Name
			mqd.Device.SWVersion = Version

			c.PublishDiscovery(mqd)
		}

		log.Debug("Finished iterating through robots")
	}

	log.Info("Finished publishing MQTT Discovery")
}

func (c *mqttClient) commandPower(client mqtt.Client, msg mqtt.Message) {
	payload := strings.ToUpper(string(msg.Payload()))
	cmd := cmdPowerOff
	if payload == "ON" {
		cmd = cmdPowerOn
	}

	obj, err := c.adapt(litterRobotState{
		LitterRobotID: c.parseIdentifierFromTopic(msg.Topic()),
	})
	if err != nil {
		return
	}

	c.commandUpdateChan <- struct {
		Command int64
		State   litterRobotState
	}{
		cmd,
		obj,
	}

	serial := c.parseIdentifierFromTopic(msg.Topic())
	state := litterRobotState(c.lastState[serial])
	if cmd == cmdPowerOff {
		state.UnitStatus = "OFF"
	} else if cmd == cmdPowerOn {
		state.UnitStatus = "RDY"
	}

	c.updateState(state)
}

func (c *mqttClient) commandCycle(client mqtt.Client, msg mqtt.Message) {
	cmd := cmdCycle

	obj, err := c.adapt(litterRobotState{
		LitterRobotID: c.parseIdentifierFromTopic(msg.Topic()),
	})
	if err != nil {
		return
	}

	c.commandUpdateChan <- struct {
		Command int64
		State   litterRobotState
	}{
		cmd,
		obj,
	}

	serial := c.parseIdentifierFromTopic(msg.Topic())
	state := litterRobotState(c.lastState[serial])
	state.UnitStatus = "CCP"
	state.UnitStatusRaw = "CCP"

	c.updateState(state)
}

func (c *mqttClient) commandNightLight(client mqtt.Client, msg mqtt.Message) {
	payload := strings.ToUpper(string(msg.Payload()))
	cmd := cmdNightLightOff
	if payload == "ON" {
		cmd = cmdNightLightOn
	}

	obj, err := c.adapt(litterRobotState{
		LitterRobotID: c.parseIdentifierFromTopic(msg.Topic()),
	})
	if err != nil {
		return
	}

	c.commandUpdateChan <- struct {
		Command int64
		State   litterRobotState
	}{
		cmd,
		obj,
	}

	serial := c.parseIdentifierFromTopic(msg.Topic())
	state := litterRobotState(c.lastState[serial])
	if cmd == cmdNightLightOff {
		state.NightLightActive = false
	} else {
		state.NightLightActive = true
	}

	c.updateState(state)
}

func (c *mqttClient) commandPanelLock(client mqtt.Client, msg mqtt.Message) {
	payload := strings.ToUpper(string(msg.Payload()))
	cmd := cmdPanelLockOff
	if payload == "ON" {
		cmd = cmdPanelLockOn
	}

	obj, err := c.adapt(litterRobotState{
		LitterRobotID: c.parseIdentifierFromTopic(msg.Topic()),
	})
	if err != nil {
		return
	}

	c.commandUpdateChan <- struct {
		Command int64
		State   litterRobotState
	}{
		cmd,
		obj,
	}

	serial := c.parseIdentifierFromTopic(msg.Topic())
	state := litterRobotState(c.lastState[serial])
	if cmd == cmdPanelLockOff {
		state.PanelLockActive = false
	} else {
		state.PanelLockActive = true
	}

	c.updateState(state)
}

func (c *mqttClient) parseIdentifierFromTopic(topic string) string {
	log.WithFields(log.Fields{
		"topic": topic,
	}).Debug("Parsing serial from MQTT topic")

	shortTopic := strings.Replace(topic, c.TopicPrefix, "", 1)
	parts := strings.Split(shortTopic, "/")
	pieces := strings.Split(parts[1], "_")

	log.Debug("Finished parsing serial from MQTT topic")
	return pieces[0]
}

func (c *mqttClient) adapt(obj litterRobotState) (litterRobotState, error) {
	log.WithFields(log.Fields{
		"state": obj,
	}).Debug("Adapting state information")

	log.Debug("Finished adapting state information")
	return obj, nil
}

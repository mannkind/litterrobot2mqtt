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
	"EC":  "Emptying container",
	"SDF": "Started, Drawer is full",
}

type sink struct {
	*twomqtt.MQTT
	config    sinkOpts
	incoming  <-chan sourceRep
	outgoing  chan<- commandRep
	lastState map[string]sourceRep
}

func newSink(mqtt *twomqtt.MQTT, config sinkOpts, incoming <-chan sourceRep, outgoing chan<- commandRep) *sink {
	c := sink{
		MQTT:      mqtt,
		config:    config,
		incoming:  incoming,
		outgoing:  outgoing,
		lastState: map[string]sourceRep{},
	}

	c.MQTT.
		SetSubscribeHandler(c.subscribe).
		SetDiscoveryHandler(c.discovery).
		SetReadIncomingChannelHandler(c.read).
		Initialize()

	// Setup last known states for known robots
	for id := range c.config.KnownRobots {
		c.lastState[id] = sourceRep{
			LitterRobotID: id,
		}
	}

	return &c
}

func (c *sink) run() {
	c.Run()
}

func (c *sink) subscribe() {
	// Subscribe to topics
	subscriptions := map[string]mqtt.MessageHandler{}
	for serial := range c.config.KnownRobots {
		subscriptions[c.CommandTopic(serial, "power")] = func(client mqtt.Client, msg mqtt.Message) { c.commandPower(client, msg) }
		subscriptions[c.CommandTopic(serial, "cycle")] = func(client mqtt.Client, msg mqtt.Message) { c.commandCycle(client, msg) }
		subscriptions[c.CommandTopic(serial, "nightlightactive")] = func(client mqtt.Client, msg mqtt.Message) { c.commandNightLight(client, msg) }
		subscriptions[c.CommandTopic(serial, "panellockactive")] = func(client mqtt.Client, msg mqtt.Message) { c.commandPanelLock(client, msg) }
	}

	for topic, handler := range subscriptions {
		llog := log.WithFields(log.Fields{
			"topic": topic,
		})

		llog.Info("Subscribing to a new topic")
		if token := c.Subscribe(topic, 0, handler); !token.Wait() || token.Error() != nil {
			llog.WithFields(log.Fields{
				"error": token.Error(),
			}).Error("Error subscribing to topic")
		}
	}
}

func (c *sink) discovery() []twomqtt.MQTTDiscovery {
	mqds := []twomqtt.MQTTDiscovery{}
	if !c.Discovery {
		return mqds
	}

	for deviceName := range c.config.KnownRobots {
		log.WithFields(log.Fields{
			"robot": deviceName,
		}).Debug("Iterating through robots")

		obj := reflect.ValueOf(sourceRep{})
		for i := 0; i < obj.NumField(); i++ {
			field := obj.Type().Field(i)
			sensorName := strings.ToLower(field.Name)
			sensorOverride, sensorIgnored := twomqtt.MQTTOverride(field)
			sensorType, sensorTypeIgnored := twomqtt.MQTTDiscoveryOverride(field)

			// Skip any fields tagged as ignored
			if sensorIgnored || sensorTypeIgnored {
				continue
			}

			// Override sensor name
			if sensorOverride != "" {
				sensorName = sensorOverride
			}

			mqd := twomqtt.NewMQTTDiscovery(c.config.MQTTOpts, deviceName, sensorName, sensorType)

			if sensorType == "switch" {
				mqd.CommandTopic = c.CommandTopic(deviceName, sensorName)
			}

			if sensorName == "cleancyclewaittimeminutes" {
				mqd.UnitOfMeasurement = "min"
			}

			mqd.Device.Name = Name
			mqd.Device.SWVersion = Version

			mqds = append(mqds, *mqd)
		}
	}

	return mqds
}

func (c *sink) read() {
	for info := range c.incoming {
		c.publish(info)
	}
}

func (c *sink) publish(info sourceRep) []twomqtt.MQTTMessage {
	published := []twomqtt.MQTTMessage{}

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

		msg := c.Publish(topic, payload)
		published = append(published, msg)
	}

	if _, ok := c.config.KnownRobots[info.LitterRobotID]; !ok {
		log.WithFields(log.Fields{
			"identifier": info.LitterRobotID,
			"name":       info.NameOrIP,
		}).Warn("NEW LITTER ROBOT FOUND")
	}
	c.lastState[info.LitterRobotID] = info

	return published
}

func (c *sink) commandPower(client mqtt.Client, msg mqtt.Message) []twomqtt.MQTTMessage {
	payload := strings.ToUpper(string(msg.Payload()))
	cmd := cmdPowerOff
	if payload == "ON" {
		cmd = cmdPowerOn
	}

	c.outgoing <- commandRep{
		command:       cmd,
		litterRobotID: c.parseIdentifierFromTopic(msg.Topic()),
	}

	serial := c.parseIdentifierFromTopic(msg.Topic())
	state := sourceRep(c.lastState[serial])
	if cmd == cmdPowerOff {
		state.UnitStatus = "OFF"
	} else if cmd == cmdPowerOn {
		state.UnitStatus = "RDY"
	}

	return c.publish(state)
}

func (c *sink) commandCycle(client mqtt.Client, msg mqtt.Message) []twomqtt.MQTTMessage {
	cmd := cmdCycle

	c.outgoing <- commandRep{
		command:       cmd,
		litterRobotID: c.parseIdentifierFromTopic(msg.Topic()),
	}

	serial := c.parseIdentifierFromTopic(msg.Topic())
	state := sourceRep(c.lastState[serial])
	state.UnitStatus = "CCP"
	state.UnitStatusRaw = "CCP"

	return c.publish(state)
}

func (c *sink) commandNightLight(client mqtt.Client, msg mqtt.Message) []twomqtt.MQTTMessage {
	payload := strings.ToUpper(string(msg.Payload()))
	cmd := cmdNightLightOff
	if payload == "ON" {
		cmd = cmdNightLightOn
	}

	c.outgoing <- commandRep{
		command:       cmd,
		litterRobotID: c.parseIdentifierFromTopic(msg.Topic()),
	}

	serial := c.parseIdentifierFromTopic(msg.Topic())
	state := sourceRep(c.lastState[serial])
	if cmd == cmdNightLightOff {
		state.NightLightActive = false
	} else {
		state.NightLightActive = true
	}

	return c.publish(state)
}

func (c *sink) commandPanelLock(client mqtt.Client, msg mqtt.Message) []twomqtt.MQTTMessage {
	payload := strings.ToUpper(string(msg.Payload()))
	cmd := cmdPanelLockOff
	if payload == "ON" {
		cmd = cmdPanelLockOn
	}

	c.outgoing <- commandRep{
		command:       cmd,
		litterRobotID: c.parseIdentifierFromTopic(msg.Topic()),
	}

	serial := c.parseIdentifierFromTopic(msg.Topic())
	state := sourceRep(c.lastState[serial])
	if cmd == cmdPanelLockOff {
		state.PanelLockActive = false
	} else {
		state.PanelLockActive = true
	}

	return c.publish(state)
}

func (c *sink) parseIdentifierFromTopic(topic string) string {
	log.WithFields(log.Fields{
		"topic": topic,
	}).Debug("Parsing serial from MQTT topic")

	shortTopic := strings.Replace(topic, c.TopicPrefix, "", 1)
	parts := strings.Split(shortTopic, "/")
	pieces := strings.Split(parts[1], "_")

	log.Debug("Finished parsing serial from MQTT topic")
	return pieces[0]
}

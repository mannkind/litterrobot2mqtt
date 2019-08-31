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
	twomqtt.Observer
	*twomqtt.MQTTProxy
	mqttClientConfig
	observers map[twomqtt.Observer]struct{}
	lastState map[string]litterRobotState
}

func newMQTTClient(mqttClientCfg mqttClientConfig, client *twomqtt.MQTTProxy) *mqttClient {
	c := mqttClient{
		MQTTProxy:        client,
		mqttClientConfig: mqttClientCfg,
		observers:        map[twomqtt.Observer]struct{}{},
		lastState:        map[string]litterRobotState{},
	}

	c.Initialize(
		c.onConnect,
		c.onDisconnect,
	)

	c.LogSettings()

	for serial, id := range c.KnownRobots {
		c.lastState[serial] = litterRobotState{
			LitterRobotSerial: serial,
			LitterRobotID:     id,
		}
	}

	return &c
}

func (c *mqttClient) Register(l twomqtt.Observer) {
	c.observers[l] = struct{}{}
}

func (c *mqttClient) run() {
	c.Run()
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

func (c *mqttClient) ReceiveCommand(cmd twomqtt.Command, e twomqtt.Event) {}
func (c *mqttClient) ReceiveState(e twomqtt.Event) {
	if e.Type != reflect.TypeOf(litterRobotState{}) {
		msg := "Unexpected event type; skipping"
		log.WithFields(log.Fields{
			"type": e.Type,
		}).Error(msg)
		return
	}

	info := e.Payload.(litterRobotState)

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

		topic := c.StateTopic(info.LitterRobotSerial, sensor)
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

	if _, ok := c.KnownRobots[info.LitterRobotSerial]; !ok {
		log.WithFields(log.Fields{
			"serial":     info.LitterRobotSerial,
			"identifier": info.LitterRobotID,
		}).Warn("NEW LITTER ROBOT FOUND")
	}
	c.lastState[info.LitterRobotSerial] = info
}

func (c *mqttClient) publishDiscovery() {
	if !c.Discovery {
		return
	}

	log.Info("Publishing MQTT Discovery")

	for serial := range c.KnownRobots {
		log.WithFields(log.Fields{
			"robot": serial,
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

			mqd := c.NewMQTTDiscovery(serial, sensor, sensorType)

			if sensorType == "switch" {
				mqd.CommandTopic = c.CommandTopic(serial, sensor)
			}

			if sensor == "cleancyclewaittimeminutes" {
				mqd.UnitOfMeasurement = "min"
			}

			c.PublishDiscovery(mqd)
		}

		log.Debug("Finished iterating through robots")
	}

	log.Info("Finished publishing MQTT Discovery")
}

func (c *mqttClient) sendCommand(cmd twomqtt.Command, e twomqtt.Event) {
	for o := range c.observers {
		o.ReceiveCommand(cmd, e)
	}
}

func (c *mqttClient) commandPower(client mqtt.Client, msg mqtt.Message) {
	payload := strings.ToUpper(string(msg.Payload()))
	cmd := cmdPowerOff
	if payload == "ON" {
		cmd = cmdPowerOn
	}

	event, err := c.adapt(litterRobotState{
		LitterRobotID: c.parseIdentifierFromTopic(msg.Topic()),
	})
	if err != nil {
		return
	}
	c.sendCommand(cmd, event)

	serial := c.parseSerialFromTopic(msg.Topic())
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

	event, err := c.adapt(litterRobotState{
		LitterRobotID: c.parseIdentifierFromTopic(msg.Topic()),
	})
	if err != nil {
		return
	}
	c.sendCommand(cmd, event)

	serial := c.parseSerialFromTopic(msg.Topic())
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

	event, err := c.adapt(litterRobotState{
		LitterRobotID: c.parseIdentifierFromTopic(msg.Topic()),
	})
	if err != nil {
		return
	}
	c.sendCommand(cmd, event)

	serial := c.parseSerialFromTopic(msg.Topic())
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

	event, err := c.adapt(litterRobotState{
		LitterRobotID: c.parseIdentifierFromTopic(msg.Topic()),
	})
	if err != nil {
		return
	}
	c.sendCommand(cmd, event)

	serial := c.parseSerialFromTopic(msg.Topic())
	state := litterRobotState(c.lastState[serial])
	if cmd == cmdPanelLockOff {
		state.PanelLockActive = false
	} else {
		state.PanelLockActive = true
	}

	c.updateState(state)
}

func (c *mqttClient) parseSerialFromTopic(topic string) string {
	log.WithFields(log.Fields{
		"topic": topic,
	}).Debug("Parsing serial from MQTT topic")

	shortTopic := strings.Replace(topic, c.TopicPrefix, "", 1)
	parts := strings.Split(shortTopic, "/")
	pieces := strings.Split(parts[1], "_")

	log.Debug("Finished parsing serial from MQTT topic")
	return pieces[0]
}

func (c *mqttClient) parseIdentifierFromTopic(topic string) string {
	log.WithFields(log.Fields{
		"topic": topic,
	}).Debug("Parsing identifier based on MQTT topic")

	identifier := c.parseSerialFromTopic(topic)

	if id, ok := c.KnownRobots[identifier]; ok {
		identifier = id
	}

	log.Debug("Finished identifier based on MQTT topic")
	return identifier
}

func (c *mqttClient) adapt(obj litterRobotState) (twomqtt.Event, error) {
	log.WithFields(log.Fields{
		"state": obj,
	}).Debug("Adapting state information")

	event := twomqtt.Event{
		Type:    reflect.TypeOf(obj),
		Payload: obj,
	}

	log.Debug("Finished adapting state information")
	return event, nil
}

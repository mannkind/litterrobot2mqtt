package main

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	mqttExtDI "github.com/mannkind/paho.mqtt.golang.ext/di"
	mqttExtHA "github.com/mannkind/paho.mqtt.golang.ext/ha"
	log "github.com/sirupsen/logrus"
)

const (
	sensorUniqueTemplate = "%s.%s.%s"
	sensorTopicTemplate  = "%s/%s/%s/%s"
)

type mqttClient struct {
	observers map[observer]struct{}

	discovery       bool
	discoveryPrefix string
	discoveryName   string
	topicPrefix     string

	client mqtt.Client

	statusMapping map[string]string
	knownRobots   map[string]string

	lastState     map[string]litterRobotState
	lastPublished map[string]string
}

func newMQTTClient(config *config, mqttFuncWrapper mqttExtDI.MQTTFuncWrapper) *mqttClient {
	c := mqttClient{
		observers:       map[observer]struct{}{},
		discovery:       config.MQTT.Discovery,
		discoveryPrefix: config.MQTT.DiscoveryPrefix,
		discoveryName:   config.MQTT.DiscoveryName,
		topicPrefix:     config.MQTT.TopicPrefix,
		lastState:       map[string]litterRobotState{},
		lastPublished:   map[string]string{},
	}

	c.knownRobots = make(map[string]string)

	for _, m := range config.KnownRobots {
		parts := strings.Split(m, ":")
		if len(parts) != 2 {
			continue
		}

		serial := parts[0]
		ip := parts[1]
		c.knownRobots[serial] = ip
	}

	c.statusMapping = map[string]string{
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

	c.client = mqttFuncWrapper(
		config.MQTT,
		c.onConnect,
		c.onDisconnect,
		c.availabilityTopic(),
	)

	return &c
}

func (c *mqttClient) register(l observer) {
	c.observers[l] = struct{}{}
}

func (c *mqttClient) run() {
	c.runAfter(0 * time.Second)
}

func (c *mqttClient) runAfter(delay time.Duration) {
	time.Sleep(delay)

	log.Info("Connecting to MQTT")
	if token := c.client.Connect(); !token.Wait() || token.Error() != nil {
		log.WithFields(log.Fields{
			"error": token.Error(),
		}).Error("Error connecting to MQTT")

		delay = c.adjustReconnectDelay(delay)

		log.WithFields(log.Fields{
			"delay": delay,
		}).Info("Sleeping before attempting to reconnect to MQTT")

		c.runAfter(delay)
	}
}

func (c *mqttClient) adjustReconnectDelay(delay time.Duration) time.Duration {
	var maxDelay float64 = 120
	defaultDelay := 2 * time.Second

	// No delay, set to default delay
	if delay.Seconds() == 0 {
		delay = defaultDelay
	} else {
		// Increment the delay
		delay = delay * 2

		// If the delay is above two minutes, reset to default
		if delay.Seconds() > maxDelay {
			delay = defaultDelay
		}
	}

	return delay
}

func (c *mqttClient) onConnect(client mqtt.Client) {
	log.Info("Connected to MQTT")
	c.publish(c.availabilityTopic(), "online")
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
	for serial := range c.knownRobots {
		subscriptions[c.commandTopic(serial, "power")] = c.commandPower
		subscriptions[c.commandTopic(serial, "cycle")] = c.commandCycle
		subscriptions[c.commandTopic(serial, "nightlightactive")] = c.commandNightLight
		subscriptions[c.commandTopic(serial, "panellockactive")] = c.commandPanelLock
	}

	for topic, handler := range subscriptions {
		llog := log.WithFields(log.Fields{
			"topic": topic,
		})

		llog.Info("Subscribing to a new topic")
		if token := c.client.Subscribe(topic, 0, handler); !token.Wait() || token.Error() != nil {
			llog.WithFields(log.Fields{
				"error": token.Error(),
			}).Error("Error subscribing to topic")
		}
	}
}

func (c *mqttClient) receiveCommand(cmd int64, e event) {}
func (c *mqttClient) receiveState(e event) {
	info := e.data
	c.updateState(info)
}

func (c *mqttClient) updateState(info litterRobotState) {
	obj := reflect.ValueOf(info)
	for i := 0; i < obj.NumField(); i++ {
		field := obj.Type().Field(i)
		val := obj.Field(i)
		sensor := strings.ToLower(field.Name)
		sensorOverride := field.Tag.Get("mqtt")
		sensorType := field.Tag.Get("mqttDiscoveryType")

		// Skip any fields tagged as ignored for mqtt
		if strings.Contains(sensorOverride, ",ignore") {
			continue
		}

		// Override sensor name
		if sensorOverride != "" {
			sensor = sensorOverride
		}

		// Skip any fields tagged as ignores for discovery
		if strings.Contains(sensorType, ",ignore") {
			continue
		}

		topic := c.stateTopic(info.litterRobotSerial, sensor)
		payload := ""

		switch val.Kind() {
		case reflect.Bool:
			payload = "OFF"
			if val.Bool() {
				payload = "ON"
			}
		case reflect.String:
			payload = val.String()
			if payloadOverride, ok := c.statusMapping[payload]; ok && sensor == "unitstatus" {
				payload = payloadOverride
			}
		}

		if payload == "" {
			continue
		}

		c.publish(topic, payload)
	}

	if _, ok := c.knownRobots[info.litterRobotSerial]; !ok {
		log.WithFields(log.Fields{
			"serial":     info.litterRobotSerial,
			"identifier": info.litterRobotID,
		}).Warn("NEW LITTER ROBOT FOUND")
	}
	c.lastState[info.litterRobotSerial] = info
}

func (c *mqttClient) availabilityTopic() string {
	return fmt.Sprintf("%s/status", c.topicPrefix)
}

func (c *mqttClient) stateTopic(identifier string, sensor string) string {
	return c.genericTopic(identifier, sensor, "state")
}

func (c *mqttClient) commandTopic(identifier string, sensor string) string {
	return c.genericTopic(identifier, sensor, "command")
}

func (c *mqttClient) genericTopic(identifier string, sensor string, rw string) string {
	return fmt.Sprintf(sensorTopicTemplate, c.topicPrefix, identifier, sensor, rw)
}

func (c *mqttClient) uniqueName(identifier string, sensor string) string {
	return fmt.Sprintf(strings.ReplaceAll(sensorUniqueTemplate, ".", " "), c.discoveryName, identifier, sensor)
}

func (c *mqttClient) uniqueID(identifier string, sensor string) string {
	return fmt.Sprintf(sensorUniqueTemplate, c.discoveryName, identifier, sensor)
}

func (c *mqttClient) publishDiscovery() {
	if !c.discovery {
		return
	}

	for serial := range c.knownRobots {
		obj := reflect.ValueOf(litterRobotState{})
		for i := 0; i < obj.NumField(); i++ {
			field := obj.Type().Field(i)
			sensor := strings.ToLower(field.Name)
			sensorOverride := field.Tag.Get("mqtt")
			sensorType := field.Tag.Get("mqttDiscoveryType")

			// Skip any fields tagged as ignored for mqtt
			if strings.Contains(sensorOverride, ",ignore") {
				continue
			}

			// Override sensor name
			if sensorOverride != "" {
				sensor = sensorOverride
			}

			// Skip any fields tagged as ignores for discovery
			if strings.Contains(sensorType, ",ignore") {
				continue
			}

			mqd := mqttExtHA.MQTTDiscovery{
				DiscoveryPrefix: c.discoveryPrefix,
				Component:       sensorType,
				NodeID:          c.discoveryName,
				ObjectID:        sensor,

				AvailabilityTopic: c.availabilityTopic(),
				Name:              c.uniqueName(serial, sensor),
				StateTopic:        c.stateTopic(serial, sensor),
				UniqueID:          c.uniqueID(serial, sensor),
			}

			if sensorType == "switch" {
				mqd.CommandTopic = c.commandTopic(serial, sensor)
			}

			if sensor == "cleancyclewaittimeminutes" {
				mqd.UnitOfMeasurement = "min"
			}

			mqd.PublishDiscovery(c.client)
		}
	}
}

func (c *mqttClient) sendCommand(cmd int64, e event) {
	for s := range c.observers {
		s.receiveCommand(cmd, e)
	}
}

func (c *mqttClient) commandPower(client mqtt.Client, msg mqtt.Message) {
	payload := strings.ToUpper(string(msg.Payload()))
	cmd := cmdPowerOff
	if payload == "ON" {
		cmd = cmdPowerOn
	}
	c.sendCommand(cmd, event{
		version: 1,
		data: litterRobotState{
			litterRobotID: c.parseIdentifierFromTopic(msg.Topic()),
		},
	})

	serial := c.parseSerialFromTopic(msg.Topic())
	state := litterRobotState(c.lastState[serial])
	if cmd == cmdPowerOff {
		state.unitStatus = "OFF"
	} else if cmd == cmdPowerOn {
		state.unitStatus = "CCP"
	}

	c.updateState(state)
}

func (c *mqttClient) commandCycle(client mqtt.Client, msg mqtt.Message) {
	cmd := cmdCycle

	c.sendCommand(cmd, event{
		version: 1,
		data: litterRobotState{
			litterRobotID: c.parseIdentifierFromTopic(msg.Topic()),
		},
	})

	serial := c.parseSerialFromTopic(msg.Topic())
	state := litterRobotState(c.lastState[serial])
	state.unitStatus = "CCP"
	state.unitStatusRaw = "CCP"

	c.updateState(state)
}

func (c *mqttClient) commandNightLight(client mqtt.Client, msg mqtt.Message) {
	payload := strings.ToUpper(string(msg.Payload()))
	cmd := cmdNightLightOff
	if payload == "ON" {
		cmd = cmdNightLightOn
	}

	c.sendCommand(cmd, event{
		version: 1,
		data: litterRobotState{
			litterRobotID: c.parseIdentifierFromTopic(msg.Topic()),
		},
	})

	serial := c.parseSerialFromTopic(msg.Topic())
	state := litterRobotState(c.lastState[serial])
	if cmd == cmdNightLightOff {
		state.nightLightActive = false
	} else {
		state.nightLightActive = true
	}

	c.updateState(state)
}

func (c *mqttClient) commandPanelLock(client mqtt.Client, msg mqtt.Message) {
	payload := strings.ToUpper(string(msg.Payload()))
	cmd := cmdPanelLockOff
	if payload == "ON" {
		cmd = cmdPanelLockOn
	}

	c.sendCommand(cmd, event{
		version: 1,
		data: litterRobotState{
			litterRobotID: c.parseIdentifierFromTopic(msg.Topic()),
		},
	})

	serial := c.parseSerialFromTopic(msg.Topic())
	state := litterRobotState(c.lastState[serial])
	if cmd == cmdPanelLockOff {
		state.panelLockActive = false
	} else {
		state.panelLockActive = true
	}

	c.updateState(state)
}

func (c *mqttClient) parseSerialFromTopic(topic string) string {
	shortTopic := strings.Replace(topic, c.topicPrefix, "", 1)
	parts := strings.Split(shortTopic, "/")
	return parts[1]
}

func (c *mqttClient) parseIdentifierFromTopic(topic string) string {
	identifier := c.parseSerialFromTopic(topic)

	if id, ok := c.knownRobots[identifier]; ok {
		identifier = id
	}

	return identifier
}

func (c *mqttClient) publish(topic string, payload string) {
	llog := log.WithFields(log.Fields{
		"topic":   topic,
		"payload": payload,
	})
	// Should we publish this again?
	// NOTE: We must allow the availability topic to publish duplicates
	if lastPayload, ok := c.lastPublished[topic]; topic != c.availabilityTopic() && ok && lastPayload == payload {
		llog.Debug("Duplicate payload")
		return
	}

	llog.Info("Publishing to MQTT")

	retain := true
	if token := c.client.Publish(topic, 0, retain, payload); token.Wait() && token.Error() != nil {
		log.Error("Publishing error")
	}

	llog.Debug("Published to MQTT")
	c.lastPublished[topic] = payload
}

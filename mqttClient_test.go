package main

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/mannkind/twomqtt"
	log "github.com/sirupsen/logrus"
)

const knownRobot = "LR3A134568:A63afb501d65cb"
const knownSerial = "LR3A134568"
const knownDiscoveryName = "litterrobotDiscoveryName"
const knownPrefix = "home/litterrobotTopicPrefix"

var knownTypes = []string{
	"unitstatus", "powerstatus", "cyclecount",
}

func init() {
	log.SetLevel(log.PanicLevel)
}

func setEnvs(d, dn, tp, a string) {
	os.Setenv("MQTT_DISCOVERY", d)
	os.Setenv("MQTT_DISCOVERYNAME", dn)
	os.Setenv("MQTT_TOPICPREFIX", tp)
	os.Setenv("LITTERROBOT_KNOWN", a)
}

func clearEnvs() {
	setEnvs("false", "", "", "")
}

func TestDiscovery(t *testing.T) {
	defer clearEnvs()

	for _, knownType := range knownTypes {
		var tests = []struct {
			Known           string
			DiscoveryName   string
			TopicPrefix     string
			ExpectedTopic   string
			ExpectedPayload string
		}{
			{
				knownRobot,
				knownDiscoveryName,
				knownPrefix,
				"homeassistant/sensor/" + knownDiscoveryName + "/" + strings.ToLower(knownSerial) + "_" + knownType + "/config",
				"{\"availability_topic\":\"" + knownPrefix + "/status\",\"name\":\"" + strings.ToLower(knownSerial) + " " + knownType + "\",\"state_topic\":\"" + knownPrefix + "/" + strings.ToLower(knownSerial) + "/" + knownType + "/state\",\"unique_id\":\"" + knownDiscoveryName + "." + strings.ToLower(knownSerial) + "." + knownType + "\"}",
			},
		}

		for _, v := range tests {
			setEnvs("true", v.DiscoveryName, v.TopicPrefix, v.Known)

			c := initialize()
			c.mqttClient.publishDiscovery()

			actualPayload := c.mqttClient.LastPublishedOnTopic(v.ExpectedTopic)
			if actualPayload != v.ExpectedPayload {
				t.Errorf("Actual:%s\nExpected:%s", actualPayload, v.ExpectedPayload)
			}
		}
	}
}

func TestReceieveState(t *testing.T) {
	defer clearEnvs()

	for _, knownType := range knownTypes {
		var tests = []struct {
			Known           string
			Serial          string
			TopicPrefix     string
			ExpectedTopic   string
			ExpectedPayload string
		}{
			{
				knownRobot,
				knownSerial,
				knownPrefix,
				knownPrefix + "/" + strings.ToLower(knownSerial) + "/" + knownType + "/state",
				"Off",
			},
		}

		for _, v := range tests {
			setEnvs("false", "", v.TopicPrefix, v.Known)

			c := initialize()
			obj := litterRobotState{
				LitterRobotSerial: v.Serial,
				PowerStatus:       "Off", // Not a real status
				UnitStatus:        "OFF",
				CycleCount:        "Off", // Not a real status
			}
			event := twomqtt.Event{
				Type:    reflect.TypeOf(obj),
				Payload: obj,
			}
			c.mqttClient.ReceiveState(event)

			actualPayload := c.mqttClient.LastPublishedOnTopic(v.ExpectedTopic)
			if actualPayload != v.ExpectedPayload {
				t.Errorf("Actual:%s\nExpected:%s", actualPayload, v.ExpectedPayload)
			}
		}
	}
}

func TestSendCommand(t *testing.T) {
	defer clearEnvs()

	var tests = []struct {
		Known           string
		Serial          string
		TopicPrefix     string
		Command         twomqtt.Command
		CommandTopic    string
		CommandPayload  string
		ExpectedTopic   string
		ExpectedPayload string
	}{
		{
			knownRobot,
			knownSerial,
			knownPrefix,
			cmdPowerOff,
			knownPrefix + "/" + strings.ToLower(knownSerial) + "/power/command",
			"OFF",
			knownPrefix + "/" + strings.ToLower(knownSerial) + "/unitstatus/state",
			"Off",
		},
		{
			knownRobot,
			knownSerial,
			knownPrefix,
			cmdPanelLockOn,
			knownPrefix + "/" + strings.ToLower(knownSerial) + "/panellockactive/command",
			"ON",
			knownPrefix + "/" + strings.ToLower(knownSerial) + "/panellockactive/state",
			"ON",
		},
		{
			knownRobot,
			knownSerial,
			knownPrefix,
			cmdNightLightOn,
			knownPrefix + "/" + strings.ToLower(knownSerial) + "/nightlightactive/command",
			"ON",
			knownPrefix + "/" + strings.ToLower(knownSerial) + "/nightlightactive/state",
			"ON",
		},
	}

	for _, v := range tests {
		os.Setenv("MQTT_TOPICPREFIX", v.TopicPrefix)
		os.Setenv("LITTERROBOT_KNOWN", v.Known)
		serial := strings.ToLower(v.Serial)
		c := initialize()
		c.mqttClient.lastState[serial] = litterRobotState{
			LitterRobotSerial: v.Serial,
		}

		cmd := c.mqttClient.commandPower
		if v.Command == cmdPowerOff {
			cmd = c.mqttClient.commandPower
		} else if v.Command == cmdPanelLockOn {
			cmd = c.mqttClient.commandPanelLock
		} else if v.Command == cmdNightLightOn {
			cmd = c.mqttClient.commandNightLight
		}

		cmd(c.mqttClient.Client, &twomqtt.MoqMessage{
			TopicSrc:   v.CommandTopic,
			PayloadSrc: v.CommandPayload,
		})

		actualPayload := c.mqttClient.LastPublishedOnTopic(v.ExpectedTopic)
		if actualPayload != v.ExpectedPayload {
			t.Errorf("Actual:%s\nExpected:%s", actualPayload, v.ExpectedPayload)
		}
	}
}

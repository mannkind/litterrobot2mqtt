package main

import (
	"os"
	"strings"
	"testing"

	"github.com/mannkind/twomqtt"
	log "github.com/sirupsen/logrus"
)

const knownRobot = "A63afb501d65cb:Name"
const knownID = "A63afb501d65cb"
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
				"homeassistant/sensor/" + knownDiscoveryName + "/" + strings.ToLower(knownID) + "_" + knownType + "/config",
				"{\"availability_topic\":\"" + knownPrefix + "/status\",\"device\":{\"identifiers\":[\"" + knownPrefix + "/status\"],\"manufacturer\":\"twomqtt\",\"name\":\"x2mqtt\",\"sw_version\":\"X.X.X\"},\"name\":\"" + knownDiscoveryName + " " + strings.ToLower(knownID) + " " + knownType + "\",\"state_topic\":\"" + knownPrefix + "/" + strings.ToLower(knownID) + "/" + knownType + "/state\",\"unique_id\":\"" + knownDiscoveryName + "." + strings.ToLower(knownID) + "." + knownType + "\"}",
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
			LitterRobotID   string
			TopicPrefix     string
			ExpectedTopic   string
			ExpectedPayload string
		}{
			{
				knownRobot,
				knownID,
				knownPrefix,
				knownPrefix + "/" + strings.ToLower(knownID) + "/" + knownType + "/state",
				"Off",
			},
		}

		for _, v := range tests {
			setEnvs("false", "", v.TopicPrefix, v.Known)

			c := initialize()
			obj := litterRobotState{
				LitterRobotID: v.LitterRobotID,
				PowerStatus:   "Off", // Not a real status
				UnitStatus:    "OFF",
				CycleCount:    "Off", // Not a real status
			}

			c.mqttClient.receiveState(obj)

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
		Command         int64
		CommandTopic    string
		CommandPayload  string
		ExpectedTopic   string
		ExpectedPayload string
	}{
		{
			knownRobot,
			knownID,
			knownPrefix,
			cmdPowerOff,
			knownPrefix + "/" + strings.ToLower(knownID) + "/power/command",
			"OFF",
			knownPrefix + "/" + strings.ToLower(knownID) + "/unitstatus/state",
			"Off",
		},
		{
			knownRobot,
			knownID,
			knownPrefix,
			cmdPanelLockOn,
			knownPrefix + "/" + strings.ToLower(knownID) + "/panellockactive/command",
			"ON",
			knownPrefix + "/" + strings.ToLower(knownID) + "/panellockactive/state",
			"ON",
		},
		{
			knownRobot,
			knownID,
			knownPrefix,
			cmdNightLightOn,
			knownPrefix + "/" + strings.ToLower(knownID) + "/nightlightactive/command",
			"ON",
			knownPrefix + "/" + strings.ToLower(knownID) + "/nightlightactive/state",
			"ON",
		},
	}

	for _, v := range tests {
		os.Setenv("MQTT_TOPICPREFIX", v.TopicPrefix)
		os.Setenv("LITTERROBOT_KNOWN", v.Known)
		serial := strings.ToLower(v.Serial)
		c := initialize()
		c.mqttClient.lastState[serial] = litterRobotState{
			LitterRobotID: v.Serial,
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

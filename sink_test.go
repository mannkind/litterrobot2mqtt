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
			Known              string
			DiscoveryName      string
			TopicPrefix        string
			ExpectedName       string
			ExpectedStateTopic string
			ExpectedUniqueID   string
		}{
			{
				knownRobot,
				knownDiscoveryName,
				knownPrefix,
				knownDiscoveryName + " " + strings.ToLower(knownID) + " " + knownType,
				knownPrefix + "/" + strings.ToLower(knownID) + "/" + knownType + "/state",
				knownDiscoveryName + "." + strings.ToLower(knownID) + "." + knownType,
			},
		}

		for _, v := range tests {
			setEnvs("true", v.DiscoveryName, v.TopicPrefix, v.Known)

			c := initialize()
			mqds := c.sink.discovery()

			mqd := twomqtt.MQTTDiscovery{}
			for _, tmqd := range mqds {
				if tmqd.Name == v.ExpectedName {
					mqd = tmqd
					break
				}
			}

			if mqd.Name != v.ExpectedName {
				t.Errorf("discovery Name does not match; %s vs %s", mqd.Name, v.ExpectedName)
			}
			if mqd.StateTopic != v.ExpectedStateTopic {
				t.Errorf("discovery StateTopic does not match; %s vs %s", mqd.StateTopic, v.ExpectedStateTopic)
			}
			if mqd.UniqueID != v.ExpectedUniqueID {
				t.Errorf("discovery UniqueID does not match; %s vs %s", mqd.UniqueID, v.ExpectedUniqueID)
			}
		}
	}
}

func TestPublish(t *testing.T) {
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
			obj := sourceRep{
				LitterRobotID: v.LitterRobotID,
				PowerStatus:   "Off", // Not a real status
				UnitStatus:    "OFF",
				CycleCount:    "Off", // Not a real status
			}

			allPublished := c.sink.publish(obj)

			matching := twomqtt.MQTTMessage{}
			for _, state := range allPublished {
				if state.Topic == v.ExpectedTopic {
					matching = state
					break
				}
			}

			if matching.Payload != v.ExpectedPayload {
				t.Errorf("Actual:%s\nExpected:%s", matching.Payload, v.ExpectedPayload)
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
		c.sink.lastState[serial] = sourceRep{
			LitterRobotID: v.Serial,
		}

		cmd := c.sink.commandPower
		if v.Command == cmdPowerOff {
			cmd = c.sink.commandPower
		} else if v.Command == cmdPanelLockOn {
			cmd = c.sink.commandPanelLock
		} else if v.Command == cmdNightLightOn {
			cmd = c.sink.commandNightLight
		}

		allPublished := cmd(nil, &twomqtt.MoqMessage{
			TopicSrc:   v.CommandTopic,
			PayloadSrc: v.CommandPayload,
		})

		matching := twomqtt.MQTTMessage{}
		for _, state := range allPublished {
			if state.Topic == v.ExpectedTopic {
				matching = state
				break
			}
		}

		if matching.Payload != v.ExpectedPayload {
			t.Errorf("Actual:%s\nExpected:%s", matching.Payload, v.ExpectedPayload)
		}
	}
}

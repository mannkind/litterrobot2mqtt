package main

import (
	"time"

	"github.com/caarlos0/env"
	mqttExtCfg "github.com/mannkind/paho.mqtt.golang.ext/cfg"
	log "github.com/sirupsen/logrus"
)

type config struct {
	MQTT           *mqttExtCfg.MQTTConfig
	Local          bool          `env:"LITTERROBOT_LOCAL" envDefault:"false"`
	Email          string        `env:"LITTERROBOT_EMAIL" envDefault:""`
	Password       string        `env:"LITTERROBOT_PASSWORD" envDefault:""`
	APIKey         string        `env:"LITTERROBOT_APIKEY" envDefault:"Gmdfw5Cq3F3Mk6xvvO0inHATJeoDv6C3KfwfOuh0"`
	KnownRobots    []string      `env:"LITTERROBOT_KNOWN" envDefault:""`
	LookupInterval time.Duration `env:"LITTERROBOT_LOOKUPINTERVAL" envDefault:"37s"`
	DebugLogLevel  bool          `env:"LITTERROBOT_DEBUG" envDefault:"false"`
}

func newConfig(mqttCfg *mqttExtCfg.MQTTConfig) *config {
	c := config{}
	c.MQTT = mqttCfg

	if c.MQTT.ClientID == "" {
		c.MQTT.ClientID = "DefaultLitterRobot2MQTTClientID"
	}

	if c.MQTT.DiscoveryName == "" {
		c.MQTT.DiscoveryName = "litterrobot"
	}

	if c.MQTT.TopicPrefix == "" {
		c.MQTT.TopicPrefix = "home/litterrobot"
	}

	if err := env.Parse(&c); err != nil {
		log.Printf("Error unmarshaling configuration: %s", err)
	}

	redactedPassword := ""
	if len(c.MQTT.Password) > 0 {
		redactedPassword = "<REDACTED>"
	}

	redactedAPIPassword := ""
	if len(c.Password) > 0 {
		redactedAPIPassword = "<REDACTED>"
	}

	log.WithFields(log.Fields{
		"MQTT.ClientID":                 c.MQTT.ClientID,
		"MQTT.Broker":                   c.MQTT.Broker,
		"MQTT.Username":                 c.MQTT.Username,
		"MQTT.Password":                 redactedPassword,
		"MQTT.Discovery":                c.MQTT.Discovery,
		"MQTT.DiscoveryPrefix":          c.MQTT.DiscoveryPrefix,
		"MQTT.DiscoveryName":            c.MQTT.DiscoveryName,
		"MQTT.TopicPrefix":              c.MQTT.TopicPrefix,
		"LitterRobot.Local":             c.Local,
		"LitterRobot.KnownRobots":       c.KnownRobots,
		"LitterRobot.DebugLogLevel":     c.DebugLogLevel,
		"LitterRobotAPI.Email":          c.Email,
		"LitterRobotAPI.Password":       redactedAPIPassword,
		"LitterRobotAPI.LookupInterval": c.LookupInterval,
	}).Info("Environmental Settings")

	if c.DebugLogLevel {
		log.SetLevel(log.DebugLevel)
		log.Debug("Enabling the debug log level")
	}

	return &c
}

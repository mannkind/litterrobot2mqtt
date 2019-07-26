# litterrobot2mqtt

[![Software
License](https://img.shields.io/badge/License-MIT-orange.svg?style=flat-square)](https://github.com/mannkind/litterrobot2mqtt/blob/master/LICENSE.md)
[![Travis CI](https://img.shields.io/travis/mannkind/litterrobot2mqtt/master.svg?style=flat-square)](https://travis-ci.org/mannkind/litterrobot2mqtt)
[![Coverage Status](https://img.shields.io/codecov/c/github/mannkind/litterrobot2mqtt/master.svg)](http://codecov.io/github/mannkind/litterrobot2mqtt?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/mannkind/litterrobot2mqtt)](https://goreportcard.com/report/github.com/mannkind/litterrobot2mqtt)

## Installation

### Via Docker

```bash
docker run -d --name="litterrobot2mqtt" -v /etc/localtime:/etc/localtime:ro mannkind/litterrobot2mqtt
```

### Via Make

```bash
git clone https://github.com/mannkind/litterrobot2mqtt
cd litterrobot2mqtt
make
./litterrobot2mqtt
```

## Configuration

Configuration happens via environmental variables

```bash
LITTERROBOT_LOCAL       - [OPTIONAL] The flag that indicates that you want to use the local API (requires additonal setup)
LITTERROBOT_EMAIL       - [OPTIONAL] The username/email for the Litter Robot API (required, if not using the local API)
LITTERROBOT_PASSWORD    - [OPTIONAL] The password for the Litter Robot API (required, if not using the local API)
LITTERROBOT_APIKEY      - [OPTIONAL] The API Key for the Litter Robot API, defaults to "Gmdfw5Cq3F3Mk6xvvO0inHATJeoDv6C3KfwfOuh0" which is the API Key for the iOS app
LITTERROBOT_KNOWN       - [OPTIONAL] The mapping between a Litter Robot serial number and an ID (API) or IP Address (local), e.g. "LR3AAAAAAA:aaaaaaaaaaaaaa,LR3BABBBBB:bbbbbbbbbbbbbb"
MQTT_TOPICPREFIX        - [OPTIONAL] The MQTT topic on which to publish the receiver status, defaults to "litterrobot"
MQTT_DISCOVERY          - [OPTIONAL] The MQTT discovery flag for Home Assistant, defaults to false
MQTT_DISCOVERYPREFIX    - [OPTIONAL] The MQTT discovery prefix for Home Assistant, defaults to "homeassistant"
MQTT_DISCOVERYNAME      - [OPTIONAL] The MQTT discovery name for Home Assistant, defaults to "litterrobot"
MQTT_CLIENTID           - [OPTIONAL] The clientId, defaults to "DefaultLitterRobot2mqttClientID"
MQTT_BROKER             - [OPTIONAL] The MQTT broker, defaults to "tcp://mosquitto.org:1883"
MQTT_USERNAME           - [OPTIONAL] The MQTT username, default to ""
MQTT_PASSWORD           - [OPTIONAL] The MQTT password, default to ""
```

# litterrobot2mqtt

[![Software
License](https://img.shields.io/badge/License-MIT-orange.svg?style=flat-square)](https://github.com/mannkind/litterrobot2mqtt/blob/main/LICENSE.md)
[![Build Status](https://github.com/mannkind/litterrobot2mqtt/workflows/Main%20Workflow/badge.svg)](https://github.com/mannkind/litterrobot2mqtt/actions)
[![Coverage Status](https://img.shields.io/codecov/c/github/mannkind/litterrobot2mqtt/main.svg)](http://codecov.io/github/mannkind/litterrobot2mqtt?branch=main)

An experiment to publish Litter Robot statuses to MQTT.

## Use

The application can be locally built using `dotnet build` or you can utilize the multi-architecture Docker image(s).

### Example

```bash
docker run \
-e LITTERROBOT__LOGIN="youremailaddress@example.com" \
-e LITTERROBOT__PASSWORD="yourpassword" \
-e LITTERROBOT__RESOURCES__0__LRID="69103754" \
-e LITTERROBOT__RESOURCES__0__Slug="main" \
-e LITTERROBOT__MQTT__BROKER="localhost" \
-e LITTERROBOT__MQTT__DISCOVERYENABLED="true" \
mannkind/litterrobot2mqtt:latest
```

OR

```bash
LITTERROBOT__LOGIN="youremailaddress@example.com" \
LITTERROBOT__PASSWORD="yourpassword" \
LITTERROBOT__RESOURCES__0__LRID="69103754" \
LITTERROBOT__RESOURCES__0__Slug="main" \
LITTERROBOT__MQTT__BROKER="localhost" \
LITTERROBOT__MQTT__DISCOVERYENABLED="true" \
./litterrobot2mqtt 
```


## Configuration

Configuration happens via environmental variables

```bash
LITTERROBOT__LOGIN                              - The Litter Robot Login
LITTERROBOT__PASSWORD                           - The Litter Robot Password
LITTERROBOT__RESOURCES__#__LRID                 - The Litter Robot ID for a specific Litter Robot
LITTERROBOT__RESOURCES__#__Slug                 - The slug to identify the specific Litter Robot ID
LITTERROBOT__POLLINGINTERVAL                    - [OPTIONAL] The delay between litter robot status lookups, defaults to "0.00:00:31"
LITTERROBOT__MQTT__TOPICPREFIX                  - [OPTIONAL] The MQTT topic on which to publish the collection lookup results, defaults to "home/litterrobot"
LITTERROBOT__MQTT__DISCOVERYENABLED             - [OPTIONAL] The MQTT discovery flag for Home Assistant, defaults to false
LITTERROBOT__MQTT__DISCOVERYPREFIX              - [OPTIONAL] The MQTT discovery prefix for Home Assistant, defaults to "homeassistant"
LITTERROBOT__MQTT__DISCOVERYNAME                - [OPTIONAL] The MQTT discovery name for Home Assistant, defaults to "litterrobot"
LITTERROBOT__MQTT__BROKER                       - [OPTIONAL] The MQTT broker, defaults to "test.mosquitto.org"
LITTERROBOT__MQTT__USERNAME                     - [OPTIONAL] The MQTT username, default to ""
LITTERROBOT__MQTT__PASSWORD                     - [OPTIONAL] The MQTT password, default to ""
```

## Prior Implementations

### Golang
* Last Commit: [5b623935f3a73db0ed986997f1b22375f5465867](https://github.com/mannkind/litterrobot2mqtt/commit/5b623935f3a73db0ed986997f1b22375f5465867)
* Last Docker Image: [mannkind/litterrobot2mqtt:v0.2.19360.0428](https://hub.docker.com/layers/mannkind/litterrobot2mqtt/v0.2.19360.0428/images/sha256-b65876a4036ae9fa37260dd173350477174c82188165747076af08a533fb26e6?context=explore)
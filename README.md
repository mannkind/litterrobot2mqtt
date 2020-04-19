# litterrobot2mqtt

[![Software
License](https://img.shields.io/badge/License-MIT-orange.svg?style=flat-square)](https://github.com/mannkind/litterrobot2mqtt/blob/master/LICENSE.md)
[![Build Status](https://github.com/mannkind/litterrobot2mqtt/workflows/Main%20Workflow/badge.svg)](https://github.com/mannkind/litterrobot2mqtt/actions)
[![Coverage Status](https://img.shields.io/codecov/c/github/mannkind/litterrobot2mqtt/master.svg)](http://codecov.io/github/mannkind/litterrobot2mqtt?branch=master)

An experiment to publish Litter Robot statuses to MQTT.

## Use

The application can be locally built using `dotnet build` or you can utilize the multi-architecture Docker image(s).

### Example

```bash
docker run \
-e LITTERROBOT__SOURCE__LOGIN="youremailaddress@example.com" \
-e LITTERROBOT__SOURCE__PASSWORD="yourpassword" \
-e LITTERROBOT__SHARED__RESOURCES__0__LRID="69103754" \
-e LITTERROBOT__SHARED__RESOURCES__0__Slug="main" \
-e LITTERROBOT__SINK__BROKER="localhost" \
-e LITTERROBOT__SINK__DISCOVERYENABLED="true" \
mannkind/litterrobot2mqtt:latest
```

OR

```bash
LITTERROBOT__SOURCE__LOGIN="youremailaddress@example.com" \
LITTERROBOT__SOURCE__PASSWORD="yourpassword" \
LITTERROBOT__SHARED__RESOURCES__0__LRID="69103754" \
LITTERROBOT__SHARED__RESOURCES__0__Slug="main" \
LITTERROBOT__SINK__BROKER="localhost" \
LITTERROBOT__SINK__DISCOVERYENABLED="true" \
./litterrobot2mqtt 
```


## Configuration

Configuration happens via environmental variables

```bash
LITTERROBOT__SOURCE__LOGIN                      - The Litter Robot Login
LITTERROBOT__SOURCE__PASSWORD                   - The Litter Robot Password
LITTERROBOT__SHARED__RESOURCES__#__ID           - The Litter Robot ID for a specific Litter Robot
LITTERROBOT__SHARED__RESOURCES__#__Slug         - The slug to identify the specific Litter Robot ID
LITTERROBOT__SOURCE__POLLINGINTERVAL            - [OPTIONAL] The delay between litter robot status lookups, defaults to "0.00:00:31"
LITTERROBOT__SINK__TOPICPREFIX                  - [OPTIONAL] The MQTT topic on which to publish the collection lookup results, defaults to "home/litterrobot"
LITTERROBOT__SINK__DISCOVERYENABLED             - [OPTIONAL] The MQTT discovery flag for Home Assistant, defaults to false
LITTERROBOT__SINK__DISCOVERYPREFIX              - [OPTIONAL] The MQTT discovery prefix for Home Assistant, defaults to "homeassistant"
LITTERROBOT__SINK__DISCOVERYNAME                - [OPTIONAL] The MQTT discovery name for Home Assistant, defaults to "litterrobot"
LITTERROBOT__SINK__BROKER                       - [OPTIONAL] The MQTT broker, defaults to "test.mosquitto.org"
LITTERROBOT__SINK__USERNAME                     - [OPTIONAL] The MQTT username, default to ""
LITTERROBOT__SINK__PASSWORD                     - [OPTIONAL] The MQTT password, default to ""
```

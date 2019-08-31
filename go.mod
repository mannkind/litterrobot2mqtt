module github.com/mannkind/litterrobot2mqtt

require (
	github.com/caarlos0/env/v6 v6.0.0
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/google/wire v0.3.0
	github.com/mannkind/litterrobot v0.2.0
	github.com/mannkind/twomqtt v0.1.0
	github.com/sirupsen/logrus v1.4.2
	golang.org/x/tools/gopls v0.1.3 // indirect
)

// local development
// replace github.com/mannkind/litterrobot => ../litterrobot
// replace github.com/mannkind/twomqtt => ../twomqtt

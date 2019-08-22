module github.com/mannkind/litterrobot2mqtt

require (
	github.com/caarlos0/env v3.5.0+incompatible
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/google/wire v0.3.0
	github.com/mannkind/litterrobot v0.1.2
	github.com/mannkind/paho.mqtt.golang.ext v0.3.0
	github.com/sirupsen/logrus v1.4.2
	golang.org/x/tools/gopls v0.1.3 // indirect
)

// local development
// replace github.com/mannkind/litterrobot => ../litterrobot
// replace github.com/mannkind/paho.mqtt.golang.ext => ../paho.mqtt.golang.ext

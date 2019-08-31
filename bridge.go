package main

type bridge struct {
	mqttClient    *mqttClient
	serviceClient *serviceClient
}

func newBridge(mqttClient *mqttClient, serviceClient *serviceClient) *bridge {
	bridge := bridge{
		mqttClient:    mqttClient,
		serviceClient: serviceClient,
	}

	return &bridge
}

func (b *bridge) run() {
	b.mqttClient.Register(b.serviceClient)
	b.serviceClient.Register(b.mqttClient)

	b.mqttClient.run()
	b.serviceClient.run()
}
